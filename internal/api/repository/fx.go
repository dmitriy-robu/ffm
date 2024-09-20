package repository

import (
	"go-fitness/internal/api/service"
	"go-fitness/internal/api/service/video"
	"go.uber.org/fx"
)

func NewRepository() fx.Option {
	return fx.Module(
		"repository",
		fx.Provide(
			fx.Annotate(
				NewVideoRepository,
				fx.As(new(video.VideoRepository)),
				fx.As(new(video.UpdateVideoStatus)),
			),

			fx.Annotate(
				NewGoalRepository,
				fx.As(new(service.GoalRepository)),
			),

			fx.Annotate(
				NewUserRepository,
				fx.As(new(service.UserRepository)),
			),

			fx.Annotate(
				NewWorkoutRepository,
				fx.As(new(service.WorkoutRepository)),
			),

			fx.Annotate(
				NewProgramRepository,
				fx.As(new(service.ProgramRepository)),
			),

			fx.Annotate(
				NewTabRepository,
				fx.As(new(service.TabRepository)),
			),

			fx.Annotate(
				NewPortalRepository,
				fx.As(new(service.PortalRepository)),
			),
		),
	)
}
