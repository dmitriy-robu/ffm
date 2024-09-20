package handler

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/http/handler/fileupload"
	"log/slog"
	"net/http"
)

type PortalHandler struct {
	log           *slog.Logger
	portalService PortalService
	validation    *validator.Validate
	localizer     *i18n.Localizer
}

type PortalService interface {
	Upload(context.Context, *data.VideoData) error
}

func NewPortalHandler(
	log *slog.Logger,
	portalService PortalService,
	validator *validator.Validate,
	localizer *i18n.Localizer,
) *PortalHandler {
	return &PortalHandler{
		log:           log,
		validation:    validator,
		localizer:     localizer,
		portalService: portalService,
	}
}

func (h *PortalHandler) VideoUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "PortalHandler.VideoUpload"

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

		log.Info("uploading video")

		if err = h.portalService.Upload(ctx, videoData); err != nil {
			log.Error("failed_to_process_upload", sl.Err(err))
			response.Respond(w, response.Response{
				Status:  http.StatusConflict,
				Message: h.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: err.Error()}),
			})
			return
		}

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: "ok",
		})
		return
	}
}
