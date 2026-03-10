package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/go-playground/validator/v10"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	service  domain.AuthService
	validate *validator.Validate
}

// NewAuthHandler creates a new AuthHandler with the given service and logger.
func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{
		service:  service,
		validate: validator.New(),
	}
}

// RequestOTPRequest defines the payload for requesting an OTP.
type RequestOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyOTPRequest defines the payload for verifying an OTP.
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code"  validate:"required,len=6"`
}

// TokenResponse defines the successful authentication response.
type TokenResponse struct {
	// ExpiresAt is the timestamp when the access token expires.
	ExpiresAt time.Time `json:"expires_at"`
	// AccessToken is the PASETO access token.
	//
	//nolint:gosec
	AccessToken string `json:"access_token"`
	// RefreshToken is the PASETO refresh token.
	//
	//nolint:gosec
	RefreshToken string `json:"refresh_token"`
}

// RequestOTP handles POST /v1/auth/otp/request
//
// @Summary		Request OTP
// @Description	Validates email and generates a 6-digit verification code sent via email.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body	RequestOTPRequest	true	"Email address"
// @Success		202		"OTP requested successfully (Accepted)"
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		429		{object}	map[string]string	"Rate limit exceeded"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/auth/otp/request [post]
func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.service.RequestOTP(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOTPRateLimited):
			respondError(w, r, err.Error(), http.StatusTooManyRequests)
		case errors.Is(err, domain.ErrNotFound):
			// We return 202 even if the user is not found to prevent user enumeration
			w.WriteHeader(http.StatusAccepted)
		default:
			slog.ErrorContext(r.Context(), "failed to request OTP", "error", err, "email", req.Email)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// VerifyOTP handles POST /v1/auth/otp/verify
//
// @Summary		Verify OTP
// @Description	Validates the 6-digit code and returns a PASETO token pair.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		VerifyOTPRequest	true	"Email and Code"
// @Success		200		{object}	TokenResponse
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Invalid or expired OTP"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/auth/otp/verify [post]
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	pair, err := h.service.VerifyOTP(r.Context(), req.Email, req.Code)
	if err != nil {
		handleError(w, r, err, "failed to verify OTP")
		return
	}

	respondJSON(w, r, TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, http.StatusOK)
}

// RefreshToken handles POST /v1/auth/token/refresh
//
// @Summary		Refresh Token
// @Description	Uses a valid refresh token to obtain a new access token.
// @Tags			auth
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	TokenResponse
// @Failure		401	{object}	map[string]string	"Invalid or expired refresh token"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/auth/token/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		respondError(w, r, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}
	token := authHeader[7:]

	pair, err := h.service.RefreshToken(r.Context(), token)
	if err != nil {
		handleError(w, r, err, "failed to refresh token")
		return
	}

	respondJSON(w, r, TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, http.StatusOK)
}
