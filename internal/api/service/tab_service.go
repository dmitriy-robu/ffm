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

type TabService struct {
	log          *slog.Logger
	videoService *video.VideoService
	tabRepo      TabRepository
}

type TabRepository interface {
	Store(context.Context, types.Tab) error
	CheckIfNameExists(context.Context, string) bool
	GetByID(context.Context, int) (types.Tab, error)
	Update(context.Context, types.Tab) error
}

func NewTabService(
	log *slog.Logger,
	videoService *video.VideoService,
	tabRepo TabRepository,
) *TabService {
	return &TabService{
		log:          log,
		videoService: videoService,
		tabRepo:      tabRepo,
	}
}

func (s *TabService) Store(ctx context.Context, data data.TabData) error {
	const op = "TabService.Store"

	log := s.log.With(
		sl.String("op", op),
	)

	log.Info("storing")

	ok := s.tabRepo.CheckIfNameExists(ctx, data.Name)
	if ok {
		log.Error("tab already exists")
		return errors.New("tab_already_exists")
	}

	videoID, err := s.videoService.ProcessUpload(ctx, *data.VideoData)
	if err != nil {
		log.Error("failed to process upload", sl.Err(err))
		return err
	}

	if err := s.tabRepo.Store(ctx, types.Tab{
		Name:        data.Name,
		Description: data.Description,
		VideoID:     &videoID,
	}); err != nil {
		log.Error("failed to store tab", sl.Err(err))
		return errors.New("failed_to_store_tab")
	}

	return nil
}

func (s *TabService) UpdateVideo(ctx context.Context, tabID int, videoData *data.VideoData) error {
	const op = "TabService.Update"

	log := s.log.With(
		sl.String("op", op),
		sl.Int("tabID", tabID),
	)

	log.Info("updating")

	tab, err := s.tabRepo.GetByID(ctx, tabID)
	if err != nil {
		log.Error("failed to get tab", sl.Err(err))
		return errors.New("failed_to_get_tab")
	}

	var videoID int64

	if videoData != nil {
		videoID, err = s.videoService.ProcessUpload(ctx, *videoData)
		if err != nil {
			log.Error("failed to process upload", sl.Err(err))
			return err
		}
	}

	tab.VideoID = &videoID

	if err := s.tabRepo.Update(ctx, tab); err != nil {
		log.Error("failed to update tab", sl.Err(err))
		return errors.New("failed_to_update_tab")
	}

	return nil
}
