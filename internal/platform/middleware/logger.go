package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/pkg/ulid"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}
	return n, nil
}

// RequestLogger returns a middleware that logs each request as a structured slog entry.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx := r.Context()

			// Generate request ID
			requestID := ulid.New()
			w.Header().Set("X-Request-ID", requestID)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			tenantID, ok := TenantIDFromCtx(ctx)
			if !ok {
				slog.WarnContext(ctx, "logger: failed to extract tenant ID from context, defaulting to unknown")
				tenantID = "unknown"
			}

			userID, ok := UserIDFromCtx(ctx)
			if !ok {
				slog.WarnContext(ctx, "logger: failed to extract user ID from context, defaulting to anonymous")
				userID = "anonymous"
			}

			slog.InfoContext(ctx, "request",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Int64("latency_ms", time.Since(start).Milliseconds()),
				slog.String("tenant_id", tenantID),
				slog.String("user_id", userID),
			)
		})
	}
}
