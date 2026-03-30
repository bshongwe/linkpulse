package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/ports"
)

// EventConsumer implements the ports.EventConsumer interface using Kafka
type EventConsumer struct {
	reader   *kafka.Reader
	handlers []ports.EventHandler
	logger   *zap.Logger
	mu       sync.RWMutex
	done     chan struct{}
}

// NewEventConsumer creates a new Kafka event consumer
func NewEventConsumer(brokers []string, topic string, groupID string, logger *zap.Logger) (*EventConsumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		Topic:           topic,
		GroupID:         groupID,
		CommitInterval:  commitInterval,
		StartOffset:     kafka.LastOffset,
		MaxBytes:        10e6, // 10MB
		SessionTimeout:  30000, // 30 seconds
		RebalanceTimeout: 90000, // 90 seconds
	})

	return &EventConsumer{
		reader:   reader,
		handlers: []ports.EventHandler{},
		logger:   logger,
		done:     make(chan struct{}),
	}, nil
}

// Start begins consuming click events from Kafka
func (ec *EventConsumer) Start(ctx context.Context) error {
	if len(ec.handlers) == 0 {
		return fmt.Errorf(errWrap, errNoHandlers, fmt.Errorf("at least one handler must be registered"))
	}

	ec.logger.Info("starting Kafka event consumer", zap.String("topic", ec.reader.Config().Topic))

	go ec.consume(ctx)
	return nil
}

// Close gracefully shuts down the consumer
func (ec *EventConsumer) Close() error {
	close(ec.done)
	return ec.reader.Close()
}

// RegisterHandler adds an event handler to be called when events are received
func (ec *EventConsumer) RegisterHandler(handler ports.EventHandler) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.handlers = append(ec.handlers, handler)
}

// consume reads messages from Kafka and invokes handlers
func (ec *EventConsumer) consume(ctx context.Context) {
	for {
		select {
		case <-ec.done:
			ec.logger.Info("consumer shutdown requested")
			return
		case <-ctx.Done():
			ec.logger.Info("consumer context cancelled")
			return
		default:
		}

		// Read message from Kafka with timeout
		readCtx, cancel := context.WithTimeout(ctx, readTimeout)
		msg, err := ec.reader.ReadMessage(readCtx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			ec.logger.Error("failed to read message", zap.Error(err))
			continue
		}

		// Unmarshal the click event
		var event domain.ClickEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			ec.logger.Error("failed to unmarshal click event",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int("partition", msg.Partition),
			)
			continue
		}

		// Validate the event
		if !event.IsValid() {
			ec.logger.Warn("received invalid click event",
				zap.String("linkID", event.LinkID.String()),
				zap.String("shortCode", event.ShortCode),
			)
			continue
		}

		// Invoke all registered handlers
		ec.handleEvent(ctx, &event)
	}
}

// handleEvent processes a click event with all registered handlers
func (ec *EventConsumer) handleEvent(ctx context.Context, event *domain.ClickEvent) {
	ec.mu.RLock()
	handlers := ec.handlers
	ec.mu.RUnlock()

	for _, handler := range handlers {
		// Run handler in a goroutine to prevent blocking the consumer
		go func(h ports.EventHandler) {
			if err := h(ctx, event); err != nil {
				ec.logger.Error("event handler failed",
					zap.Error(err),
					zap.String("linkID", event.LinkID.String()),
				)
			}
		}(handler)
	}
}
