package middleware

import (
	"context"
	"net/http"
	"time"

	"pengbook/api/pkg/logger"
)

// TimezoneHeader is the request header carrying the client's IANA timezone
// (e.g. "Asia/Jakarta"). Set on every request by the frontend http client.
const TimezoneHeader = "X-Timezone"

const locationKey contextKey = "client_location"

// LocationFromContext returns the client's timezone from the request context.
// It never returns nil: missing or unknown zones fall back to UTC so
// presentational rendering never fails a request for a cosmetic reason.
func LocationFromContext(ctx context.Context) *time.Location {
	if loc, ok := ctx.Value(locationKey).(*time.Location); ok && loc != nil {
		return loc
	}
	return time.UTC
}

// ContextWithLocation returns a new context carrying the given location.
// Useful for testing handlers that render in the client's timezone.
func ContextWithLocation(ctx context.Context, loc *time.Location) context.Context {
	if loc == nil {
		loc = time.UTC
	}
	return context.WithValue(ctx, locationKey, loc)
}

// Timezone returns a middleware that resolves the client's timezone from the
// X-Timezone header into the request context. Unknown or missing zones fall
// back to UTC. Storage always stays UTC; only calendar rendering follows it.
func Timezone(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := time.UTC
		if tz := r.Header.Get(TimezoneHeader); tz != "" {
			if l, err := time.LoadLocation(tz); err == nil {
				loc = l
			} else {
				log := logger.FromContext(r.Context())
				log.Warn("middleware: invalid timezone header, falling back to UTC", "timezone", tz)
			}
		}

		ctx := context.WithValue(r.Context(), locationKey, loc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
