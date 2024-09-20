package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-fitness/internal/api/http/handler"
	md "go-fitness/internal/api/http/middleware"
	"net/http"
	"time"
)

func NewRouter(
	handlers *handler.Handlers,
	md *md.Middleware,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.Timeout(10 * time.Minute))
	r.Use(md.Logger.New())
	r.Use(cors())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))

		return
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/videos/AQmjberxHcZj4ck", func(r chi.Router) {
			// FOR TESTING PURPOSES without auth
			r.Group(func(r chi.Router) {
				// r.Get("/drive", handlers.Drive.ProcessParse())
				// r.Get("/drive/auth", handlers.Drive.InitiateAuth())
				// r.Get("/oauth2callback", handlers.Drive.HandleAuthCallback())

				r.Get("/delete-files", handlers.Video.DeleteVideoFilesIfDoesntExistInTable())
				r.Get("/posters", handlers.Poster.CreatePosterFromUploadedTsFiles())
				r.Get("/poster/{uuid}", handlers.Poster.GetPosterByUUID())

				r.Get("/{uuid}", handlers.Video.GetVideo())
				r.Get("/{uuid}/{resolution}", handlers.Video.GetVideo())

				r.Post("/store", handlers.Workout.StoreWorkout())

				r.Post("/portal/store", handlers.Portal.VideoUpload())

				r.Put("/goals/{id}/update", handlers.Goal.VideoUpload())

				r.Post("/tabs/store", handlers.Tab.Store())
				r.Put("/tabs/{id}/update", handlers.Tab.UpdateVideo())
			})
		})

		r.Route("/portal/ms/videos", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Get("/{uuid}", handlers.Video.GetVideo())
				r.Get("/{uuid}/{resolution}", handlers.Video.GetVideo())
			})
		})

		r.Route("/admin/ms/videos", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(md.AdminAuthMiddleware.New())
				r.Get("/{uuid}", handlers.Video.GetVideo())
				r.Get("/{uuid}/{resolution}", handlers.Video.GetVideo())
			})
		})

		r.Route("/admin/ms/workouts", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				//r.Use(md.AdminAuthMiddleware.New())
				r.Post("/store", handlers.Workout.StoreWorkout())
			})
		})

		r.Route("/admin/ms/goals", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(md.AdminAuthMiddleware.New())
				r.Put("/{id}/update", handlers.Goal.VideoUpload())
			})
		})

		r.Route("/admin/ms/tabs", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(md.AdminAuthMiddleware.New())
				r.Post("/store", handlers.Tab.Store())
				r.Put("/{id}/update", handlers.Tab.UpdateVideo())
			})
		})

		r.Route("/client/ms/videos", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(md.ClientAuthMiddleware.New())
				r.Get("/{uuid}", handlers.Video.GetVideo())
				r.Get("/{uuid}/{resolution}", handlers.Video.GetVideo())
			})
		})
	})

	return r
}

func cors() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}
}
