package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/handler"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	// Initialize mocks
	authSvc := new(mocks.AuthService)
	// Other services can be nil as we are just checking routing table presence

	// Middleware dependencies
	tokenParser := func(token string) (*paseto.Claims, error) { return nil, nil }
	rateLimiterStore := middleware.NewRateLimiterStore()
	idempotencyStore := new(mocks.IdempotencyStore)

	// Create Server
	s := &Server{
		authHandler:      handler.NewAuthHandler(authSvc),
		tokenParser:      tokenParser,
		rateLimiterStore: rateLimiterStore,
		idempotencyStore: idempotencyStore,
	}

	mux := s.routes()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Healthz",
			method:         http.MethodGet,
			path:           "/healthz",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Auth Request OTP",
			method:         http.MethodPost,
			path:           "/v1/auth/otp/request",
			expectedStatus: http.StatusBadRequest, // Bad request because body is empty, but route matches
		},
		{
			name:           "Auth Verify OTP",
			method:         http.MethodPost,
			path:           "/v1/auth/otp/verify",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Auth Refresh Token",
			method:         http.MethodPost,
			path:           "/v1/auth/token/refresh",
			expectedStatus: http.StatusUnauthorized, // Middleware triggers before handler
		},
		{
			name:           "NotFound",
			method:         http.MethodGet,
			path:           "/v1/non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	s := &Server{}

	t.Run("Valid GET", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()

		s.handleHealthz(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "OK")
	})

	t.Run("Invalid Methods", func(t *testing.T) {
		t.Parallel()

		methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/healthz", nil)
			rr := httptest.NewRecorder()

			s.handleHealthz(rr, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		}
	})
}
