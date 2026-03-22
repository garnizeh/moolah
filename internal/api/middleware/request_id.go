package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type contextKey string

const requestIDKey contextKey = "request_id"

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// RequestID is a middleware that generates a unique request ID for each request.
func RequestID(next http.Handler) http.Handler {
	var (
		entropy   = ulid.Monotonic(rand.Reader, 0)
		entropyMu sync.Mutex
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		entropyMu.Lock()
		id, err := ulid.New(ulid.Timestamp(start), entropy)
		entropyMu.Unlock()

		var requestID string
		if err != nil {
			slog.Error("failed to generate ULID", slog.Any("error", err))
			// Fallback to a timestamp-based ID to ensure continuity
			requestID = fmt.Sprintf("req-%d", start.UnixNano())
		} else {
			requestID = id.String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// Set header for HTMX/Clients to see the request ID
		w.Header().Set("X-Request-ID", requestID)

		// Create a logger with the request ID
		logger := slog.Default().With(slog.String("request_id", requestID))

		// Wrap response writer to capture status
		sw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r.WithContext(ctx))

		logger.Info("request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status", sw.status),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

// FromContext retrieves the request ID from the context.
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}
