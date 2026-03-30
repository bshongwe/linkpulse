package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/ports"
)

// EventPublisher implements the ports.EventPublisher interface using Kafka
type EventPublisher struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// NewEventPublisher creates a new Kafka event publisher
func NewEventPublisher(brokers []string, topic string, logger *zap.Logger) (*EventPublisher, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &EventPublisher{
		writer: writer,
		logger: logger,
	}, nil
}

// PublishClickEvent publishes a click event to Kafka
func (ep *EventPublisher) PublishClickEvent(ctx context.Context, event *domain.ClickEvent) error {
	if !event.IsValid() {
		return fmt.Errorf(errWrap, errInvalidEvent, fmt.Errorf("click event validation failed"))
	}

	// Marshal the event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf(errWrap, errMarshalFailed, err)
	}

	// Publish to Kafka with link ID as key for partitioning
	msg := kafka.Message{
		Key:   []byte(event.LinkID.String()),
		Value: payload,
	}

	if err := ep.writer.WriteMessages(ctx, msg); err != nil {
		ep.logger.Error("failed to publish click event",
			zap.Error(err),
			zap.String("linkID", event.LinkID.String()),
		)
		return fmt.Errorf(errWrap, errPublishFailed, err)
	}

	ep.logger.Debug("click event published",
		zap.String("linkID", event.LinkID.String()),
		zap.String("shortCode", event.ShortCode),
	)

	return nil
}

// Close gracefully closes the publisher
func (ep *EventPublisher) Close() error {
	if ep.writer != nil {
		return ep.writer.Close()
	}
	return nil
}
