package middleware

import (
	"context"
	"net/http"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/pkg/paseto"
)

// SessionAuth is the middleware that validates the PASETO JWT from a cookie
// and adds the user_id and tenant_id to the context.
// If the token is missing or invalid, it redirects to the login page.
func SessionAuth(parseToken func(token string) (*paseto.Claims, error), loginURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("moolah_token")
			if err != nil {
				// No cookie, redirect to login
				RedirectForClient(w, r, loginURL)
				return
			}

			claims, err := parseToken(cookie.Value)
			if err != nil {
				// Invalid token, clear cookie and redirect
				http.SetCookie(w, &http.Cookie{
					Name:   "moolah_token",
					MaxAge: -1,
					Path:   "/",
				})
				RedirectForClient(w, r, loginURL)
				return
			}

			// Add claims to context using the same keys as the API Auth middleware
			ctx := context.WithValue(r.Context(), middleware.TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, middleware.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, middleware.RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RedirectIfAuthenticated is a middleware that redirects the user if they already have a valid session.
// Useful for /login or /register pages.
func RedirectIfAuthenticated(parseToken func(token string) (*paseto.Claims, error), dashboardURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("moolah_token")
			if err == nil {
				_, err := parseToken(cookie.Value)
				if err == nil {
					// Valid session exists, redirect to dashboard
					RedirectForClient(w, r, dashboardURL)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RedirectForClient handles redirects for both standard and HTMX requests.
func RedirectForClient(w http.ResponseWriter, r *http.Request, targetURL string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", targetURL)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, targetURL, http.StatusSeeOther)
}
