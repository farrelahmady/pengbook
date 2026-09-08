package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"pengbook/api/pkg/response"
)

// Recovery returns a chi middleware that recovers from panics, logs the panic
// and stack trace, and responds with a generic 500 instead of crashing.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
