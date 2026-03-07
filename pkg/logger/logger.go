// Package logger provides a thin wrapper around log/slog for structured logging.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New returns a *slog.Logger configured for the given level and format.
// It accepts an optional io.Writer (w) for testing; if nil, it defaults to os.Stdout.
// level:  "debug" | "info" | "warn" | "error"  (default: "info")
// format: "json"  | "text"                     (default: "json")
func New(w io.Writer, level, format string) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}

	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if strings.ToLower(format) == "text" {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}

	return slog.New(handler)
}
