package ports

import (
	"context"
	"time"
)

// TokenBlacklist stores revoked refresh token JTIs until their natural expiry.
type TokenBlacklist interface {
	Revoke(ctx context.Context, jti string, ttl time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}
