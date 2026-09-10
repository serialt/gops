//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"
	"net/http"

	"gops/internal/biz"
	"gops/internal/config"
	"gops/internal/data"
	"gops/internal/server"

	"github.com/google/wire"
)

//go:generate wire
func InitializeServer(cfg *config.Config, log *slog.Logger) (*http.Server, func(), error) {
	wire.Build(
		data.ProviderSet,
		biz.ProviderSet,
		server.ProviderSet,
	)

	return nil, nil, nil
}
