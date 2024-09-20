package handler

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type VideoHandler struct {
	log          *slog.Logger
	videoService VideoService
	validation   *validator.Validate
	localizer    *i18n.Localizer
}

type VideoService interface {
	ProcessGetVideoPlayListByUUID(context.Context, string) ([]byte, error)
	ProcessGetVideoTS(context.Context, string) ([]byte, error)
	ProcessGetVideoM3U8(string) ([]byte, error)
	DeleteAllVideoFilesIfDoestExistInTable(context.Context)
}

func NewVideoHandler(
	log *slog.Logger,
	videoService VideoService,
	localizer *i18n.Localizer,
	validator *validator.Validate,
) *VideoHandler {
	return &VideoHandler{
		log:          log,
		videoService: videoService,
		validation:   validator,
		localizer:    localizer,
	}
}

func (h *VideoHandler) DeleteVideoFilesIfDoesntExistInTable() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "VideoHandler.DeleteAllVideos"

		log := h.log.With(
			sl.String("op", op),
		)

		log.Info("deleting all videos")

		go func() {
			ctx := context.WithoutCancel(r.Context())
			h.videoService.DeleteAllVideoFilesIfDoestExistInTable(ctx)
		}()

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: "ok",
			Data:    "ok",
		})
		return
	}
}

// GetVideo returns video by uuid
func (h *VideoHandler) GetVideo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op string = "VideoHandler.GetVideo"

		log := h.log.With(
			sl.String("op", op),
		)

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var (
			video []byte
			err   error
		)

		switch {
		case strings.Contains(r.URL.Path, ".ts"):
			video, err = h.videoService.ProcessGetVideoTS(ctx, r.URL.Path)
			w.Header().Set("Content-Type", "video/mp2ts")
		case strings.Contains(r.URL.Path, ".m3u8"):
			video, err = h.videoService.ProcessGetVideoM3U8(r.URL.Path)
			w.Header().Set("Content-Type", "application/x-mpegURL")
			w.Header().Set("Cache-Control", "no-cache")
		default:
			videoUUID := chi.URLParam(r, "uuid")
			if videoUUID == "" {
				log.Error("uuid is required")
				http.Error(w, "UUID is required", http.StatusBadRequest)
				return
			}

			video, err = h.videoService.ProcessGetVideoPlayListByUUID(ctx, videoUUID)
			w.Header().Set("Content-Type", "application/x-mpegURL")
			w.Header().Set("Cache-Control", "no-cache")
		}

		if err != nil {
			log.Error("failed to get video", sl.Err(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Length", strconv.Itoa(len(video)))

		select {
		case <-ctx.Done():
			log.Warn("context canceled before writing video")
			return
		default:
			if _, err := w.Write(video); err != nil {
				if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
					log.Warn("Client disconnected before response could be sent", sl.Err(err))
					return
				}
				log.Error("Failed to write video", sl.Err(err))
				return
			}
		}
	}
}
