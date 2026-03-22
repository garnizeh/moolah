package middleware

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		id := ulid.MustNew(ulid.Timestamp(t), entropy).String()

		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		
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

func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}
