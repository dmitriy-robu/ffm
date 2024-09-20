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

type PortalService struct {
	log          *slog.Logger
	videoService *video.VideoService
	portalRepo   PortalRepository
}

type PortalRepository interface {
	Store(context.Context, types.Portal) error
}

func NewPortalService(
	log *slog.Logger,
	videoService *video.VideoService,
	portalRepo PortalRepository,
) *PortalService {
	return &PortalService{
		log:          log,
		videoService: videoService,
		portalRepo:   portalRepo,
	}
}

func (s *PortalService) Upload(ctx context.Context, videoData *data.VideoData) error {
	const op = "PortalService.Update"

	log := s.log.With(
		sl.String("op", op),
	)

	videoID, err := s.videoService.ProcessUpload(ctx, *videoData)
	if err != nil {
		log.Error("failed to process upload", sl.Err(err))
		return err
	}

	portal := types.Portal{
		VideoID: videoID,
		Name:    "Cine sunt",
	}

	if err := s.portalRepo.Store(ctx, portal); err != nil {
		log.Error("failed to store portal video", sl.Err(err))
		return errors.New("failed_to_store")
	}

	return nil
}
