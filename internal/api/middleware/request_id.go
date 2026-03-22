package middleware

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID is a middleware that generates a unique request ID for each request.
func RequestID(next http.Handler) http.Handler {
	var (
		entropy   = ulid.Monotonic(rand.Reader, 0)
		entropyMu sync.Mutex
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		entropyMu.Lock()
		id := ulid.MustNew(ulid.Timestamp(t), entropy).String()
		entropyMu.Unlock()

		ctx := context.WithValue(r.Context(), requestIDKey, id)

		// Set header for HTMX/Clients to see the request ID
		w.Header().Set("X-Request-ID", id)

		// Create a logger with the request ID
		logger := slog.Default().With(slog.String("request_id", id))

		next.ServeHTTP(w, r.WithContext(ctx))

		logger.Info("request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
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
