package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_RequestOTP(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "test@example.com"
		service.On("RequestOTP", mock.Anything, email).Return(nil)

		reqBody, err := json.Marshal(RequestOTPRequest{Email: email})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.RequestOTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
		service.AssertExpectations(t)
	})

	t.Run("invalid_email", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		reqBody, err := json.Marshal(RequestOTPRequest{Email: "invalid-email"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.RequestOTP(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("rate_limited", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "test@example.com"
		service.On("RequestOTP", mock.Anything, email).Return(domain.ErrOTPRateLimited)

		reqBody, err := json.Marshal(RequestOTPRequest{Email: email})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.RequestOTP(rr, req)

		require.Equal(t, http.StatusTooManyRequests, rr.Code)
	})

	t.Run("user_not_found_returns_accepted", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "unknown@example.com"
		service.On("RequestOTP", mock.Anything, email).Return(domain.ErrNotFound)

		reqBody, err := json.Marshal(RequestOTPRequest{Email: email})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.RequestOTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
	})

	t.Run("invalid_request_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()

		h.RequestOTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "test@example.com"
		service.On("RequestOTP", mock.Anything, email).Return(errors.New("db error"))

		reqBody, err := json.Marshal(RequestOTPRequest{Email: email})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/request", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.RequestOTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestAuthHandler_VerifyOTP(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "test@example.com"
		code := "123456"
		expectedPair := &domain.TokenPair{
			AccessToken:  "access",
			RefreshToken: "refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
		}

		service.On("VerifyOTP", mock.Anything, email, code).Return(expectedPair, nil)

		reqBody, err := json.Marshal(VerifyOTPRequest{Email: email, Code: code})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var resp TokenResponse
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, expectedPair.AccessToken, resp.AccessToken)
		require.Equal(t, expectedPair.RefreshToken, resp.RefreshToken)
		service.AssertExpectations(t)
	})

	t.Run("invalid_otp", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "test@example.com"
		code := "000000"
		service.On("VerifyOTP", mock.Anything, email, code).Return(nil, domain.ErrInvalidOTP)

		reqBody, err := json.Marshal(VerifyOTPRequest{Email: email, Code: code})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid_request_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("user_not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "unknown@example.com"
		service.On("VerifyOTP", mock.Anything, email, "123456").Return(nil, domain.ErrNotFound)

		reqBody, err := json.Marshal(VerifyOTPRequest{Email: email, Code: "123456"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		email := "test@example.com"
		service.On("VerifyOTP", mock.Anything, email, "123456").Return(nil, errors.New("boom"))

		reqBody, err := json.Marshal(VerifyOTPRequest{Email: email, Code: "123456"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("invalid_email", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		reqBody, err := json.Marshal(VerifyOTPRequest{Email: "invalid-email", Code: "123456"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("invalid_code", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		reqBody, err := json.Marshal(VerifyOTPRequest{Email: "test@example.com", Code: "123"})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/otp/verify", bytes.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.VerifyOTP(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		refreshToken := "old-refresh-token"
		expectedPair := &domain.TokenPair{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
		}

		service.On("RefreshToken", mock.Anything, refreshToken).Return(expectedPair, nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+refreshToken)
		rr := httptest.NewRecorder()

		h.RefreshToken(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var resp TokenResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, expectedPair.AccessToken, resp.AccessToken)
		service.AssertExpectations(t)
	})

	t.Run("missing_header", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/refresh", nil)
		rr := httptest.NewRecorder()

		h.RefreshToken(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("expired_token", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		refreshToken := "expired-token"
		service.On("RefreshToken", mock.Anything, refreshToken).Return(nil, domain.ErrTokenExpired)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+refreshToken)
		rr := httptest.NewRecorder()

		h.RefreshToken(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("user_not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		refreshToken := "valid-but-no-user"
		service.On("RefreshToken", mock.Anything, refreshToken).Return(nil, domain.ErrNotFound)

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+refreshToken)
		rr := httptest.NewRecorder()

		h.RefreshToken(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AuthService)
		h := NewAuthHandler(service)

		refreshToken := "valid-token"
		service.On("RefreshToken", mock.Anything, refreshToken).Return(nil, errors.New("boom"))

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+refreshToken)
		rr := httptest.NewRecorder()

		h.RefreshToken(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
