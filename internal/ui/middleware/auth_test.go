package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/stretchr/testify/assert"
)

func TestSessionAuth(t *testing.T) {
	t.Parallel()

	dummyParser := func(token string) (*paseto.Claims, error) {
		if token == "valid-token" {
			return &paseto.Claims{
				UserID:   "user-123",
				TenantID: "tenant-456",
				Role:     "admin",
			}, nil
		}
		return nil, errors.New("invalid token")
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey)
		tenantID := r.Context().Value(middleware.TenantIDKey)
		assert.Equal(t, "user-123", userID)
		assert.Equal(t, "tenant-456", tenantID)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("valid session", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "moolah_token", Value: "valid-token"})
		w := httptest.NewRecorder()

		handler := SessionAuth(dummyParser, "/login")(nextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing cookie redirects", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		w := httptest.NewRecorder()

		handler := SessionAuth(dummyParser, "/login")(nextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})

	t.Run("invalid token redirects and clears cookie", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req.AddCookie(&http.Cookie{Name: "moolah_token", Value: "bad-token"})
		w := httptest.NewRecorder()

		handler := SessionAuth(dummyParser, "/login")(nextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))

		// Check if cookie was cleared
		cookies := w.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "moolah_token" && c.MaxAge == -1 {
				found = true
			}
		}
		assert.True(t, found)
	})
}

func TestRedirectIfAuthenticated(t *testing.T) {
	t.Parallel()

	dummyParser := func(token string) (*paseto.Claims, error) {
		if token == "valid-token" {
			return &paseto.Claims{}, nil
		}
		return nil, errors.New("invalid token")
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("authenticated user redirects to dashboard", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		req.AddCookie(&http.Cookie{Name: "moolah_token", Value: "valid-token"})
		w := httptest.NewRecorder()

		handler := RedirectIfAuthenticated(dummyParser, "/dashboard")(nextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/dashboard", w.Header().Get("Location"))
	})

	t.Run("unauthenticated user proceeds to next", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		w := httptest.NewRecorder()

		handler := RedirectIfAuthenticated(dummyParser, "/dashboard")(nextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid token proceeds to next", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		req.AddCookie(&http.Cookie{Name: "moolah_token", Value: "bad-token"})
		w := httptest.NewRecorder()

		handler := RedirectIfAuthenticated(dummyParser, "/dashboard")(nextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
