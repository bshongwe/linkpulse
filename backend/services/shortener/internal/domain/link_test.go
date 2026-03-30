package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/domain"
)

func baseLink() *domain.ShortLink {
	return &domain.ShortLink{
		ID:          uuid.New(),
		ShortCode:   "abc12345",
		OriginalURL: "https://example.com",
		WorkspaceID: uuid.New(),
		CreatedBy:   uuid.New(),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestShortLink_IsExpired(t *testing.T) {
	t.Run("no expiry set returns false", func(t *testing.T) {
		l := baseLink()
		if l.IsExpired() {
			t.Error("expected IsExpired() = false when ExpiresAt is nil")
		}
	})

	t.Run("future expiry returns false", func(t *testing.T) {
		l := baseLink()
		future := time.Now().Add(24 * time.Hour)
		l.ExpiresAt = &future
		if l.IsExpired() {
			t.Error("expected IsExpired() = false for future expiry")
		}
	})

	t.Run("past expiry returns true", func(t *testing.T) {
		l := baseLink()
		past := time.Now().Add(-1 * time.Second)
		l.ExpiresAt = &past
		if !l.IsExpired() {
			t.Error("expected IsExpired() = true for past expiry")
		}
	})
}

func TestShortLink_CanAccess(t *testing.T) {
	t.Run("active and not expired returns true", func(t *testing.T) {
		l := baseLink()
		if !l.CanAccess() {
			t.Error("expected CanAccess() = true for active, non-expired link")
		}
	})

	t.Run("inactive returns false", func(t *testing.T) {
		l := baseLink()
		l.IsActive = false
		if l.CanAccess() {
			t.Error("expected CanAccess() = false for inactive link")
		}
	})

	t.Run("expired returns false", func(t *testing.T) {
		l := baseLink()
		past := time.Now().Add(-1 * time.Second)
		l.ExpiresAt = &past
		if l.CanAccess() {
			t.Error("expected CanAccess() = false for expired link")
		}
	})

	t.Run("inactive and expired returns false", func(t *testing.T) {
		l := baseLink()
		l.IsActive = false
		past := time.Now().Add(-1 * time.Second)
		l.ExpiresAt = &past
		if l.CanAccess() {
			t.Error("expected CanAccess() = false for inactive and expired link")
		}
	})
}
