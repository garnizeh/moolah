package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_RequestOTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		payload        any
		setupMock      func(m *mocks.AuthService)
		expectedStatus int
	}{
		{
			name: "Success",
			payload: RequestOTPRequest{
				Email: "test@example.com",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("RequestOTP", mock.Anything, "test@example.com").Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "Invalid Email",
			payload: RequestOTPRequest{
				Email: "invalid-email",
			},
			setupMock:      func(m *mocks.AuthService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Invalid Body",
			payload:        "invalid",
			setupMock:      func(m *mocks.AuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Rate Limited",
			payload: RequestOTPRequest{
				Email: "test@example.com",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("RequestOTP", mock.Anything, "test@example.com").Return(domain.ErrOTPRateLimited)
			},
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name: "User Not Found Returns Accepted",
			payload: RequestOTPRequest{
				Email: "notfound@example.com",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("RequestOTP", mock.Anything, "notfound@example.com").Return(domain.ErrNotFound)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "Internal Error",
			payload: RequestOTPRequest{
				Email: "error@example.com",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("RequestOTP", mock.Anything, "error@example.com").Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := new(mocks.AuthService)
			tt.setupMock(svc)
			h := NewAuthHandler(svc, slog.Default())

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			h.RequestOTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	t.Parallel()

	pair := &domain.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	tests := []struct {
		name           string
		payload        any
		setupMock      func(m *mocks.AuthService)
		expectedStatus int
	}{
		{
			name: "Success",
			payload: VerifyOTPRequest{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("VerifyOTP", mock.Anything, "test@example.com", "123456").Return(pair, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Payload",
			payload: VerifyOTPRequest{
				Email: "test@example.com",
				Code:  "123", // too short
			},
			setupMock:      func(m *mocks.AuthService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "Invalid OTP",
			payload: VerifyOTPRequest{
				Email: "test@example.com",
				Code:  "123456",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("VerifyOTP", mock.Anything, "test@example.com", "123456").Return(nil, domain.ErrInvalidOTP)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "User Not Found",
			payload: VerifyOTPRequest{
				Email: "notfound@example.com",
				Code:  "123456",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("VerifyOTP", mock.Anything, "notfound@example.com", "123456").Return((*domain.TokenPair)(nil), domain.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Internal Error",
			payload: VerifyOTPRequest{
				Email: "error@example.com",
				Code:  "123456",
			},
			setupMock: func(m *mocks.AuthService) {
				m.On("VerifyOTP", mock.Anything, "error@example.com", "123456").Return((*domain.TokenPair)(nil), errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := new(mocks.AuthService)
			tt.setupMock(svc)
			h := NewAuthHandler(svc, slog.Default())

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			h.VerifyOTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp TokenResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, pair.AccessToken, resp.AccessToken)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Parallel()

	pair := &domain.TokenPair{
		AccessToken:  "new-access",
		RefreshToken: "new-refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(m *mocks.AuthService)
		expectedStatus int
	}{
		{
			name:       "Success",
			authHeader: "Bearer valid-refresh",
			setupMock: func(m *mocks.AuthService) {
				m.On("RefreshToken", mock.Anything, "valid-refresh").Return(pair, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Header",
			authHeader:     "",
			setupMock:      func(m *mocks.AuthService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Header Format",
			authHeader:     "Basic 123",
			setupMock:      func(m *mocks.AuthService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Expired Token",
			authHeader: "Bearer expired",
			setupMock: func(m *mocks.AuthService) {
				m.On("RefreshToken", mock.Anything, "expired").Return((*domain.TokenPair)(nil), domain.ErrTokenExpired)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "User Not Found",
			authHeader: "Bearer ghost-token",
			setupMock: func(m *mocks.AuthService) {
				m.On("RefreshToken", mock.Anything, "ghost-token").Return((*domain.TokenPair)(nil), domain.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "Internal Error",
			authHeader: "Bearer error",
			setupMock: func(m *mocks.AuthService) {
				m.On("RefreshToken", mock.Anything, "error").Return((*domain.TokenPair)(nil), errors.New("unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := new(mocks.AuthService)
			tt.setupMock(svc)
			h := NewAuthHandler(svc, slog.Default())

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/refresh", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			h.RefreshToken(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			svc.AssertExpectations(t)
		})
	}
}
