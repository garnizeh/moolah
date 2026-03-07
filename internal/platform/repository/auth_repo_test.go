package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthRepo_CreateOTPRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := domain.CreateOTPRequestInput{
		Email:     "test@example.com",
		CodeHash:  "hashedcode",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("CreateOTPRequest", ctx, mock.MatchedBy(func(arg sqlc.CreateOTPRequestParams) bool {
			return arg.Email == input.Email && arg.CodeHash == input.CodeHash && arg.ExpiresAt.Time.Equal(input.ExpiresAt) && arg.ID != ""
		})).Return(sqlc.OtpRequest{
			ID:        "01H7XFRP9K1A1A1A1A1A1A1A1C",
			Email:     input.Email,
			CodeHash:  input.CodeHash,
			Used:      false,
			ExpiresAt: pgtype.Timestamptz{Time: input.ExpiresAt, Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		otp, err := repo.CreateOTPRequest(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, otp)
		assert.Equal(t, input.Email, otp.Email)
		assert.Equal(t, input.CodeHash, otp.CodeHash)
		assert.False(t, otp.Used)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("CreateOTPRequest", ctx, mock.Anything).Return(sqlc.OtpRequest{}, errors.New("db error"))

		otp, err := repo.CreateOTPRequest(ctx, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create otp request")
		assert.Nil(t, otp)
	})
}

func TestAuthRepo_GetActiveOTPRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "test@example.com"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("GetActiveOTPByEmail", ctx, email).Return(sqlc.OtpRequest{
			ID:        "01H7XFRP9K1A1A1A1A1A1A1A1C",
			Email:     email,
			CodeHash:  "hashedcode",
			Used:      false,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		otp, err := repo.GetActiveOTPRequest(ctx, email)

		require.NoError(t, err)
		assert.NotNil(t, otp)
		assert.Equal(t, email, otp.Email)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("GetActiveOTPByEmail", ctx, email).Return(sqlc.OtpRequest{}, pgx.ErrNoRows)

		otp, err := repo.GetActiveOTPRequest(ctx, email)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
		assert.Nil(t, otp)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("GetActiveOTPByEmail", ctx, email).Return(sqlc.OtpRequest{}, errors.New("db error"))

		otp, err := repo.GetActiveOTPRequest(ctx, email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get active otp request")
		assert.Nil(t, otp)
	})
}

func TestAuthRepo_MarkOTPUsed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "01H7XFRP9K1A1A1A1A1A1A1A1C"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("MarkOTPUsed", ctx, id).Return(nil)

		err := repo.MarkOTPUsed(ctx, id)

		require.NoError(t, err)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("MarkOTPUsed", ctx, id).Return(pgx.ErrNoRows)

		err := repo.MarkOTPUsed(ctx, id)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidOTP)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("MarkOTPUsed", ctx, id).Return(errors.New("db error"))

		err := repo.MarkOTPUsed(ctx, id)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to mark otp as used")
	})
}

func TestAuthRepo_DeleteExpiredOTPRequests(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("DeleteExpiredOTPs", ctx).Return(nil)

		err := repo.DeleteExpiredOTPRequests(ctx)

		require.NoError(t, err)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewAuthRepository(mockQuerier)

		mockQuerier.On("DeleteExpiredOTPs", ctx).Return(errors.New("db error"))

		err := repo.DeleteExpiredOTPRequests(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete expired otp requests")
	})
}
