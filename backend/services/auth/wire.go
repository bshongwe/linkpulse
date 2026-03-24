//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/adapters/postgres"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/ports"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/presentation/http"
	"github.com/bshongwe/linkpulse/backend/shared/config"
	"github.com/bshongwe/linkpulse/backend/shared/logger"
	"github.com/bshongwe/linkpulse/backend/shared/otel"
)

// Wire set for Auth service
var Set = wire.NewSet(
	config.Load,
	postgres.NewDB,
	postgres.NewUserRepository,
	application.NewAuthService,
	http.NewHandler,
	logger.Init,
	otel.Init,
)

func Initialize(cfg *config.Config) (*http.Handler, func(), error) {
	wire.Build(
		Set,
		wire.Bind(new(ports.UserRepository), new(*postgres.UserRepository)),
	)
	return nil, nil, nil // wire will replace this
}
