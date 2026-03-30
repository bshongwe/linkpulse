//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/adapters/postgres"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/adapters/redis"
	"github.com/bshongwe/linkpulse/backend/services/shortener/internal/application"
	httphandler "github.com/bshongwe/linkpulse/backend/services/shortener/internal/presentation/http"
	"github.com/bshongwe/linkpulse/backend/shared/config"
)

var Set = wire.NewSet(
	postgres.NewDB,
	postgres.NewLinkRepository,
	redis.NewClient,
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
