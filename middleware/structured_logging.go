package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)

			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unmatched"
			}
			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("route", route),
				slog.Int("status", recorder.status),
				slog.Int64("duration_ms", time.Since(started).Milliseconds()),
			)
		})
	}
}

func SafeRecovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recover() != nil {
					logger.Error("http handler panic",
						slog.String("method", r.Method),
						slog.String("route", chi.RouteContext(r.Context()).RoutePattern()),
					)
					writeInternalError(w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func writeInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"message":"internal server error","content":null}`))
}
