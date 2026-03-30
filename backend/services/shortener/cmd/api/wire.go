//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/adapters/postgres"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/adapters/redis"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/application"
	httphandler "github.com/bshongwe/linkpulse/backend/services/shortener/internal/presentation/http"
	"github.com/bshongwe/linkpulse/backend/shared/config"
)

// provideDB wraps postgres.NewDB to include cleanup function for Wire composition
func provideDB(cfg *config.DatabaseConfig) (*sqlx.DB, func(), error) {
	db, err := postgres.NewDB(cfg)
	if err != nil {
		return nil, nil, err
	}
	return db, func() { _ = db.Close() }, nil
}

// provideRedisClient wraps redis.NewClient to include cleanup function for Wire composition
func provideRedisClient(cfg *config.RedisConfig) (*redis.Client, func(), error) {
	client, err := redis.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, func() { _ = client.Close() }, nil
}

var Set = wire.NewSet(
	provideDB,
	postgres.NewLinkRepository,
	provideRedisClient,
	redis.NewCache,
	application.NewShortenerService,
	httphandler.NewShortenerHandler,
	wire.FieldsOf(new(*config.Config), "Database"),
	wire.FieldsOf(new(*config.Config), "Redis"),
)

func Initialize() (*httphandler.ShortenerHandler, func(), error) {
	wire.Build(
		config.Load,
		Set,
	)
	return nil, nil, nil // wire will replace this
}
