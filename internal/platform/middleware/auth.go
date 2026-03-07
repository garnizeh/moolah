package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/paseto"
)

type contextKey int

const (
	tenantIDKey contextKey = iota
	userIDKey
	roleKey
)

// ErrorResponse represents the JSON body returned for authentication and authorization errors.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// TokenParser is a function type that parses a token string into Claims.
type TokenParser func(string) (*paseto.Claims, error)

// RequireAuth validates the Bearer token and injects claims into context.
func RequireAuth(parse TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				sendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "malformed bearer token")
				return
			}

			token := parts[1]
			claims, err := parse(token)
			if err != nil {
				if errors.Is(err, paseto.ErrTokenExpired) {
					sendError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "token has expired")
					return
				}
				sendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or malformed token")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, domain.Role(claims.Role))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces that the authenticated user has at least the specified role level.
// Chain this AFTER RequireAuth.
func RequireRole(requiredRole domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := RoleFromCtx(r.Context())
			if !ok {
				// This implies RequireAuth was not used before RequireRole
				sendError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			if !userRole.CanAccess(requiredRole) {
				sendError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantIDFromCtx retrieves the tenant ID from the context.
func TenantIDFromCtx(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(tenantIDKey).(string)
	return val, ok
}

// UserIDFromCtx retrieves the user ID from the context.
func UserIDFromCtx(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userIDKey).(string)
	return val, ok
}

// RoleFromCtx retrieves the user role from the context.
func RoleFromCtx(ctx context.Context) (domain.Role, bool) {
	val, ok := ctx.Value(roleKey).(domain.Role)
	return val, ok
}

func sendError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message

	// Ignore error from json.NewEncoder as it is unlikely to fail for this simple struct
	_ = json.NewEncoder(w).Encode(resp)
}
