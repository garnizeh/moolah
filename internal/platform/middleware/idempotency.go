package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
)

const (
	idempotencyHeader = "Idempotency-Key"
	idempotencyTTL    = 24 * time.Hour
)

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
func Idempotency(store domain.IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			clientKey := r.Header.Get(idempotencyHeader)
			if clientKey == "" {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(map[string]string{"error": "missing_idempotency_key"})
				if err != nil {
					http.Error(w, "failed to encode error response", http.StatusInternalServerError)
				}
				return
			}

			// Validate key length to prevent abuse
			if len(clientKey) > 128 {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid_idempotency_key"})
				if err != nil {
					http.Error(w, "failed to encode error response", http.StatusInternalServerError)
				}
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
				_, err = w.Write(cached.Body)
				if err != nil {
					// We can't use http.Error here because we already sent w.WriteHeader
					// but we also can't just ignore it.
					return
				}
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
				err := json.NewEncoder(w).Encode(map[string]string{"error": "idempotency_key_in_flight"})
				if err != nil {
					http.Error(w, "failed to encode error response", http.StatusInternalServerError)
				}
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
				err := store.SetResponse(r.Context(), redisKey, domain.CachedResponse{
					StatusCode: rec.statusCode,
					Body:       rec.body.Bytes(),
				}, idempotencyTTL)
				if err != nil {
					// We log it and continue. In production you'd use a real logger.
					fmt.Printf("idempotency: failed to set response: %v\n", err)
				}
			}
		})
	}
}
