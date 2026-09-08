// Package logger provides creation of a structured logger (log/slog) with
// JSON output.
package logger

import (
	"log/slog"
	"os"
)

// New creates a *slog.Logger with JSON output to stdout.
//
// Used in cmd/api/main.go and injected into the server & middleware.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
