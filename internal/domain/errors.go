package domain

import "errors"

// Sentinel errors — all in internal/domain/errors.go
var (
	// ErrNotFound is returned when a requested resource does not exist or has been soft-deleted.
	ErrNotFound = errors.New("not found")

	// ErrForbidden is returned when the caller lacks permission for a resource.
	ErrForbidden = errors.New("forbidden")

	// ErrConflict is returned when an operation would violate a uniqueness constraint.
	ErrConflict = errors.New("conflict")

	// ErrInvalidInput is returned when request validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidOTP is returned when an OTP code is wrong, expired, or already used.
	ErrInvalidOTP = errors.New("invalid or expired OTP")

	// ErrOTPRateLimited is returned when too many OTP requests are made.
	ErrOTPRateLimited = errors.New("OTP rate limit exceeded")

	// ErrUnauthorized is returned when no valid token is present.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTokenExpired is returned when the PASETO token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrAssetNotFound is returned when a requested global asset does not exist.
	ErrAssetNotFound = errors.New("asset not found")

	// ErrAssetConfigNotFound is returned when a requested tenant asset config does not exist.
	ErrAssetConfigNotFound = errors.New("tenant asset configuration not found")
)
