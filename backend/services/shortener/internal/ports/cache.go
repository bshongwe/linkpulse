package ports

import (
	"context"
	"time"
)

// CachePort defines the interface for caching operations
type CachePort interface {
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// IncrementInt increments an integer value in cache
	IncrementInt(ctx context.Context, key string, delta int64) (int64, error)

	// SetWithoutTTL stores a value in cache without expiration
	SetWithoutTTL(ctx context.Context, key string, value interface{}) error
}
