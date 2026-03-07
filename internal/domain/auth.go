package domain

import (
	"context"
	"time"
)

// OTPRequest represents a pending or consumed OTP challenge.
// It tracks the email associated with the request, the bcrypt hash of the 6-digit code,
// whether it has been used, and its expiration time.
type OTPRequest struct {
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CodeHash  string    `json:"-"` // bcrypt hash of the 6-digit code, never serialized
	Used      bool      `json:"used"`
}

// CreateOTPRequestInput is the value object used to create a new OTP challenge.
type CreateOTPRequestInput struct {
	ExpiresAt time.Time `validate:"required"`
	Email     string    `validate:"required,email"`
	CodeHash  string    `validate:"required"`
}

// AuthRepository defines persistence operations for OTP challenges and authentication state.
// It decouples the auth service from specific database implementations.
type AuthRepository interface {
	// CreateOTPRequest persists a new OTP challenge in the database.
	CreateOTPRequest(ctx context.Context, input CreateOTPRequestInput) (*OTPRequest, error)

	// GetActiveOTPRequest retrieves the most recent unused, non-expired OTP for the given email.
	// Returns domain.ErrInvalidOTP if no valid request is found.
	GetActiveOTPRequest(ctx context.Context, email string) (*OTPRequest, error)

	// MarkOTPUsed marks the given OTP request as consumed to prevent reuse.
	MarkOTPUsed(ctx context.Context, id string) error

	// DeleteExpiredOTPRequests removes all expired OTP rows from the database.
	// This is typically called by a periodic cleanup job.
	DeleteExpiredOTPRequests(ctx context.Context) error
}

// Claims holds the data encoded in a PASETO token.
// It contains identification for the user and their tenant, as well as their role and metadata.
type Claims struct {
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Role      Role      `json:"role"`
}
