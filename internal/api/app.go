package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/patrickmn/go-cache"
	"go-fitness/external/config"
	"go-fitness/external/db"
	"go-fitness/internal/api/http/handler"
	"go-fitness/internal/api/http/middleware"
	"go-fitness/internal/api/repository"
	"go-fitness/internal/api/service"
	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		fx.Options(
			repository.NewRepository(),
			service.NewService(),
			handler.NewHandler(),
			middleware.NewMiddleware(),
			db.NewDataBase(),
		),
		fx.Provide(
			config.NewConfig,
			validator.New,
			NewCache,
			NewLogger,
			NewRouter,
			NewConfiguredServer,
			NewBundle,
			NewLocalizer,
		),
		fx.Invoke(RunServer),
		//fx.Invoke(cmd.NewCmd),
	)
}

func NewCache() *cache.Cache {
	return cache.New(cache.NoExpiration, cache.NoExpiration)
}
