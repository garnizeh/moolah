package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	idempotencyHeader = "Idempotency-Key"
	idempotencyTTL    = 24 * time.Hour
)

// CachedResponse represents a stored HTTP response.
type CachedResponse struct {
	Body       []byte `json:"body"`
	StatusCode int    `json:"status_code"`
}

// IdempotencyStore defines the contract for storing idempotency keys and responses.
type IdempotencyStore interface {
	Get(ctx context.Context, key string) (*CachedResponse, error)
	SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error)
	SetResponse(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error
}

// responseRecorder captures the response to be cached.
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	n, err := r.ResponseWriter.Write(b)
	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}
	return n, nil
}

// Idempotency middleware ensures that POST requests are processed at most once.
func Idempotency(store IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			clientKey := r.Header.Get(idempotencyHeader)
			if clientKey == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error": "missing_idempotency_key"}`))
				return
			}

			// Validate key length to prevent abuse
			if len(clientKey) > 128 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error": "invalid_idempotency_key"}`))
				return
			}

			// Extract userID from context using the global key from auth.go.
			userID, _ := UserIDFromCtx(r.Context())
			if userID == "" {
				userID = "anonymous"
			}

			// Composite key: idempotency:{userID}:{clientKey}
			redisKey := fmt.Sprintf("idempotency:%s:%s", userID, clientKey)

			// 1. Check if we have a cached response
			cached, err := store.Get(r.Context(), redisKey)
			if err != nil {
				http.Error(w, "internal_error", http.StatusInternalServerError)
				return
			}

			if cached != nil {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cached.StatusCode)
				// gosec G705: This is a cached response from our own store, not raw user input.
				/* #nosec G705 */
				_, _ = w.Write(cached.Body)
				return
			}

			// 2. Try to acquire lock
			ok, err := store.SetLocked(r.Context(), redisKey, idempotencyTTL)
			if err != nil {
				http.Error(w, "internal_error", http.StatusInternalServerError)
				return
			}

			if !ok {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error": "idempotency_key_in_flight"}`))
				return
			}

			// 3. Execute handler and record response
			rec := &responseRecorder{
				ResponseWriter: w,
				body:           new(bytes.Buffer),
				statusCode:     http.StatusOK, // Default success
			}

			next.ServeHTTP(rec, r)

			// 4. Cache only successful/client error responses (exclude 5xx)
			if rec.statusCode < http.StatusInternalServerError {
				_ = store.SetResponse(r.Context(), redisKey, CachedResponse{
					StatusCode: rec.statusCode,
					Body:       rec.body.Bytes(),
				}, idempotencyTTL)
			}
		})
	}
}
