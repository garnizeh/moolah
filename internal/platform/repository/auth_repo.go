package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type authRepo struct {
	q sqlc.Querier
}

// NewAuthRepository creates a new concrete implementation of domain.AuthRepository.
func NewAuthRepository(q sqlc.Querier) domain.AuthRepository {
	return &authRepo{q: q}
}

// CreateOTPRequest persists a new OTP challenge in the database.
func (r *authRepo) CreateOTPRequest(ctx context.Context, input domain.CreateOTPRequestInput) (*domain.OTPRequest, error) {
	arg := sqlc.CreateOTPRequestParams{
		ID:       ulid.New(),
		Email:    input.Email,
		CodeHash: input.CodeHash,
		Used:     false,
		ExpiresAt: pgtype.Timestamptz{
			Time:  input.ExpiresAt,
			Valid: true,
		},
	}

	o, err := r.q.CreateOTPRequest(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create otp request: %w", err)
	}

	return r.mapOTPRequest(o), nil
}

// GetActiveOTPRequest retrieves the most recent unused, non-expired OTP for the given email.
func (r *authRepo) GetActiveOTPRequest(ctx context.Context, email string) (*domain.OTPRequest, error) {
	o, err := r.q.GetActiveOTPByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidOTP
		}
		return nil, fmt.Errorf("failed to get active otp request: %w", err)
	}

	return r.mapOTPRequest(o), nil
}

// MarkOTPUsed marks the given OTP request as consumed to prevent reuse.
func (r *authRepo) MarkOTPUsed(ctx context.Context, id string) error {
	err := r.q.MarkOTPUsed(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInvalidOTP
		}
		return fmt.Errorf("failed to mark otp as used: %w", err)
	}
	return nil
}

// DeleteExpiredOTPRequests removes all expired OTP rows from the database.
func (r *authRepo) DeleteExpiredOTPRequests(ctx context.Context) error {
	err := r.q.DeleteExpiredOTPs(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete expired otp requests: %w", err)
	}
	return nil
}

func (r *authRepo) mapOTPRequest(o sqlc.OtpRequest) *domain.OTPRequest {
	return &domain.OTPRequest{
		ID:        o.ID,
		Email:     o.Email,
		CodeHash:  o.CodeHash,
		Used:      o.Used,
		ExpiresAt: o.ExpiresAt.Time,
		CreatedAt: o.CreatedAt.Time,
	}
}
