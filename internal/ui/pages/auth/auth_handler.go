package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/paseto"
)

type AuthHandler struct {
	authService domain.AuthService
	tokenParser func(token string) (*paseto.Claims, error)
	isProd      bool
}

func NewAuthHandler(
	authService domain.AuthService,
	tokenParser func(token string) (*paseto.Claims, error),
	isProd bool,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		tokenParser: tokenParser,
		isProd:      isProd,
	}
}

// Login renders the OTP request page.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	props := OTPRequestProps{}
	if err := OTPRequest(props).Render(r.Context(), w); err != nil {
		slog.ErrorContext(r.Context(), "failed to render login page", "error", err)
	}
}

// RequestOTP handles the form submission to request a new OTP.
func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit for form submissions
	email := r.FormValue("email")
	if email == "" {
		props := OTPRequestProps{Error: "Email is required"}
		if err := OTPRequestForm(props).Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "failed to render otp request form", "error", err)
		}
		return
	}

	err := h.authService.RequestOTP(r.Context(), email)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to request otp", "error", err, "email", email)
	}

	props := OTPVerifyProps{Email: email}
	if err := OTPVerifyForm(props).Render(r.Context(), w); err != nil {
		slog.ErrorContext(r.Context(), "failed to render otp verify form", "error", err)
	}
}

// VerifyOTP handles the form submission to verify the OTP and sign in.
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit for form submissions
	email := r.FormValue("email")
	code := r.FormValue("code")

	if email == "" || code == "" {
		props := OTPVerifyProps{Email: email, Error: "Code is required"}
		if err := OTPVerifyForm(props).Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "failed to render otp verify form", "error", err)
		}
		return
	}

	tokens, err := h.authService.VerifyOTP(r.Context(), email, code)
	if err != nil {
		slog.WarnContext(r.Context(), "failed to verify otp", "error", err, "email", email)
		props := OTPVerifyProps{Email: email, Error: "Invalid or expired code"}
		if err := OTPVerifyForm(props).Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "failed to render otp verify form", "error", err)
		}
		return
	}

	// Set HttpOnly cookie with the access token
	http.SetCookie(w, &http.Cookie{
		Name:     "moolah_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(tokens.ExpiresAt).Seconds()),
	})

	// HTMX redirect to dashboard
	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "moolah_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	w.Header().Set("HX-Redirect", "/web/login")
	w.WriteHeader(http.StatusOK)
}
