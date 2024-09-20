package service

import (
	"context"
	"errors"
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/types"
	"log/slog"
)

type WorkoutService struct {
	log          *slog.Logger
	workoutRepo  WorkoutRepository
	videoService VideoServiceInterface
}

type VideoServiceInterface interface {
	ProcessUpload(ctx context.Context, data data.VideoData) (int64, error)
}

type WorkoutRepository interface {
	GetWorkoutByName(context.Context, string) (types.Workout, error)
	Create(context.Context, types.Workout) (int64, error)
	AddWorkoutToProgramMonth(context.Context, int64, int64) error
	CheckIfNameExists(context.Context, string) bool
	Update(context.Context, int64, types.Workout) error
}

func NewWorkoutService(
	log *slog.Logger,
	workoutRepo WorkoutRepository,
	videoService VideoServiceInterface,
) *WorkoutService {
	return &WorkoutService{
		log:          log,
		workoutRepo:  workoutRepo,
		videoService: videoService,
	}
}

func (s *WorkoutService) ProcessWorkout(ctx context.Context, data data.WorkoutData) error {
	const op string = "WorkoutService.ProcessWorkout"

	log := s.log.With(
		sl.String("op", op),
	)

	if ok := s.workoutRepo.CheckIfNameExists(ctx, data.Name); ok {
		log.Warn("workout already exists")
		return errors.New("workout_already_exists")
	}

	if data.VideoData == nil {
		log.Error("video data is nil")
		return errors.New("invalid_video_data")
	}

	videoID, err := s.videoService.ProcessUpload(ctx, *data.VideoData)
	if err != nil {
		log.Error("failed to process upload", sl.Err(err))
		return err
	}

	return s.StoreOrUpdate(ctx, data, videoID)
}

func (s *WorkoutService) StoreOrUpdate(ctx context.Context, data data.WorkoutData, videoID int64) error {
	const op string = "WorkoutService.StoreOrUpdate"

	log := s.log.With(
		sl.String("op", op),
		sl.Any("data", data),
		sl.Int64("video_id", videoID),
	)

	workout := types.Workout{
		Name:        data.Name,
		Description: data.Description,
		VideoID:     videoID,
	}

	w, err := s.workoutRepo.GetWorkoutByName(ctx, data.Name)
	if err == nil {
		if err = s.workoutRepo.Update(ctx, w.ID, workout); err != nil {
			log.Error("failed to update workout", sl.Err(err))
			return errors.New("failed_to_update_workout")
		}
	} else {
		workoutID, err := s.workoutRepo.Create(ctx, workout)
		if err != nil {
			log.Error("failed to create workout", sl.Err(err))
			return errors.New("failed_to_create_workout")
		}

		if data.ProgramMonthID != nil {
			if err = s.workoutRepo.AddWorkoutToProgramMonth(ctx, workoutID, *data.ProgramMonthID); err != nil {
				log.Error("failed to add workout to program month", sl.Err(err))
				return errors.New("failed_to_add_workout_to_program_month")
			}
		}
	}

	return nil
}
