package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type PosterHandler struct {
	log           *slog.Logger
	posterService PosterService
	validation    *validator.Validate
	localizer     *i18n.Localizer
}

type PosterService interface {
	CreatePosterFromUploadedTsFiles(ctx context.Context) error
	GetPosterByUUID(ctx context.Context, uuid string) (string, error)
}

func NewPosterHandler(
	log *slog.Logger,
	posterService PosterService,
	localizer *i18n.Localizer,
	validator *validator.Validate,
) *PosterHandler {
	return &PosterHandler{
		log:           log,
		posterService: posterService,
		validation:    validator,
		localizer:     localizer,
	}
}

// CreatePosterFromUploadedTsFiles creates poster from uploaded ts files
func (h *PosterHandler) CreatePosterFromUploadedTsFiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "PosterHandler.CreatePosterFromUploadedTsFiles"

		log := h.log.With(
			sl.String("op", op),
		)

		log.Info("creating poster from uploaded ts files")

		go func() {
			ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
			defer cancel()
			err := h.posterService.CreatePosterFromUploadedTsFiles(ctx)
			if err != nil {
				log.Error("failed to create poster from uploaded ts files", sl.Err(err))
				return
			}
		}()

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: "ok",
			Data:    "ok",
		})
		return
	}
}

// GetPosterByUUID returns poster by uuid
func (h *PosterHandler) GetPosterByUUID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "PosterHandler.GetPosterByUUID"

		log := h.log.With(
			sl.String("op", op),
		)

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		uuid := chi.URLParam(r, "uuid")
		if uuid == "" {
			log.Error("uuid is required")
			response.Respond(w, response.Response{
				Status:  http.StatusInternalServerError,
				Message: "internal server error",
				Data:    "uuid is required",
			})
			return
		}

		poster, err := h.posterService.GetPosterByUUID(ctx, uuid)
		if err != nil {
			log.Error("failed to get poster", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusInternalServerError,
				Message: "internal server error",
				Data:    err.Error(),
			})
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "image/jpeg")

		posterFile, err := os.Open(poster)
		if err != nil {
			log.Error("failed to open poster file", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusInternalServerError,
				Message: "internal server error",
				Data:    err.Error(),
			})
			return
		}
		defer posterFile.Close()

		_, err = io.Copy(w, posterFile)
		if err != nil {
			log.Error("failed to write poster", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusInternalServerError,
				Message: "internal server error",
				Data:    err.Error(),
			})
			return
		}
	}
}
