package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go-fitness/external/config"
	"go-fitness/external/logger/sl"
	"go.uber.org/fx"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func NewServer(opts ...ServerOption) *http.Server {
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func NewConfiguredServer(cfg *config.Config, r *chi.Mux) *http.Server {
	return NewServer(
		WithAddr(cfg.HTTPServer.ApiPort),
		WithReadTimeout(cfg.HTTPServer.Timeout),
		WithWriteTimeout(cfg.HTTPServer.Timeout),
		WithIdleTimeout(cfg.HTTPServer.IdleTimeout),
		WithHandler(r),
	)
}

type ServerOption func(*http.Server)

func WithAddr(addr string) ServerOption {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) {
		s.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) {
		s.WriteTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(s *http.Server) {
		s.IdleTimeout = timeout
	}
}

func WithHandler(handler http.Handler) ServerOption {
	return func(s *http.Server) {
		s.Handler = handler
	}
}

func RunServer(
	lc fx.Lifecycle,
	log *slog.Logger,
	server *http.Server,
) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				log.Info("Starting server", sl.Any("config", server.Addr))

				if os.Getenv("JWT_SECRET") == "" {
					panic("JWT_SECRET is not set")
				}

				if err := server.ListenAndServe(); err != nil {
					log.Error("Server failed", sl.Err(err))
				} else {
					log.Info("Server started")
				}
			}()
			return nil
		},

		OnStop: func(ctx context.Context) error {
			log.Error("Server stopped")
			return server.Shutdown(ctx)
		},
	})
}
