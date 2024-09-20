package middleware

import (
	"go.uber.org/fx"
)

type Middleware struct {
	Logger               *LoggerMiddleware
	ClientAuthMiddleware *ClientAuthMiddleware
	AdminAuthMiddleware  *AdminAuthMiddleware
}

func NewMiddlewares(
	logger *LoggerMiddleware,
	clientAuth *ClientAuthMiddleware,
	adminAuth *AdminAuthMiddleware,
) *Middleware {
	return &Middleware{
		Logger:               logger,
		ClientAuthMiddleware: clientAuth,
		AdminAuthMiddleware:  adminAuth,
	}
}

func NewMiddleware() fx.Option {
	return fx.Module(
		"middleware",
		fx.Provide(
			NewLoggerMiddleware,
			NewClientAuthMiddleware,
			NewAdminAuthMiddleware,
			NewMiddlewares,
		),
	)
}
