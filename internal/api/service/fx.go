package service

import (
	"go-fitness/internal/api/http/handler"
	"go-fitness/internal/api/service/video"
	"go.uber.org/fx"
)

func NewService() fx.Option {
	return fx.Module(
		"service",
		fx.Provide(
			video.NewVideoService,
			NewWorkoutService,
			NewUserService,
			//video.NewWorkerPool,
			//video.NewTranscodeService,

			fx.Annotate(
				video.NewWorkerPool,
				fx.As(new(video.TaskQueue)),
			),

			fx.Annotate(
				video.NewTranscodeService,
				fx.As(new(video.Transcoder)),
			),

			fx.Annotate(
				video.NewVideoService,
				fx.As(new(handler.VideoService)),
				//fx.As(new(video.UploadAndTranscodeQueueInterface)),
				fx.As(new(VideoServiceInterface)),
				fx.As(new(handler.PosterService)),
			),

			fx.Annotate(
				NewWorkoutService,
				fx.As(new(handler.WorkoutService)),
			),

			fx.Annotate(
				NewGoogleDriveService,
				fx.As(new(handler.GoogleDriveService)),
			),

			fx.Annotate(
				NewGoalService,
				fx.As(new(handler.GoalUploadService)),
			),

			fx.Annotate(
				NewTabService,
				fx.As(new(handler.TabService)),
			),

			fx.Annotate(
				NewPortalService,
				fx.As(new(handler.PortalService)),
			),
		),
	)
}
