package logger

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// parameter that will be outputted with each log
		log = log.With(
			slog.String("component", "middleware/logger"),
		)
		// at the start
		log.Info("Logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			// this part adds stuff before the request
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			// after the whole request, this deffer kicks in
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			// go to next handler
			next.ServeHTTP(ww, r)
		}

		// return the handler itselft
		return http.HandlerFunc(fn)
	}
}
