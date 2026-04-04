package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapters "github.com/bshongwe/linkpulse/backend/services/web-bff/internal/adapters/http"
	"github.com/bshongwe/linkpulse/backend/services/web-bff/internal/application"
	httphandlers "github.com/bshongwe/linkpulse/backend/services/web-bff/internal/presentation/http"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Get environment and port
	environment := os.Getenv("LINKPULSE_ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	port := 8080
	if p := os.Getenv("LINKPULSE_SERVER_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	// Get JWT secret from environment
	jwtSecret := os.Getenv("LINKPULSE_JWT_ACCESS_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-access-key-change-in-production-2026"
	}

	// Initialize logger
	var log *zap.Logger
	var err error
	if environment == "production" {
		log, err = zap.NewProduction()
	} else {
		log, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	// Set Gin mode
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize HTTP clients (adapters)
	shortenerClient := httpadapters.NewShortenerHTTPClient(
		"http://shortener-service:8082",
		log,
	)
	analyticsClient := httpadapters.NewAnalyticsHTTPClient(
		"http://analytics-service:8082",
		log,
	)

	// Create BFF service
	bffService := application.NewBFFService(shortenerClient, analyticsClient, log)

	// Create HTTP handler
	handler := httphandlers.NewHandler(bffService, log)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configure CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Register BFF routes with JWT secret
	handler.RegisterRoutes(router, jwtSecret)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Web BFF service started", zap.Int("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown server", zap.Error(err))
	}

	log.Info("Web BFF service stopped")
}
