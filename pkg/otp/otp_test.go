package otp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestOTP(t *testing.T) {
	t.Parallel()

	t.Run("Generate format and leading zeros", func(t *testing.T) {
		t.Parallel()
		// Higher cost is slow, use min for testing speed
		plain, hash, err := GenerateWithCost(bcrypt.MinCost)
		require.NoError(t, err)

		assert.Len(t, plain, 6)
		assert.Regexp(t, "^[0-9]{6}$", plain)
		assert.NotEmpty(t, hash)
		assert.Contains(t, hash, "$2a$") // bcrypt prefix
	})

	t.Run("Generate probability of leading zero", func(t *testing.T) {
		t.Parallel()
		// 1/10 chance of starting with 0. 100 tries should find at least one.
		found := false
		for range 100 {
			plain, _, err := GenerateWithCost(bcrypt.MinCost)
			require.NoError(t, err)
			if plain[0] == '0' {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to generate at least one code with a leading zero in 100 tries")
	})

	t.Run("Verify success and failure", func(t *testing.T) {
		t.Parallel()
		plain, hash, err := GenerateWithCost(bcrypt.MinCost)
		require.NoError(t, err)

		assert.NoError(t, Verify(plain, hash))
		require.Error(t, Verify("000000", hash))
		require.Error(t, Verify(plain, "invalid-hash"))
	})

	t.Run("Uniqueness check", func(t *testing.T) {
		t.Parallel()
		// Birthday paradox: with 100 random 6-digit codes (1M space),
		// p(collision) is ~0.5%. While rare, it can happen in CI.
		// We allow up to 1 collision in 100 tries for high-volume CI stability.
		const iterations = 100
		codes := make(map[string]bool)
		collisions := 0

		for range iterations {
			plain, _, err := GenerateWithCost(bcrypt.MinCost)
			require.NoError(t, err)
			if codes[plain] {
				collisions++
			}
			codes[plain] = true
		}
		assert.LessOrEqual(t, collisions, 1, "Too many duplicate codes in 100 iterations")
	})

	t.Run("Generate with default cost", func(t *testing.T) {
		t.Parallel()
		plain, hash, err := Generate()
		require.NoError(t, err)
		assert.Len(t, plain, 6)
		assert.NoError(t, Verify(plain, hash))
	})
}
