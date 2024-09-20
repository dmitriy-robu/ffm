package db

import (
	"context"
	"database/sql"
	"fmt"
	"go-fitness/external/config"
	"go-fitness/external/logger/sl"
	"go.uber.org/fx"
	"log/slog"
)

func NewMysqlDatabase(lc fx.Lifecycle, log *slog.Logger, cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4,utf8&parseTime=True&loc=Local",
		cfg.DB.MysqlUser,
		cfg.DB.MysqlPassword,
		cfg.DB.MysqlHost+":"+cfg.DB.MysqlPort,
		cfg.DB.MysqlDBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("failed to open database connection", sl.Err(err))
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err = db.Ping(); err != nil {
		log.Error("failed to ping database", sl.Err(err))
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Starting database connection")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing database connection")
			return db.Close()
		},
	})

	return db, nil
}
