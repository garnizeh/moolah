package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIncomeSchedulerJob_processDueIncome(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tenantID := "tenant-1"
	now := time.Now()

	t.Run("creates income event and updates position next_income_at", func(t *testing.T) {
		t.Parallel()

		posRepo := new(mocks.PositionRepository)
		incomeRepo := new(mocks.PositionIncomeEventRepository)
		job := NewIncomeSchedulerJob(posRepo, incomeRepo, logger, 1*time.Hour)

		interval := 30
		nextIncomeAt := now.Add(-1 * time.Hour)
		pos := domain.Position{
			ID:                 "pos-1",
			TenantID:           tenantID,
			AccountID:          "acc-1",
			AssetID:            "asset-1",
			Currency:           "USD",
			Quantity:           decimal.NewFromFloat(100),
			LastPriceCents:     1000, // $10
			IncomeType:         domain.IncomeTypeDividend,
			NextIncomeAt:       &nextIncomeAt,
			IncomeIntervalDays: &interval,
			IncomeAmountCents:  new(int64(500)), // $5 fixed
		}

		posRepo.On("ListDueIncome", mock.Anything, mock.Anything).Return([]domain.Position{pos}, nil)
		incomeRepo.On("Create", mock.Anything, tenantID, mock.MatchedBy(func(in domain.CreatePositionIncomeEventInput) bool {
			return in.AmountCents == 500 && in.PositionID == "pos-1"
		})).Return(&domain.PositionIncomeEvent{}, nil)

		expectedNextIncomeAt := nextIncomeAt.AddDate(0, 0, interval)
		posRepo.On("Update", mock.Anything, tenantID, "pos-1", mock.MatchedBy(func(in domain.UpdatePositionInput) bool {
			return in.NextIncomeAt != nil && in.NextIncomeAt.Equal(expectedNextIncomeAt)
		})).Return(&domain.Position{}, nil)

		job.processDueIncome(context.Background())

		posRepo.AssertExpectations(t)
		incomeRepo.AssertExpectations(t)
	})

	t.Run("calculates correctly for rate-based income", func(t *testing.T) {
		t.Parallel()

		posRepo := new(mocks.PositionRepository)
		incomeRepo := new(mocks.PositionIncomeEventRepository)
		job := NewIncomeSchedulerJob(posRepo, incomeRepo, logger, 1*time.Hour)

		nextIncomeAt := now.Add(-1 * time.Hour)
		pos := domain.Position{
			ID:             "pos-rate",
			TenantID:       tenantID,
			Quantity:       decimal.NewFromFloat(100),
			LastPriceCents: 1000, // value = 100 * 10 = 100000 cents ($1000)
			IncomeType:     domain.IncomeTypeInterest,
			NextIncomeAt:   &nextIncomeAt,
			IncomeRateBps:  new(100), // 1% = 100 bps
		}

		// calculation: 100000 * 1% = 1000 cents ($10)
		expectedAmount := int64(1000)

		posRepo.On("ListDueIncome", mock.Anything, mock.Anything).Return([]domain.Position{pos}, nil)
		incomeRepo.On("Create", mock.Anything, tenantID, mock.MatchedBy(func(in domain.CreatePositionIncomeEventInput) bool {
			return in.AmountCents == expectedAmount
		})).Return(&domain.PositionIncomeEvent{}, nil)
		posRepo.On("Update", mock.Anything, tenantID, "pos-rate", mock.Anything).Return(&domain.Position{}, nil)

		job.processDueIncome(context.Background())

		posRepo.AssertExpectations(t)
		incomeRepo.AssertExpectations(t)
	})

	t.Run("handles repository errors gracefully", func(t *testing.T) {
		t.Parallel()

		posRepo := new(mocks.PositionRepository)
		incomeRepo := new(mocks.PositionIncomeEventRepository)
		job := NewIncomeSchedulerJob(posRepo, incomeRepo, logger, 1*time.Hour)

		posRepo.On("ListDueIncome", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

		// Should not panic, just log
		job.processDueIncome(context.Background())

		posRepo.AssertExpectations(t)
	})
}

func TestIncomeSchedulerJob_Run(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	posRepo := new(mocks.PositionRepository)
	incomeRepo := new(mocks.PositionIncomeEventRepository)
	job := NewIncomeSchedulerJob(posRepo, incomeRepo, logger, 10*time.Millisecond)

	posRepo.On("ListDueIncome", mock.Anything, mock.Anything).Return([]domain.Position{}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := job.Run(ctx)
	require.NoError(t, err)

	posRepo.AssertExpectations(t)
}
