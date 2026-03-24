package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/bshongwe/linkpulse/backend/services/auth"
	httphandler "github.com/bshongwe/linkpulse/backend/services/auth/internal/presentation/http"
	"github.com/bshongwe/linkpulse/backend/shared/config"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
	"github.com/bshongwe/linkpulse/backend/shared/otel"
)

func main() {
	// fmt.Println("LinkPulse Auth Service starting...")
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.OTel.Environment)

	// Initialize OpenTelemetry
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownOtel, err := otel.Init(ctx, &cfg.OTel)
	if err != nil {
		logger.Log.Error("Failed to init OpenTelemetry", zap.Error(err))
	}
	defer shutdownOtel()

	// Wire DB + Auth Service
	handler, cleanup, err := auth.Initialize(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize dependencies", zap.Error(err))
	}
	defer cleanup()

	r := gin.Default()

	// Routes
	r.GET("/health", handler.Health)
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	srv := &http.Server{
		Addr:    ":" + string(rune(cfg.Server.Port)),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	logger.Log.Info("Auth service started", zap.Int("port", cfg.Server.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdown)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}
}
