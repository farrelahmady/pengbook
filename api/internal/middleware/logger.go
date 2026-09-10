// Package middleware contains chi HTTP middleware for the application.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"pengbook/api/pkg/logger"
)

// Logger returns a chi middleware that logs every HTTP request (method, path,
// status code, and duration).
//
// It enriches the logger with the request ID from chi's RequestID middleware
// and stores the request-scoped logger in context so that downstream handlers,
// services, and repositories can use logger.FromContext(ctx) to get a logger
// with request_id (and later user_id) automatically attached.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			// Create request-scoped logger with request_id
			reqID := chiMiddleware.GetReqID(r.Context())
			reqLog := log.With("request_id", reqID)

			// Store in context for downstream use (service, repository)
			ctx := logger.WithContext(r.Context(), reqLog)

			next.ServeHTTP(sw, r.WithContext(ctx))

			reqLog.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration", time.Since(start).String(),
			)
		})
	}
}

// statusWriter captures the written HTTP status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status before delegating to the underlying writer.
func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
