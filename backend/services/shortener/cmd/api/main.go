package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/shared/config"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
	"github.com/bshongwe/linkpulse/backend/shared/middleware"
	"github.com/bshongwe/linkpulse/backend/shared/otel"

	httphandler "github.com/bshongwe/linkpulse/backend/services/shortener/internal/presentation/http"
)

func buildCORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{})
	if allowedOrigins != "" {
		for _, o := range strings.Split(allowedOrigins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowedSet[o] = struct{}{}
			}
		}
	}

	if len(allowedSet) == 0 {
		logger.Log.Warn("CORS: No allowed origins configured. Set LINKPULSE_ALLOWED_ORIGINS to restrict access.")
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowedSet[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Init(cfg.OTel.Environment)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownOtel, err := otel.Init(ctx, &cfg.OTel)
	if err != nil {
		logger.Log.Error("Failed to init OpenTelemetry", zap.Error(err))
		// Provide a no-op cleanup function when OTel initialization fails
		// to prevent panic in the deferred cleanup call
		shutdownOtel = func() { /* nothing to cleanup */ }
	}
	defer shutdownOtel()

	handler, cleanup, err := Initialize()
	if err != nil {
		logger.Log.Fatal("Failed to initialize dependencies", zap.Error(err))
	}
	defer cleanup()

	r := gin.Default()
	r.Use(buildCORSMiddleware(os.Getenv("LINKPULSE_ALLOWED_ORIGINS")))

	// Public — redirect endpoint (no auth needed, it's the hot path)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "shortener"})
	})

	// Protected — all link management routes require a valid JWT
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(middleware.NewJWTValidator(cfg.JWT.AccessSecret)))
	httphandler.RegisterRoutes(protected, handler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	logger.Log.Info("Shortener service started", zap.Int("port", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdown)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}
}
