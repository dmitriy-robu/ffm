package db

import "go.uber.org/fx"

func NewDataBase() fx.Option {
	return fx.Module(
		"database",
		fx.Provide(
			NewMysqlDatabase,
			fx.Annotate(
				NewMysql,
				fx.As(new(SqlInterface)),
			),
		),
	)
}
