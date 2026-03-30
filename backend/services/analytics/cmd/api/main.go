package main

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/adapters/kafka"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/adapters/timescaledb"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/adapters/websocket"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/application"
	httphandler "github.com/bshongwe/linkpulse/backend/services/analytics/internal/presentation/http"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
)

func main() {
	// Initialize logger
	log := logger.NewLogger("analytics")
	defer log.Sync()

	// Database configuration from environment
	dbURL := fmt.Sprintf(
		"postgres://linkpulse:password@localhost:5432/linkpulse?sslmode=disable",
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping database", zap.Error(err))
	}

	log.Info("connected to database")

	// Initialize adapters
	clickRepo := timescaledb.NewClickRepository(pool)
	log.Info("initialized TimescaleDB click repository")

	kafkaBrokers := []string{"localhost:9092"}
	eventPublisher, err := kafka.NewEventPublisher(kafkaBrokers, "click-events", log)
	if err != nil {
		log.Fatal("failed to create Kafka publisher", zap.Error(err))
	}
	defer eventPublisher.Close()

	eventConsumer, err := kafka.NewEventConsumer(kafkaBrokers, "click-events", "analytics-service", log)
	if err != nil {
		log.Fatal("failed to create Kafka consumer", zap.Error(err))
	}
	defer eventConsumer.Close()

	clickNotifier := websocket.NewClickNotifier(log)

	// Initialize application service
	analyticsService := application.NewAnalyticsService(
		clickRepo,
		clickNotifier,
		eventPublisher,
		nil, // LocationService - optional for now
	)

	// Register Kafka consumer handlers
	eventConsumer.RegisterHandler(func(ctx context.Context, event *application.ClickEvent) error {
		return analyticsService.RecordClick(ctx, event)
	})

	// Start Kafka consumer
	go func() {
		if err := eventConsumer.Start(context.Background()); err != nil {
			log.Error("consumer failed", zap.Error(err))
		}
	}()
	log.Info("started Kafka consumer")

	// Initialize HTTP router
	router := gin.Default()

	// Create HTTP handler
	handler := httphandler.NewHandler(analyticsService, log)
	handler.RegisterRoutes(router)

	log.Info("registered HTTP routes")

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(stdhttp.StatusOK, gin.H{"status": "healthy"})
	})

	// Start HTTP server
	server := &stdhttp.Server{
		Addr:         ":8082", // Analytics service port
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		log.Info("starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Info("received signal", zap.Any("signal", sig))

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}

	log.Info("server stopped gracefully")
}
