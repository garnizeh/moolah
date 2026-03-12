package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTPRateLimiter(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	store := NewRateLimiterStore()

	// Create a handler that just returns 200 OK
	handler := store.OTPRateLimiter()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("under limit - same email", func(t *testing.T) {
		t.Parallel()
		email := "test@example.com"
		body := map[string]string{"email": email}
		b, err := json.Marshal(body)
		require.NoError(t, err)

		for range otpRateLimit {
			req := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader(b))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}
	})

	t.Run("over limit - same email", func(t *testing.T) {
		t.Parallel()
		email := "limit@example.com"
		body := map[string]string{"email": email}
		b, err := json.Marshal(body)
		require.NoError(t, err)

		// Fill the bucket
		for range otpRateLimit {
			req := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader(b))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}

		// 6th request should fail
		req6 := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader(b))
		rr6 := httptest.NewRecorder()
		handler.ServeHTTP(rr6, req6)

		assert.Equal(t, http.StatusTooManyRequests, rr6.Code)
		assert.Contains(t, rr6.Body.String(), "RATE_LIMITED")

		retryAfter := rr6.Header().Get("Retry-After")
		assert.NotEmpty(t, retryAfter)
		seconds, err := strconv.Atoi(retryAfter)
		require.NoError(t, err)
		assert.Positive(t, seconds)
	})

	t.Run("different emails are independent", func(t *testing.T) {
		t.Parallel()
		email1 := "user1@example.com"
		email2 := "user2@example.com"

		b1, err := json.Marshal(map[string]string{"email": email1})
		require.NoError(t, err)
		b2, err := json.Marshal(map[string]string{"email": email2})
		require.NoError(t, err)

		// Exhaust user1
		for range otpRateLimit {
			req := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader(b1))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}

		// User1 should be limited
		req1 := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader(b1))
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusTooManyRequests, rr1.Code)

		// User2 should be OK
		req2 := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader(b2))
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusOK, rr2.Code)
	})

	t.Run("ignores non-POST requests", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/auth/otp/request", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ignores invalid bodies", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/auth/otp/request", bytes.NewReader([]byte("not-json")))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code) // Passes through to next handler
	})
}

func TestRateLimiterStore_Cleanup(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	// Use a very short interval for testing
	interval := 50 * time.Millisecond
	store := NewRateLimiterStoreWithInterval(interval)

	email := "stale@example.com"
	store.mu.Lock()
	store.limiters[email] = &emailLimiter{
		lastSeen: time.Now().Add(-2 * interval), // Already stale
	}
	store.mu.Unlock()

	// Wait for cleanup to run (at least 2 ticks to be safe)
	time.Sleep(3 * interval)

	store.mu.RLock()
	_, exists := store.limiters[email]
	store.mu.RUnlock()

	assert.False(t, exists, "stale email should have been cleaned up")
}

// Test that Close stops the cleanup goroutine promptly and does not prevent the
// rate-limiter middleware from functioning afterwards.
func TestRateLimiterStore_Close(t *testing.T) {
	// Create a store with a fast cleanup interval for the test.
	store := NewRateLimiterStoreWithInterval(10 * time.Millisecond)

	// Build a simple handler chain using the OTPRateLimiter middleware.
	h := store.OTPRateLimiter()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Prepare a POST request with an email payload to exercise the limiter.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"test@example.com"}`))
	req.Header.Set("Content-Type", "application/json")

	// First call should succeed (200).
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// Call Close in a goroutine and ensure it returns quickly (does not block).
	done := make(chan struct{})
	go func() {
		store.Close()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Close() blocked or took too long")
	}

	// After Close, middleware should still be callable (cleanup goroutine stopped).
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
