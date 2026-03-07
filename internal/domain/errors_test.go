package domain_test

import (
	"fmt"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestDomainErrors_Wrapping(t *testing.T) {
	t.Parallel()

	type errorPair struct {
		sentinel error
		name     string
	}

	testPairs := []errorPair{
		{domain.ErrNotFound, "NotFound"},
		{domain.ErrForbidden, "Forbidden"},
		{domain.ErrConflict, "Conflict"},
		{domain.ErrInvalidInput, "InvalidInput"},
		{domain.ErrInvalidOTP, "InvalidOTP"},
		{domain.ErrOTPRateLimited, "OTPRateLimited"},
		{domain.ErrUnauthorized, "Unauthorized"},
		{domain.ErrTokenExpired, "TokenExpired"},
	}

	for _, tt := range testPairs {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wrapped := fmt.Errorf("wrapped: %w", tt.sentinel)
			require.ErrorIs(t, wrapped, tt.sentinel, "Wrapped error should match sentinel")
		})
	}
}
