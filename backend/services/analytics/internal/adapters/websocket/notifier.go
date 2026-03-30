package websocket

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/domain"
	"github.com/bshongwe/linkpulse/backend/services/analytics/internal/ports"
)

// Subscriber represents a single WebSocket subscriber
type Subscriber struct {
	id      string
	linkID  uuid.UUID
	handler func(*domain.ClickEvent)
	done    chan struct{}
}

// ClickNotifier implements the ports.ClickNotifier interface for real-time notifications
type ClickNotifier struct {
	subscribers map[uuid.UUID][]*Subscriber // linkID -> subscribers
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewClickNotifier creates a new WebSocket click notifier
func NewClickNotifier(logger *zap.Logger) ports.ClickNotifier {
	return &ClickNotifier{
		subscribers: make(map[uuid.UUID][]*Subscriber),
		logger:      logger,
	}
}

// NotifyClick broadcasts a click event to all subscribers of a link
func (cn *ClickNotifier) NotifyClick(ctx context.Context, linkID uuid.UUID, event *domain.ClickEvent) error {
	cn.mu.RLock()
	subscribers, exists := cn.subscribers[linkID]
	cn.mu.RUnlock()

	if !exists || len(subscribers) == 0 {
		return nil // No subscribers for this link
	}

	// Broadcast to all subscribers in parallel
	for _, sub := range subscribers {
		go func(s *Subscriber) {
			select {
			case <-s.done:
				// Subscriber has been unsubscribed
				return
			case <-ctx.Done():
				// Context cancelled
				return
			default:
				s.handler(event)
			}
		}(sub)
	}

	return nil
}

// Subscribe registers a listener for click events on a specific link
func (cn *ClickNotifier) Subscribe(linkID uuid.UUID, handler func(*domain.ClickEvent)) (func(), error) {
	if handler == nil {
		return nil, fmt.Errorf(errWrap, errInvalidHandler, fmt.Errorf("handler cannot be nil"))
	}

	subscriber := &Subscriber{
		id:      uuid.New().String(),
		linkID:  linkID,
		handler: handler,
		done:    make(chan struct{}),
	}

	cn.mu.Lock()
	cn.subscribers[linkID] = append(cn.subscribers[linkID], subscriber)
	cn.mu.Unlock()

	cn.logger.Debug("subscriber registered",
		zap.String("subscriberID", subscriber.id),
		zap.String("linkID", linkID.String()),
	)

	// Return unsubscribe function
	unsubscribe := func() {
		cn.Unsubscribe(linkID, subscriber.id)
	}

	return unsubscribe, nil
}

// Unsubscribe removes a subscriber from a link
func (cn *ClickNotifier) Unsubscribe(linkID uuid.UUID, subscriberID string) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	subscribers, exists := cn.subscribers[linkID]
	if !exists {
		return
	}

	// Find and remove the subscriber
	for i, sub := range subscribers {
		if sub.id == subscriberID {
			close(sub.done)
			cn.subscribers[linkID] = append(subscribers[:i], subscribers[i+1:]...)
			
			cn.logger.Debug("subscriber unregistered",
				zap.String("subscriberID", subscriberID),
				zap.String("linkID", linkID.String()),
			)

			// Clean up empty subscriber lists
			if len(cn.subscribers[linkID]) == 0 {
				delete(cn.subscribers, linkID)
			}
			return
		}
	}
}

// UnsubscribeAll removes all subscribers for a link
func (cn *ClickNotifier) UnsubscribeAll(linkID uuid.UUID) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	subscribers, exists := cn.subscribers[linkID]
	if !exists {
		return
	}

	for _, sub := range subscribers {
		close(sub.done)
	}

	delete(cn.subscribers, linkID)
	cn.logger.Debug("all subscribers unregistered", zap.String("linkID", linkID.String()))
}

// GetSubscriberCount returns the number of active subscribers for a link
func (cn *ClickNotifier) GetSubscriberCount(linkID uuid.UUID) int {
	cn.mu.RLock()
	defer cn.mu.RUnlock()

	subscribers, exists := cn.subscribers[linkID]
	if !exists {
		return 0
	}

	return len(subscribers)
}
