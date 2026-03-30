package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/ports"
	"github.com/bshongwe/linkpulse/backend/shared/config"
	sharedErrors "github.com/bshongwe/linkpulse/backend/shared/errors"
)

func NewClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	// Ping Redis with timeout to verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return client, nil
}

// cache implements ports.CachePort using Redis.
type cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) ports.CachePort {
	return &cache{client: client}
}

func (c *cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *cache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		// Normalize redis.Nil (key not found) to standardized ErrCacheMiss sentinel
		if err == redis.Nil {
			return nil, sharedErrors.ErrCacheMiss
		}
		return nil, err
	}
	return val, nil
}

func (c *cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *cache) IncrementInt(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, key, delta).Result()
}

func (c *cache) SetWithoutTTL(ctx context.Context, key string, value interface{}) error {
	return c.client.Set(ctx, key, value, 0).Err()
}
