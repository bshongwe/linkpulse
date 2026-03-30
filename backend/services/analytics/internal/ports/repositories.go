package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
)

// ClickRepository defines the interface for click data persistence
type ClickRepository interface {
	// RecordClick persists a single click event
	RecordClick(ctx context.Context, event *domain.ClickEvent) error

	// GetSummary retrieves analytics summary for a link within a time range
	GetSummary(ctx context.Context, linkID uuid.UUID, since time.Time) (*domain.AnalyticsSummary, error)

	// GetLiveCount retrieves the current click count for a short code
	GetLiveCount(ctx context.Context, shortCode string) (int64, error)

	// GetClicksByTimeRange retrieves clicks for a link within a specific time range
	GetClicksByTimeRange(ctx context.Context, linkID uuid.UUID, start, end time.Time) ([]*domain.ClickEvent, error)

	// GetCountryDistribution retrieves click distribution by country
	GetCountryDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error)

	// GetDeviceDistribution retrieves click distribution by device type
	GetDeviceDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error)
}

// EventPublisher defines the interface for publishing click events to message queue
type EventPublisher interface {
	// PublishClickEvent publishes a click event to the message broker
	PublishClickEvent(ctx context.Context, event *domain.ClickEvent) error

	// Close closes the publisher connection
	Close() error
}

// EventConsumer defines the interface for consuming click events from message queue
type EventConsumer interface {
	// RegisterHandler registers a handler to process incoming click events
	RegisterHandler(handler EventHandler)

	// Start begins consuming events from the message broker
	Start(ctx context.Context) error

	// Close closes the consumer connection
	Close() error
}

// EventHandler is a function that handles incoming click events
type EventHandler func(ctx context.Context, event *domain.ClickEvent) error

// LocationService defines the interface for IP geolocation lookups
type LocationService interface {
	// GetCountryCode returns the country code for an IP address
	GetCountryCode(ctx context.Context, ipAddress string) (string, error)

	// Close closes the location service connection
	Close() error
}

// ClickNotifier defines the interface for real-time click notifications (WebSocket)
type ClickNotifier interface {
	// NotifyClick broadcasts a click event to subscribers
	NotifyClick(ctx context.Context, linkID uuid.UUID, event *domain.ClickEvent) error

	// Subscribe registers a listener for click events on a specific link
	Subscribe(linkID uuid.UUID, handler func(*domain.ClickEvent)) (func(), error)
}

// AnalyticsService defines the interface for analytics business logic
type AnalyticsService interface {
	// RecordClick processes and records a click event
	RecordClick(ctx context.Context, event *domain.ClickEvent) error

	// GetAnalytics retrieves aggregated analytics for a link
	GetAnalytics(ctx context.Context, linkID uuid.UUID, since time.Time) (*domain.AnalyticsSummary, error)

	// GetLiveCount retrieves the current click count for a short code
	GetLiveCount(ctx context.Context, shortCode string) (int64, error)

	// GetCountryDistribution retrieves click distribution by country
	GetCountryDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error)

	// GetDeviceDistribution retrieves click distribution by device type
	GetDeviceDistribution(ctx context.Context, linkID uuid.UUID) (map[string]int64, error)

	// GetClicksByTimeRange retrieves click events within a time range
	GetClicksByTimeRange(ctx context.Context, linkID uuid.UUID, start, end time.Time) ([]*domain.ClickEvent, error)
}
