// Package log provides logging utilities for the application.
package log

import (
	"io"
	"log/slog"
	"os"
)

// InitWithWriter initializes the logger with the given writer.
// It uses JSON format in production and text format in development.
// It sets the default logger to be used by other packages.
func InitWithWriter(w io.Writer) {
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
