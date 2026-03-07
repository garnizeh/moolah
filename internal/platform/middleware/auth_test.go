package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuth(t *testing.T) {
	t.Parallel()

	validClaims := &paseto.Claims{
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
		TenantID:  "tenant_123",
		UserID:    "user_456",
		Role:      string(domain.RoleAdmin),
	}

	mockParser := func(claims *paseto.Claims, err error) TokenParser {
		return func(token string) (*paseto.Claims, error) {
			return claims, err
		}
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := TenantIDFromCtx(r.Context())
		userID, _ := UserIDFromCtx(r.Context())
		role, _ := RoleFromCtx(r.Context())

		assert.Equal(t, validClaims.TenantID, tenantID)
		assert.Equal(t, validClaims.UserID, userID)
		assert.Equal(t, domain.Role(validClaims.Role), role)

		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		parser         TokenParser
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer valid_token",
			parser:         mockParser(validClaims, nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing header",
			authHeader:     "",
			parser:         mockParser(nil, nil),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "Malformed header - no bearer",
			authHeader:     "invalid_token",
			parser:         mockParser(nil, nil),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
		{
			name:           "Expired token",
			authHeader:     "Bearer expired_token",
			parser:         mockParser(nil, paseto.ErrTokenExpired),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_EXPIRED",
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid_token",
			parser:         mockParser(nil, paseto.ErrTokenInvalid),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "UNAUTHORIZED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			handler := RequireAuth(tt.parser)(finalHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedCode != "" {
				var resp ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCode, resp.Error.Code)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		userRole       domain.Role
		requiredRole   domain.Role
		expectedStatus int
	}{
		{
			name:           "Admin accessing member endpoint",
			userRole:       domain.RoleAdmin,
			requiredRole:   domain.RoleMember,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Member accessing member endpoint",
			userRole:       domain.RoleMember,
			requiredRole:   domain.RoleMember,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Member accessing admin endpoint",
			userRole:       domain.RoleMember,
			requiredRole:   domain.RoleAdmin,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Sysadmin accessing admin endpoint",
			userRole:       domain.RoleSysadmin,
			requiredRole:   domain.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			// Mock claims in context
			claims := &paseto.Claims{
				TenantID: "tenant_123",
				UserID:   "user_456",
				Role:     string(tt.userRole),
			}
			mockParser := func(token string) (*paseto.Claims, error) {
				return claims, nil
			}

			rr := httptest.NewRecorder()

			// Chain RequireAuth and RequireRole
			handler := RequireAuth(mockParser)(RequireRole(tt.requiredRole)(finalHandler))
			req.Header.Set("Authorization", "Bearer valid_token")
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestContextHelpers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("TenantIDFromCtx", func(t *testing.T) {
		t.Parallel()
		val, ok := TenantIDFromCtx(ctx)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("UserIDFromCtx", func(t *testing.T) {
		t.Parallel()
		val, ok := UserIDFromCtx(ctx)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("RoleFromCtx", func(t *testing.T) {
		t.Parallel()
		val, ok := RoleFromCtx(ctx)
		assert.False(t, ok)
		assert.Empty(t, string(val))
	})
}
