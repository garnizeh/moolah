package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/ui/middleware"
)

const maxFormBodyBytes int64 = 1 << 20 // 1MB

type AuthHandler struct {
	authService domain.AuthService
	isDev       bool
}

func NewAuthHandler(
	authService domain.AuthService,
	isDev bool,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		isDev:       isDev,
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
	if !parseFormWithLimit(w, r, maxFormBodyBytes) {
		return
	}

	email := strings.TrimSpace(r.Form.Get("email"))
	if email == "" {
		props := OTPRequestProps{Error: "Email is required"}
		if err := OTPRequestForm(props).Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "failed to render otp request form", "error", err)
		}
		return
	}

	err := h.authService.RequestOTP(r.Context(), email)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to request otp", "error", err, "masked_email", maskEmail(email))
	}

	props := OTPVerifyProps{Email: email}
	if err := OTPVerifyForm(props).Render(r.Context(), w); err != nil {
		slog.ErrorContext(r.Context(), "failed to render otp verify form", "error", err)
	}
}

// VerifyOTP handles the form submission to verify the OTP and sign in.
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if !parseFormWithLimit(w, r, maxFormBodyBytes) {
		return
	}

	email := strings.TrimSpace(r.Form.Get("email"))
	code := strings.TrimSpace(r.Form.Get("code"))

	if email == "" || code == "" {
		props := OTPVerifyProps{Email: email, Error: "Email and code are required"}
		if err := OTPVerifyForm(props).Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "failed to render otp verify form", "error", err)
		}
		return
	}

	tokens, err := h.authService.VerifyOTP(r.Context(), email, code)
	if err != nil {
		slog.WarnContext(r.Context(), "failed to verify otp", "error", err, "masked_email", maskEmail(email))
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
		Secure:   !h.isDev,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(tokens.ExpiresAt).Seconds()),
	})

	// Redirect to dashboard (handles both HTMX and non-HTMX)
	middleware.RedirectForClient(w, r, "/dashboard")
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "moolah_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   !h.isDev,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	middleware.RedirectForClient(w, r, "/web/login")
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "****"
	}
	user := parts[0]
	domain := parts[1]

	if len(user) <= 2 {
		return user + "*****@" + domain
	}
	return user[:2] + "*****@" + domain
}

func parseFormWithLimit(w http.ResponseWriter, r *http.Request, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseForm(); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "invalid form payload", http.StatusBadRequest)
		}
		return false
	}
	return true
}
