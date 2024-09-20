package api

import (
	"go-fitness/external/config"
	"go-fitness/external/logger/handler/slogpretty"
	"log/slog"
	"os"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	var log *slog.Logger

	switch cfg.Env {
	case cfg.ENVState.Local:
		log = setupPrettySlogLogger()
	case cfg.ENVState.Dev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case cfg.ENVState.Prod:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlogLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	newPrettyHandler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(newPrettyHandler)
}
