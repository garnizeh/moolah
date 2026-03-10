package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondJSON_Error(t *testing.T) {
	t.Parallel()
	// A function cannot be serialized to JSON, so it will trigger an error in json.NewEncoder(w).Encode(data)
	invalidData := func() {}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Since we can't easily capture the slog output from the helper without complex wiring,
	// we just ensure the function executes and doesn't panic.
	// The status code is still set before the encode error.
	respondJSON(rr, req, invalidData, http.StatusOK)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRespondError(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	msg := "test error message"
	status := http.StatusBadRequest

	respondError(rr, req, msg, status)

	assert.Equal(t, status, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, msg, resp["error"])
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err            error
		name           string
		expectedStatus int
	}{
		{domain.ErrNotFound, "NotFound", http.StatusNotFound},
		{domain.ErrForbidden, "Forbidden", http.StatusForbidden},
		{domain.ErrConflict, "Conflict", http.StatusConflict},
		{domain.ErrInvalidInput, "InvalidInput", http.StatusUnprocessableEntity},
		{domain.ErrInvalidOTP, "InvalidOTP", http.StatusUnauthorized},
		{domain.ErrOTPRateLimited, "OTPRateLimited", http.StatusTooManyRequests},
		{domain.ErrTokenExpired, "TokenExpired", http.StatusUnauthorized},
		{domain.ErrUnauthorized, "Unauthorized", http.StatusUnauthorized},
		{errors.New("something else"), "Internal", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			handleError(rr, req, tt.err, "test message")

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
