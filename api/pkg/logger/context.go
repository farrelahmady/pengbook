package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// WithContext stores a logger in the context.
func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// FromContext extracts the logger from the context.
// If no logger is found, it returns slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}
