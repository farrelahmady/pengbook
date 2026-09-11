// Package logger provides creation of a structured logger (log/slog) with
// JSON output.
package logger

import (
	"log/slog"
	"os"
)

// New creates a *slog.Logger with JSON output to stdout.
//
// The logLevel parameter accepts: "debug", "info", "warn", "error".
// Defaults to "info" if not recognized.
//
// Used in cmd/api/main.go and injected into the server & middleware.
func New(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
