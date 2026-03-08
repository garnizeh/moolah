package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
)

const (
	idempotencyHeader = "Idempotency-Key"
	idempotencyTTL    = 24 * time.Hour
)

// responseRecorder captures the response to be cached for future idempotent requests.
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

// Idempotency middleware ensures that POST requests are processed at most once
// by caching the response associated with a specific Idempotency-Key.
func Idempotency(store domain.IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply idempotency logic to POST requests (standard practice for mutations).
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			// Extract and validate Idempotency-Key from headers.
			clientKey := r.Header.Get(idempotencyHeader)
			if clientKey == "" {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(map[string]string{"error": "missing_idempotency_key"})
				if err != nil {
					slog.ErrorContext(ctx, "idempotency: failed to write missing key error response", "error", err)
				}
				return
			}

			// Validate key length to prevent ReDoS or memory abuse.
			if len(clientKey) > 128 {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid_idempotency_key"})
				if err != nil {
					slog.ErrorContext(ctx, "idempotency: failed to write invalid key error response", "error", err)
				}
				return
			}

			// Extract userID from context (provided by Auth middleware).
			userID, ok := UserIDFromCtx(r.Context())
			if !ok {
				slog.WarnContext(ctx, "idempotency: failed to extract user ID from context, defaulting to anonymous")
			}
			if userID == "" {
				userID = "anonymous"
			}

			// Composite key prevents collisions between different users using the same client key.
			redisKey := fmt.Sprintf("idempotency:%s:%s", userID, clientKey)

			// 1. Check for a previously cached response.
			cached, err := store.Get(r.Context(), redisKey)
			if err != nil {
				// #nosec G706: slog is a structured logger that escapes control characters, preventing log injection.
				slog.ErrorContext(ctx, "idempotency: store get error", "error", err, "key", redisKey)
				http.Error(w, "internal_error", http.StatusInternalServerError)
				return
			}

			if cached != nil {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(cached.StatusCode)

				// Use existing 'err' to avoid shadowing and satisfy govet.
				var n int
				/* #nosec G705 */
				n, err = w.Write(cached.Body)
				if err != nil {
					/* #nosec G706 */
					slog.ErrorContext(ctx, "idempotency: failed to write cached response body",
						"user_id", userID,
						"idempotency_key", clientKey,
						"bytes_written", n,
						"error", err,
					)
					return
				}
				return
			}

			// 2. Try to acquire an atomic lock to prevent "in-flight" race conditions.
			ok, err = store.SetLocked(r.Context(), redisKey, idempotencyTTL)
			if err != nil {
				// #nosec G706: redisKey is safe here because slog handles escaping of user-provided data.
				slog.ErrorContext(ctx, "idempotency: lock acquisition error", "error", err, "key", redisKey)
				http.Error(w, "internal_error", http.StatusInternalServerError)
				return
			}

			if !ok {
				w.WriteHeader(http.StatusConflict)
				eerr := json.NewEncoder(w).Encode(map[string]string{"error": "idempotency_key_in_flight"})
				if eerr != nil {
					// #nosec G706: redisKey is safe here because slog handles escaping of user-provided data.
					slog.ErrorContext(ctx, "idempotency: failed to write in-flight error response", "error", eerr, "key", redisKey)
				}
				return
			}

			// 3. Execute the next handler while recording the output.
			rec := &responseRecorder{
				ResponseWriter: w,
				body:           new(bytes.Buffer),
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			// 4. Cache only successful or client-side error responses (exclude 5xx server errors).
			if rec.statusCode < http.StatusInternalServerError {
				err = store.SetResponse(r.Context(), redisKey, domain.CachedResponse{
					StatusCode: rec.statusCode,
					Body:       rec.body.Bytes(),
				}, idempotencyTTL)
				if err != nil {
					// #nosec G706: Using structured logging to safely record the failure.
					slog.WarnContext(ctx, "idempotency: failed to cache response", "error", err, "key", redisKey)
				}
			}
		})
	}
}
