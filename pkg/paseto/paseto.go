package paseto

import (
	"errors"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
)

var (
	// ErrTokenExpired is returned when a valid token is past its expiry time.
	ErrTokenExpired = errors.New("paseto: token expired")

	// ErrTokenInvalid is returned when the token cannot be decrypted or parsed.
	ErrTokenInvalid = errors.New("paseto: token invalid")
)

// Claims holds the application-specific payload embedded in the token.
type Claims struct {
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
}

// Seal encrypts claims into a PASETO v4 local token string.
func Seal(claims Claims, key paseto.V4SymmetricKey) (string, error) {
	token := paseto.NewToken()

	err := token.Set("tenant_id", claims.TenantID)
	if err != nil {
		return "", fmt.Errorf("paseto: failed to set tenant_id: %w", err)
	}

	err = token.Set("user_id", claims.UserID)
	if err != nil {
		return "", fmt.Errorf("paseto: failed to set user_id: %w", err)
	}

	err = token.Set("role", claims.Role)
	if err != nil {
		return "", fmt.Errorf("paseto: failed to set role: %w", err)
	}

	token.SetIssuedAt(claims.IssuedAt)
	token.SetExpiration(claims.ExpiresAt)

	return token.V4Encrypt(key, nil), nil
}

// Parse decrypts and validates a token, returning the embedded Claims.
func Parse(tokenStr string, key paseto.V4SymmetricKey) (*Claims, error) {
	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(key, tokenStr, nil)
	if err != nil {
		if err.Error() == "this token has expired" {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
	}

	claims := &Claims{}

	tenantID, err := token.GetString("tenant_id")
	if err != nil {
		return nil, fmt.Errorf("%w: missing tenant_id", ErrTokenInvalid)
	}
	claims.TenantID = tenantID

	userID, err := token.GetString("user_id")
	if err != nil {
		return nil, fmt.Errorf("%w: missing user_id", ErrTokenInvalid)
	}
	claims.UserID = userID

	role, err := token.GetString("role")
	if err != nil {
		return nil, fmt.Errorf("%w: missing role", ErrTokenInvalid)
	}
	claims.Role = role

	iat, err := token.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("%w: missing issued_at", ErrTokenInvalid)
	}
	claims.IssuedAt = iat

	exp, err := token.GetExpiration()
	if err != nil {
		return nil, fmt.Errorf("%w: missing expiration", ErrTokenInvalid)
	}
	claims.ExpiresAt = exp

	return claims, nil
}
