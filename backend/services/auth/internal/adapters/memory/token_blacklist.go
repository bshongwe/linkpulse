package memory

import (
	"context"
	"sync"
	"time"

	"github.com/bshongwe/linkpulse/backend/services/auth/internal/ports"
)

// InMemoryTokenBlacklist is a simple in-memory implementation of TokenBlacklist.
// In production, use Redis or a similar distributed cache.
type InMemoryTokenBlacklist struct {
	mu        sync.RWMutex
	revoked   map[string]time.Time // jti -> revocation time
	cleanupTk *time.Ticker
	done      chan struct{}
}

// NewInMemoryTokenBlacklist creates a new in-memory token blacklist.
func NewInMemoryTokenBlacklist() ports.TokenBlacklist {
	tb := &InMemoryTokenBlacklist{
		revoked:   make(map[string]time.Time),
		cleanupTk: time.NewTicker(1 * time.Minute),
		done:      make(chan struct{}),
	}

	// Background cleanup of expired entries
	go func() {
		for {
			select {
			case <-tb.cleanupTk.C:
				tb.cleanup()
			case <-tb.done:
				tb.cleanupTk.Stop()
				return
			}
		}
	}()

	return tb
}

// Revoke adds a token JTI to the blacklist with a TTL.
func (tb *InMemoryTokenBlacklist) Revoke(_ context.Context, jti string, ttl time.Duration) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.revoked[jti] = time.Now().Add(ttl)
	return nil
}

// IsRevoked checks if a token JTI is in the blacklist.
func (tb *InMemoryTokenBlacklist) IsRevoked(_ context.Context, jti string) (bool, error) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	expiry, exists := tb.revoked[jti]
	if !exists {
		return false, nil
	}
	// Check if the entry has expired
	if time.Now().After(expiry) {
		return false, nil
	}
	return true, nil
}

// cleanup removes expired entries from the blacklist.
func (tb *InMemoryTokenBlacklist) cleanup() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	now := time.Now()
	for jti, expiry := range tb.revoked {
		if now.After(expiry) {
			delete(tb.revoked, jti)
		}
	}
}

// Close stops the background cleanup routine.
func (tb *InMemoryTokenBlacklist) Close() {
	close(tb.done)
}
