// handler/drive.go

package handler

import (
	"context"
	"go-fitness/external/logger/sl"
	"go-fitness/external/response"
	"log/slog"
	"net/http"
)

type DriveHandler struct {
	log          *slog.Logger
	driveManager GoogleDriveService
}

type GoogleDriveService interface {
	GetAuthURL(ctx context.Context) string
	ExchangeCodeForToken(ctx context.Context, code string) error
	ProcessParse(ctx context.Context)
}

func NewDriveHandler(
	log *slog.Logger,
	driveManager GoogleDriveService,
) *DriveHandler {
	return &DriveHandler{
		log:          log,
		driveManager: driveManager,
	}
}

// InitiateAuth Endpoint to initiate Google OAuth authentication
func (h *DriveHandler) InitiateAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op string = "handler.DriveHandler.InitiateAuth"

		log := h.log.With(
			sl.String("op", op),
		)

		log.Info("Initiating Google OAuth authentication")

		ctx := context.Background()
		authURL := h.driveManager.GetAuthURL(ctx)

		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// HandleAuthCallback Endpoint to handle Google OAuth callback and save the token
func (h *DriveHandler) HandleAuthCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op string = "handler.DriveHandler.HandleAuthCallback"

		log := h.log.With(
			sl.String("op", op),
		)

		log.Info("Handling Google OAuth callback")

		ctx := context.Background()
		code := r.URL.Query().Get("code")

		if code == "" {
			log.Error("No code in the request")
			http.Error(w, "No code in the request", http.StatusBadRequest)
			return
		}

		err := h.driveManager.ExchangeCodeForToken(ctx, code)
		if err != nil {
			log.Error("Unable to exchange code for token", sl.Err(err))
			http.Error(w, "Unable to exchange code for token", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/api/v1/drive", http.StatusFound)
	}
}

// ProcessParse Endpoint to process parsing Google Drive
func (h *DriveHandler) ProcessParse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op string = "handler.DriveHandler.ProcessParse"

		log := h.log.With(
			sl.String("op", op),
		)

		log.Info("Processing parse")

		ctx := context.Background()
		go h.driveManager.ProcessParse(ctx)

		response.Respond(w, response.Response{
			Status:  http.StatusOK,
			Message: "ok",
		})
	}
}
