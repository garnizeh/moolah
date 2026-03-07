package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/paseto"
)

type contextKey int

const (
	TenantIDKey contextKey = iota
	UserIDKey
	RoleKey
	// Deprecated: use TenantIDKey instead. Keeping for backward compatibility
	// until all tests are updated.
	tenantIDKey = TenantIDKey
	// Deprecated: use UserIDKey instead.
	userIDKey = UserIDKey
)

// TenantIDFromCtx extracts the tenant ID from the context.
func TenantIDFromCtx(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	return tenantID, ok
}

// UserIDFromCtx extracts the user ID from the context.
func UserIDFromCtx(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// RoleFromCtx extracts the user role from the context.
func RoleFromCtx(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

// ErrorResponse represents the JSON body returned for authentication and authorization errors.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// TokenParser is a function type that parses a token string into Claims.
type TokenParser func(token string) (*paseto.Claims, error)

// Auth is the middleware that validates the JWT and adds the user_id and tenant_id to the context.
func Auth(parseToken TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				unauthorized(w, "missing_token", "Authorization header is required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				unauthorized(w, "invalid_token_format", "Authorization header must be in the format 'Bearer <token>'")
				return
			}

			claims, err := parseToken(parts[1])
			if err != nil {
				unauthorized(w, "invalid_token", "The token provided is invalid or expired")
				return
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth is an alias for Auth to satisfy existing calls.
func RequireAuth(parseToken TokenParser) func(http.Handler) http.Handler {
	return Auth(parseToken)
}

// RequireRole is a middleware that ensures the user has one of the required roles.
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				unauthorized(w, "missing_role", "User role not found in context")
				return
			}

			allowed := false
			for _, role := range roles {
				if string(role) == userRole {
					allowed = true
					break
				}
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(ErrorResponse{
					Error: struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					}{
						Code:    "forbidden",
						Message: "You don't have permission to perform this action",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func unauthorized(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}
