// Package middleware contains chi HTTP middleware for the application.
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger returns a chi middleware that logs every HTTP request (method, path,
// status code, and duration).
//
// A statusWriter wraps the ResponseWriter so the real status code can be
// captured even when the handler does not call WriteHeader explicitly.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			log.Info("http request",
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
