package handler

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/http/handler/fileupload"
	"log/slog"
	"net/http"
	"strconv"
)

type GoalHandler struct {
	log         *slog.Logger
	goalService GoalUploadService
	validation  *validator.Validate
	localizer   *i18n.Localizer
}

type GoalUploadService interface {
	Upload(ctx context.Context, goalID int, videData *data.VideoData) error
}

func NewGoalHandler(
	log *slog.Logger,
	goalService GoalUploadService,
	validator *validator.Validate,
	localizer *i18n.Localizer,
) *GoalHandler {
	return &GoalHandler{
		log:         log,
		validation:  validator,
		localizer:   localizer,
		goalService: goalService,
	}
}

func (h *GoalHandler) VideoUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "GoalHandler.VideoUpload"

		log := h.log.With(
			sl.String("op", op),
		)

		log.Info("processing upload")

		ctx := r.Context()

		videoData, err := fileupload.ParseAndExtractFile(r, log)
		if err != nil {
			response.Respond(w, response.Response{
				Status:  http.StatusRequestEntityTooLarge,
				Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "internal_server_error"}),
				Data:    err.Error(),
			})
			return
		}

		goalID, _ := strconv.Atoi(chi.URLParam(r, "id"))

		if err = h.goalService.Upload(ctx, goalID, videoData); err != nil {
			log.Error("failed_to_process_upload", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusConflict,
				Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: err.Error()}),
			})
			return
		}

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "goal_updated_successfully"}),
		})
		return
	}
}
