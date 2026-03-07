package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	otpRateLimit  = 5
	otpRatePeriod = 15 * time.Minute
)

type emailLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterStore holds the state for the in-memory rate limiters.
type RateLimiterStore struct {
	log      *slog.Logger
	limiters map[string]*emailLimiter
	mu       sync.RWMutex
}

// NewRateLimiterStore creates a new store and starts the cleanup goroutine.
func NewRateLimiterStore(log *slog.Logger) *RateLimiterStore {
	return NewRateLimiterStoreWithInterval(log, otpRatePeriod)
}

// NewRateLimiterStoreWithInterval is for testing cleanup with custom intervals.
func NewRateLimiterStoreWithInterval(log *slog.Logger, interval time.Duration) *RateLimiterStore {
	store := &RateLimiterStore{
		limiters: make(map[string]*emailLimiter),
		log:      log,
	}

	go store.cleanup(interval)

	return store
}

func (s *RateLimiterStore) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for email, l := range s.limiters {
			if time.Since(l.lastSeen) > interval {
				delete(s.limiters, email)
			}
		}
		s.mu.Unlock()
	}
}

// OTPRateLimiter returns a middleware that enforces per-email OTP rate limiting.
func (s *RateLimiterStore) OTPRateLimiter() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Read and buffer the body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				s.log.Error("failed to read request body in rate limiter", "error", err)
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Try to extract email from JSON body
			var payload struct {
				Email string `json:"email"`
			}
			if err := json.Unmarshal(body, &payload); err != nil || payload.Email == "" {
				// If we can't get the email, let the actual handler deal with the validation
				next.ServeHTTP(w, r)
				return
			}

			s.mu.Lock()
			l, exists := s.limiters[payload.Email]
			if !exists {
				// Tokens are added at a rate of 5 per 15 mins (1 token per 3 mins).
				rLimit := rate.Every(otpRatePeriod / otpRateLimit)
				l = &emailLimiter{
					limiter: rate.NewLimiter(rLimit, otpRateLimit),
				}
				s.limiters[payload.Email] = l
			}
			l.lastSeen = time.Now()

			if !l.limiter.Allow() {
				// Before unlocking, we'll try to calculate a delay.
				// For fixed windows, this is harder to calculate accurately with token-bucket,
				// but we'll provide a reasonable guess based on the next token arrival.
				var retryAfter string
				res := l.limiter.Reserve()
				if !res.OK() {
					retryAfter = "900" // Fallback to 15 mins
				} else {
					delay := res.DelayFrom(time.Now())
					res.Cancel() // We don't want to consume this token
					// Give him a few extra seconds to ensure token arrival
					retryAfter = strconv.FormatInt(int64(delay.Seconds()+1), 10)
				}
				s.mu.Unlock()

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", retryAfter)
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "RATE_LIMITED",
						"message": "Too many OTP requests. Please wait before retrying.",
					},
				})
				return
			}
			s.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
