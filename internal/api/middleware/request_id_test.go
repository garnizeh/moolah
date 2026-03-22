package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	// Simple handler to check context
	var capturedID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	middlewareHandler := RequestID(nextHandler)

	// Create request and recorder
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Serve
	middlewareHandler.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check header
	respID := rec.Header().Get("X-Request-ID")
	assert.NotEmpty(t, respID, "X-Request-ID header should not be empty")

	// Check context capture
	assert.Equal(t, respID, capturedID, "ID in context should match ID in header")
}

func TestFromContext_Empty(t *testing.T) {
	id := FromContext(httptest.NewRequest("GET", "/", nil).Context())
	assert.Equal(t, "unknown", id)
}

func TestRequestID_Uniqueness(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	middlewareHandler := RequestID(nextHandler)

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		middlewareHandler.ServeHTTP(rec, req)

		id := rec.Header().Get("X-Request-ID")
		assert.False(t, ids[id], "Duplicate Request ID detected: %s", id)
		ids[id] = true
	}
}
