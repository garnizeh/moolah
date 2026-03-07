package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	t.Run("valid token adds claims to context", func(t *testing.T) {
		t.Parallel()

		tenantID := "tenant_123"
		userID := "user_456"
		role := domain.RoleMember

		mockParser := func(token string) (*paseto.Claims, error) {
			return &paseto.Claims{
				TenantID: tenantID,
				UserID:   userID,
				Role:     string(role),
			}, nil
		}

		mw := Auth(mockParser)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tID, okT := TenantIDFromCtx(r.Context())
			uID, okU := UserIDFromCtx(r.Context())
			uRole, okR := RoleFromCtx(r.Context())

			assert.True(t, okT)
			assert.True(t, okU)
			assert.True(t, okR)
			assert.Equal(t, tenantID, tID)
			assert.Equal(t, userID, uID)
			assert.Equal(t, string(role), uRole)

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("missing authorization header returns unauthorized", func(t *testing.T) {
		t.Parallel()

		mw := Auth(nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var resp ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "missing_token", resp.Error.Code)
	})

	t.Run("invalid authorization format returns unauthorized", func(t *testing.T) {
		t.Parallel()

		mw := Auth(nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var resp ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid_token_format", resp.Error.Code)
	})

	t.Run("invalid token returns unauthorized", func(t *testing.T) {
		t.Parallel()

		mockParser := func(token string) (*paseto.Claims, error) {
			return nil, errors.New("invalid token")
		}

		mw := Auth(mockParser)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var resp ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid_token", resp.Error.Code)
	})

	t.Run("RequireAuth is an alias for Auth", func(t *testing.T) {
		t.Parallel()
		// Just verify it compiles and returns a function
		mw := RequireAuth(nil)
		assert.NotNil(t, mw)
	})
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userRole       domain.Role
		requiredRoles  []domain.Role
		expectedStatus int
	}{
		{
			name:           "user has required role",
			userRole:       domain.RoleAdmin,
			requiredRoles:  []domain.Role{domain.RoleAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user has one of required roles",
			userRole:       domain.RoleMember,
			requiredRoles:  []domain.Role{domain.RoleAdmin, domain.RoleMember},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user does not have required role",
			userRole:       domain.RoleMember,
			requiredRoles:  []domain.Role{domain.RoleAdmin},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing role in context returns unauthorized",
			userRole:       "",
			requiredRoles:  []domain.Role{domain.RoleAdmin},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mw := RequireRole(tt.requiredRoles...)
			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.userRole != "" {
				ctx := context.WithValue(req.Context(), RoleKey, string(tt.userRole))
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestContextHelpers(t *testing.T) {
	t.Parallel()

	t.Run("TenantIDFromCtx", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		val, ok := TenantIDFromCtx(ctx)
		assert.False(t, ok)
		assert.Empty(t, val)

		ctx = context.WithValue(ctx, TenantIDKey, "test")
		val, ok = TenantIDFromCtx(ctx)
		assert.True(t, ok)
		assert.Equal(t, "test", val)
	})

	t.Run("UserIDFromCtx", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		val, ok := UserIDFromCtx(ctx)
		assert.False(t, ok)
		assert.Empty(t, val)

		ctx = context.WithValue(ctx, UserIDKey, "test")
		val, ok = UserIDFromCtx(ctx)
		assert.True(t, ok)
		assert.Equal(t, "test", val)
	})
}
