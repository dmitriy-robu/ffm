package handler

import (
	"go.uber.org/fx"
)

type Handlers struct {
	Video   *VideoHandler
	Workout *WorkoutHandler
	Drive   *DriveHandler
	Goal    *GoalHandler
	Tab     *TabHandler
	Poster  *PosterHandler
	Portal  *PortalHandler
}

func NewHandlers(
	video *VideoHandler,
	workout *WorkoutHandler,
	drive *DriveHandler,
	goal *GoalHandler,
	tab *TabHandler,
	poster *PosterHandler,
	portal *PortalHandler,
) *Handlers {
	return &Handlers{
		Video:   video,
		Workout: workout,
		Drive:   drive,
		Goal:    goal,
		Tab:     tab,
		Poster:  poster,
		Portal:  portal,
	}
}

func NewHandler() fx.Option {
	return fx.Module(
		"handler",
		fx.Options(),
		fx.Provide(
			NewPortalHandler,
			NewDriveHandler,
			NewVideoHandler,
			NewWorkoutHandler,
			NewGoalHandler,
			NewTabHandler,
			NewPosterHandler,
			NewHandlers,
		),
	)
}
