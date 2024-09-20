package handler

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"go-fitness/external/validation"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/http/handler/fileupload"
	"go-fitness/internal/api/http/request"
	"log/slog"
	"net/http"
)

type WorkoutHandler struct {
	log            *slog.Logger
	workoutService WorkoutService
	validation     *validator.Validate
	localizer      *i18n.Localizer
}

type WorkoutService interface {
	ProcessWorkout(ctx context.Context, data data.WorkoutData) error
}

func NewWorkoutHandler(
	log *slog.Logger,
	workoutService WorkoutService,
	localizer *i18n.Localizer,
	validator *validator.Validate,
) *WorkoutHandler {
	return &WorkoutHandler{
		log:            log,
		workoutService: workoutService,
		validation:     validator,
		localizer:      localizer,
	}
}

// StoreWorkout processes the video upload
// It returns a http.HandlerFunc
func (h *WorkoutHandler) StoreWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "WorkoutHandler.StoreWorkout"

		log := h.log.With(
			sl.String("op", op),
		)

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

		req := request.WorkoutRequest{
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
		}

		var validateErr validator.ValidationErrors
		if err := h.validation.Struct(req); err != nil {
			errors.As(err, &validateErr)
			log.Error("invalid request", sl.Err(validateErr))
			response.Respond(w, response.Response{
				Status:  http.StatusBadRequest,
				Message: validation.ValidationError(h.localizer, validateErr).Error(),
			})
			return
		}

		workoutData := data.WorkoutData{
			Name:        req.Name,
			Description: req.Description,
			VideoData:   videoData,
		}

		log.Info("uploading video")

		if err = h.workoutService.ProcessWorkout(ctx, workoutData); err != nil {
			log.Error("failed_to_process_upload", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusConflict,
				Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: err.Error()}),
			})
			return
		}

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "workout_created_successfully"}),
		})
		return
	}
}
