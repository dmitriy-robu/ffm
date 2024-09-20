package service

import (
	"context"
	"errors"
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/service/video"
	"go-fitness/internal/api/types"
	"log/slog"
)

type GoalService struct {
	log          *slog.Logger
	videoService *video.VideoService
	goalRepo     GoalRepository
}

type GoalRepository interface {
	Store(ctx context.Context, goal types.Goal) error
	CheckIfNameExists(ctx context.Context, name string) bool
	GetByID(ctx context.Context, id int) (types.Goal, error)
	Update(ctx context.Context, goal types.Goal) error
}

func NewGoalService(
	log *slog.Logger,
	videoService *video.VideoService,
	goalRepo GoalRepository,
) *GoalService {
	return &GoalService{
		log:          log,
		videoService: videoService,
		goalRepo:     goalRepo,
	}
}

func (s *GoalService) Upload(ctx context.Context, goalID int, videoData *data.VideoData) error {
	const op = "GoalService.Update"

	log := s.log.With(
		sl.String("op", op),
		sl.Int("goal_id", goalID),
	)

	log.Info("updating goal")

	goal, err := s.goalRepo.GetByID(ctx, goalID)
	if err != nil {
		log.Error("failed to get goal", sl.Err(err))
		return errors.New("failed_to_get_goal")
	}

	videoID, err := s.videoService.ProcessUpload(ctx, *videoData)
	if err == nil {
		goal.VideoID = &videoID
	}

	if err := s.goalRepo.Update(ctx, goal); err != nil {
		log.Error("failed to update goal", sl.Err(err))
		return errors.New("failed_to_update_goal")
	}

	return nil
}
