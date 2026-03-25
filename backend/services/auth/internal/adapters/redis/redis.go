package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/bshongwe/linkpulse/backend/shared/config"
)

func NewClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

// tokenBlacklist implements ports.TokenBlacklist using Redis.
type tokenBlacklist struct {
	client *redis.Client
}

func NewTokenBlacklist(client *redis.Client) *tokenBlacklist {
	return &tokenBlacklist{client: client}
}

func (b *tokenBlacklist) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	return b.client.Set(ctx, "blacklist:"+jti, 1, ttl).Err()
}

func (b *tokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := b.client.Exists(ctx, "blacklist:"+jti).Result()
	return n > 0, err
}
