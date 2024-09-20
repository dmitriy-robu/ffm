package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"go-fitness/external/logger/sl"
	"log/slog"
	"net/http"
	"runtime"
)

type LoggerMiddleware struct {
	log *slog.Logger
}

func NewLoggerMiddleware(
	log *slog.Logger,
) *LoggerMiddleware {
	return &LoggerMiddleware{
		log: log,
	}
}

func (m *LoggerMiddleware) New() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const op string = "middleware.NewLogger"

		log := m.log.With(
			sl.String("component", op),
		)

		log.Info("logger initialized")

		fn := func(w http.ResponseWriter, r *http.Request) {
			/*	entry := log.With(
					sl.String("method", r.Method),
					sl.String("url", r.URL.Path),
					sl.String("remote_addr", r.RemoteAddr),
					sl.String("user_agent", r.UserAgent()),
					sl.String("request_id", middleware.GetReqID(r.Context())),
				)

				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

				t1 := time.Now()
				defer func() {
					entry.Info("request completed",
						sl.Int("status", ww.Status()),
						sl.Int("size", ww.BytesWritten()),
						sl.String("duration", time.Since(t1).String()),
					)
				}()*/

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			log.Debug("Number of goroutines", sl.Int("goroutines", runtime.NumGoroutine()))

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
