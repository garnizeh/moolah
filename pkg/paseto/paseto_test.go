package paseto

import (
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaseto(t *testing.T) {
	t.Parallel()

	key := paseto.NewV4SymmetricKey()
	now := time.Now().Truncate(time.Second) // PASETO uses second precision for storage
	claims := Claims{
		TenantID:  "01JKH7G6F5E4D3C2B1A0987654",
		UserID:    "01JKH7G6F5E4D3C2B1A098765X",
		Role:      "admin",
		IssuedAt:  now,
		ExpiresAt: now.Add(time.Hour),
	}

	t.Run("Round-trip success", func(t *testing.T) {
		t.Parallel()
		tokenStr, err := Seal(claims, key)
		require.NoError(t, err)

		parsed, err := Parse(tokenStr, key)

		require.NoError(t, err)
		assert.Equal(t, claims.TenantID, parsed.TenantID)
		assert.Equal(t, claims.UserID, parsed.UserID)
		assert.Equal(t, claims.Role, parsed.Role)
		assert.True(t, claims.IssuedAt.Equal(parsed.IssuedAt))
		assert.True(t, claims.ExpiresAt.Equal(parsed.ExpiresAt))
	})

	t.Run("Expired token", func(t *testing.T) {
		t.Parallel()
		expiredClaims := claims
		expiredClaims.IssuedAt = now.Add(-2 * time.Hour)
		expiredClaims.ExpiresAt = now.Add(-1 * time.Hour)

		tokenStr, err := Seal(expiredClaims, key)
		require.NoError(t, err)

		_, err = Parse(tokenStr, key)

		require.ErrorIs(t, err, ErrTokenExpired)
	})

	t.Run("Wrong key", func(t *testing.T) {
		t.Parallel()
		otherKey := paseto.NewV4SymmetricKey()
		tokenStr, err := Seal(claims, key)
		require.NoError(t, err)

		_, err = Parse(tokenStr, otherKey)

		require.ErrorIs(t, err, ErrTokenInvalid)
	})

	t.Run("Tampered token", func(t *testing.T) {
		t.Parallel()
		tokenStr, err := Seal(claims, key)
		require.NoError(t, err)

		// Tamper with the token string
		tampered := tokenStr[:len(tokenStr)-5] + "abcde"
		_, err = Parse(tampered, key)

		require.ErrorIs(t, err, ErrTokenInvalid)
	})

	t.Run("Malformed token", func(t *testing.T) {
		t.Parallel()
		_, err := Parse("v4.local.not-a-token", key)
		require.ErrorIs(t, err, ErrTokenInvalid)
	})

	t.Run("Missing custom claims", func(t *testing.T) {
		t.Parallel()

		// Missing tenant_id
		t.Run("missing tenant_id", func(t *testing.T) {
			t.Parallel()
			token := paseto.NewToken()
			token.SetExpiration(time.Now().Add(time.Hour))
			token.SetIssuedAt(time.Now())
			err := token.Set("user_id", "u1")
			require.NoError(t, err)
			err = token.Set("role", "r1")
			require.NoError(t, err)
			tokenStr := token.V4Encrypt(key, nil)
			_, err = Parse(tokenStr, key)
			require.ErrorIs(t, err, ErrTokenInvalid)
			assert.Contains(t, err.Error(), "missing tenant_id")
		})

		// Missing user_id
		t.Run("missing user_id", func(t *testing.T) {
			t.Parallel()
			token := paseto.NewToken()
			token.SetExpiration(time.Now().Add(time.Hour))
			token.SetIssuedAt(time.Now())
			err := token.Set("tenant_id", "t1")
			require.NoError(t, err)
			err = token.Set("role", "r1")
			require.NoError(t, err)
			tokenStr := token.V4Encrypt(key, nil)
			_, err = Parse(tokenStr, key)
			require.ErrorIs(t, err, ErrTokenInvalid)
			assert.Contains(t, err.Error(), "missing user_id")
		})

		// Missing role
		t.Run("missing role", func(t *testing.T) {
			t.Parallel()
			token := paseto.NewToken()
			token.SetExpiration(time.Now().Add(time.Hour))
			token.SetIssuedAt(time.Now())
			err := token.Set("tenant_id", "t1")
			require.NoError(t, err)
			err = token.Set("user_id", "u1")
			require.NoError(t, err)
			tokenStr := token.V4Encrypt(key, nil)
			_, err = Parse(tokenStr, key)
			require.ErrorIs(t, err, ErrTokenInvalid)
			assert.Contains(t, err.Error(), "missing role")
		})
	})

	t.Run("Missing standard claims", func(t *testing.T) {
		t.Parallel()

		// Missing iat
		t.Run("missing iat", func(t *testing.T) {
			t.Parallel()
			token := paseto.NewToken()
			token.SetExpiration(time.Now().Add(time.Hour))
			err := token.Set("tenant_id", "t1")
			require.NoError(t, err)
			err = token.Set("user_id", "u1")
			require.NoError(t, err)
			err = token.Set("role", "r1")
			require.NoError(t, err)
			tokenStr := token.V4Encrypt(key, nil)
			_, err = Parse(tokenStr, key)
			require.ErrorIs(t, err, ErrTokenInvalid)
			assert.Contains(t, err.Error(), "missing issued_at")
		})

		// Missing exp
		t.Run("missing exp", func(t *testing.T) {
			t.Parallel()
			token := paseto.NewToken()
			token.SetIssuedAt(time.Now())
			err := token.Set("tenant_id", "t1")
			require.NoError(t, err)
			err = token.Set("user_id", "u1")
			require.NoError(t, err)
			err = token.Set("role", "r1")
			require.NoError(t, err)
			tokenStr := token.V4Encrypt(key, nil)
			_, err = Parse(tokenStr, key)
			require.ErrorIs(t, err, ErrTokenInvalid)
			// Either our custom error from paseto.go or the library's claim error
			assert.True(t, strings.Contains(err.Error(), "missing expiration") || strings.Contains(err.Error(), "exp' not present in claims"), "Error message should mention missing expiration or exp claim")
		})
	})
}
