package handler

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
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
	"strconv"
)

type TabHandler struct {
	log        *slog.Logger
	tabService TabService
	validation *validator.Validate
	localizer  *i18n.Localizer
}

type TabService interface {
	Store(context.Context, data.TabData) error
	UpdateVideo(context.Context, int, *data.VideoData) error
}

func NewTabHandler(
	log *slog.Logger,
	tabService TabService,
	localizer *i18n.Localizer,
	validator *validator.Validate,
) *TabHandler {
	return &TabHandler{
		log:        log,
		validation: validator,
		localizer:  localizer,
		tabService: tabService,
	}
}

func (h *TabHandler) Store() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "TabHandler.StoreGoal"

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

		req := request.TabRequest{
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
		}

		var validateErr validator.ValidationErrors
		if err = h.validation.Struct(req); err != nil {
			errors.As(err, &validateErr)
			log.Error("invalid request", sl.Err(validateErr))
			response.Respond(w, response.Response{
				Status:  http.StatusBadRequest,
				Message: validation.ValidationError(h.localizer, validateErr).Error(),
			})
			return
		}

		log.Info("uploading video")

		if err = h.tabService.Store(ctx, data.TabData{
			Name:        req.Name,
			Description: req.Description,
			VideoData:   videoData,
		}); err != nil {
			log.Error("failed_to_process_upload", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusConflict,
				Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: err.Error()}),
			})
			return
		}

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "tab_stored_successfully"}),
		})
		return
	}
}

func (h *TabHandler) UpdateVideo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "TabHandler.UpdateGoal"

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

		tabID, _ := strconv.Atoi(chi.URLParam(r, "id"))

		if err = h.tabService.UpdateVideo(ctx, tabID, videoData); err != nil {
			log.Error("failed_to_process_upload", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusConflict,
				Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: err.Error()}),
			})
			return
		}

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "tab_updated_successfully"}),
		})
	}
}
