package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPositionRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z1"
	assetID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z2"
	accountID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z3"
	now := time.Now().UTC()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		incomeInterval := 30
		incomeAmount := int64(100)
		nextIncome := now.AddDate(0, 1, 0)
		qty := decimal.NewFromInt(10)

		input := domain.CreatePositionInput{
			AssetID:            assetID,
			AccountID:          accountID,
			Quantity:           qty,
			AvgCostCents:       1000,
			LastPriceCents:     1100,
			Currency:           "BRL",
			PurchasedAt:        now,
			IncomeType:         domain.IncomeTypeDividend,
			IncomeIntervalDays: &incomeInterval,
			IncomeAmountCents:  &incomeAmount,
			NextIncomeAt:       &nextIncome,
		}

		mQuerier.On("CreatePosition", ctx, mock.MatchedBy(func(arg sqlc.CreatePositionParams) bool {
			return arg.TenantID == tenantID &&
				arg.AssetID == assetID &&
				arg.Quantity.Equal(input.Quantity) &&
				arg.IncomeIntervalDays.Int32 == int32(incomeInterval) &&
				arg.IncomeAmountCents.Int64 == incomeAmount
		})).Return(sqlc.Position{
			ID:                 "POS_1",
			TenantID:           tenantID,
			AssetID:            assetID,
			AccountID:          accountID,
			Quantity:           qty,
			IncomeType:         sqlc.IncomeTypeDividend,
			IncomeIntervalDays: pgtype.Int4{Int32: 30, Valid: true},
			IncomeAmountCents:  pgtype.Int8{Int64: 100, Valid: true},
			CreatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		pos, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		require.NotNil(t, pos)
		assert.Equal(t, "POS_1", pos.ID)
		assert.Equal(t, incomeInterval, *pos.IncomeIntervalDays)
		assert.Equal(t, incomeAmount, *pos.IncomeAmountCents)
		assert.True(t, pos.Quantity.Equal(qty))
	})

	t.Run("Create Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("CreatePosition", ctx, mock.Anything).Return(sqlc.Position{}, fmt.Errorf("db error"))

		pos, err := repo.Create(ctx, tenantID, domain.CreatePositionInput{})
		require.Error(t, err)
		assert.Nil(t, pos)
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		id := "POS_1"
		mQuerier.On("GetPositionByID", ctx, sqlc.GetPositionByIDParams{
			ID:       id,
			TenantID: tenantID,
		}).Return(sqlc.Position{
			ID:       id,
			TenantID: tenantID,
		}, nil)

		pos, err := repo.GetByID(ctx, tenantID, id)
		require.NoError(t, err)
		require.NotNil(t, pos)
		assert.Equal(t, id, pos.ID)
	})

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("ListPositionsByTenant", ctx, tenantID).Return([]sqlc.Position{
			{ID: "P1", TenantID: tenantID},
			{ID: "P2", TenantID: tenantID},
		}, nil)

		list, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("ListByAccount", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("ListPositionsByAccount", ctx, sqlc.ListPositionsByAccountParams{
			TenantID:  tenantID,
			AccountID: accountID,
		}).Return([]sqlc.Position{
			{ID: "P1", TenantID: tenantID, AccountID: accountID},
		}, nil)

		list, err := repo.ListByAccount(ctx, tenantID, accountID)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("ListDueIncome", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		before := now
		mQuerier.On("ListPositionsDueIncome", ctx, pgtype.Timestamptz{Time: before, Valid: true}).
			Return([]sqlc.Position{{ID: "P1"}}, nil)

		list, err := repo.ListDueIncome(ctx, before)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("GetByID Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("GetPositionByID", ctx, mock.Anything).Return(sqlc.Position{}, fmt.Errorf("db error"))

		pos, err := repo.GetByID(ctx, tenantID, "any")
		require.Error(t, err)
		assert.Nil(t, pos)
	})

	t.Run("ListByTenant Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("ListPositionsByTenant", ctx, tenantID).Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListByTenant(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, list)
	})

	t.Run("ListByAccount Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("ListPositionsByAccount", ctx, mock.Anything).Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListByAccount(ctx, tenantID, accountID)
		require.Error(t, err)
		assert.Nil(t, list)
	})

	t.Run("ListDueIncome Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("ListPositionsDueIncome", ctx, mock.Anything).Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListDueIncome(ctx, time.Now())
		require.Error(t, err)
		assert.Nil(t, list)
	})

	t.Run("Update Success", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		id := "POS_1"
		newQty := decimal.NewFromInt(20)
		newInterval := 60

		// Mock Get for update
		mQuerier.On("GetPositionByID", ctx, sqlc.GetPositionByIDParams{ID: id, TenantID: tenantID}).
			Return(sqlc.Position{
				ID:                 id,
				Quantity:           decimal.NewFromInt(10),
				IncomeIntervalDays: pgtype.Int4{Int32: 30, Valid: true},
				IncomeType:         sqlc.IncomeTypeDividend,
			}, nil)

		// Mock Update
		mQuerier.On("UpdatePosition", ctx, mock.MatchedBy(func(arg sqlc.UpdatePositionParams) bool {
			return arg.ID == id && arg.Quantity.Equal(newQty) && arg.IncomeIntervalDays.Int32 == int32(newInterval)
		})).Return(sqlc.Position{
			ID:                 id,
			Quantity:           newQty,
			IncomeIntervalDays: pgtype.Int4{Int32: int32(newInterval), Valid: true},
		}, nil)

		pos, err := repo.Update(ctx, tenantID, id, domain.UpdatePositionInput{
			Quantity:           &newQty,
			IncomeIntervalDays: &newInterval,
		})
		require.NoError(t, err)
		require.NotNil(t, pos)
		assert.True(t, pos.Quantity.Equal(newQty))
	})

	t.Run("Update Database Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		id := "POS_1"
		mQuerier.On("GetPositionByID", ctx, mock.Anything).Return(sqlc.Position{ID: id}, nil)
		mQuerier.On("UpdatePosition", ctx, mock.Anything).Return(sqlc.Position{}, fmt.Errorf("db error"))

		pos, err := repo.Update(ctx, tenantID, id, domain.UpdatePositionInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update position")
		assert.Nil(t, pos)
	})

	t.Run("Update Not Found", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("GetPositionByID", ctx, mock.Anything).Return(sqlc.Position{}, fmt.Errorf("not found"))

		pos, err := repo.Update(ctx, tenantID, "any", domain.UpdatePositionInput{})
		require.Error(t, err)
		assert.Nil(t, pos)
	})

	t.Run("Delete Success", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		id := "POS_1"
		mQuerier.On("SoftDeletePosition", ctx, sqlc.SoftDeletePositionParams{
			ID:       id,
			TenantID: tenantID,
		}).Return(nil)

		err := repo.Delete(ctx, tenantID, id)
		assert.NoError(t, err)
	})

	t.Run("Delete Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionRepository(mQuerier)

		mQuerier.On("SoftDeletePosition", ctx, mock.Anything).Return(fmt.Errorf("db error"))

		err := repo.Delete(ctx, tenantID, "any")
		require.Error(t, err)
	})
}
