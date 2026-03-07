package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// DefaultBcryptCost is the cost used for hashing OTPs.
const DefaultBcryptCost = 12

// Generate produces a cryptographically random 6-digit code and its bcrypt hash.
// The caller must send `plain` to the user and persist `hash` in the database.
// The plain-text code must never be stored.
func Generate() (string, string, error) {
	return GenerateWithCost(DefaultBcryptCost)
}

// GenerateWithCost produces a cryptographically random 6-digit code and its bcrypt hash with a custom cost.
func GenerateWithCost(cost int) (string, string, error) {
	const maxOTP = 1_000_000
	n, err := rand.Int(rand.Reader, big.NewInt(maxOTP))
	if err != nil {
		return "", "", fmt.Errorf("otp: failed to generate random number: %w", err)
	}

	plain := fmt.Sprintf("%06d", n.Int64())
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", "", fmt.Errorf("otp: failed to hash code: %w", err)
	}

	return plain, string(hash), nil
}

// Verify returns nil if plain matches the stored bcrypt hash, or an error otherwise.
func Verify(plain, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err != nil {
		return fmt.Errorf("otp: verification failed: %w", err)
	}
	return nil
}
