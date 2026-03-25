//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/adapters/memory"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/adapters/postgres"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/application"
	"github.com/bshongwe/linkpulse/backend/services/auth/internal/presentation/http"
	"github.com/bshongwe/linkpulse/backend/shared/config"
)

// Wire set for Auth service
var Set = wire.NewSet(
	postgres.NewDB,
	postgres.NewUserRepository,
	memory.NewInMemoryTokenBlacklist,
	application.NewTokenService,
	application.NewAuthService,
	http.NewHandler,
	wire.FieldsOf(new(*config.Config), "Database"),
	wire.FieldsOf(new(*config.Config), "JWT"),
)

func Initialize() (*http.Handler, func(), error) {
	wire.Build(
		config.Load,
		Set,
	)
	return nil, nil, nil // wire will replace this
}
