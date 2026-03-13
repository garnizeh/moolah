package repository

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMasterPurchaseRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	now := time.Now()
	input := domain.CreateMasterPurchaseInput{
		AccountID:            "account_id",
		CategoryID:           "category_id",
		UserID:               "user_id",
		Description:          "iPhone",
		TotalAmountCents:     120000,
		InstallmentCount:     12,
		ClosingDay:           10,
		FirstInstallmentDate: now,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("CreateMasterPurchase", ctx, mock.MatchedBy(func(p sqlc.CreateMasterPurchaseParams) bool {
			return p.TenantID == tenantID &&
				p.AccountID == input.AccountID &&
				p.CategoryID == input.CategoryID &&
				p.UserID == input.UserID &&
				p.Description == input.Description &&
				p.TotalAmountCents == input.TotalAmountCents &&
				p.InstallmentCount == int16(input.InstallmentCount) &&
				p.ClosingDay == int16(input.ClosingDay) &&
				p.FirstInstallmentDate.Time.Equal(now)
		})).Return(sqlc.MasterPurchase{
			ID:                   "mp_id",
			TenantID:             tenantID,
			AccountID:            input.AccountID,
			CategoryID:           input.CategoryID,
			UserID:               input.UserID,
			Description:          input.Description,
			TotalAmountCents:     input.TotalAmountCents,
			InstallmentCount:     int16(input.InstallmentCount),
			PaidInstallments:     0,
			ClosingDay:           int16(input.ClosingDay),
			Status:               sqlc.MasterPurchaseStatusOpen,
			FirstInstallmentDate: pgtype.Date{Time: now, Valid: true},
			CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		got, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		assert.Equal(t, "mp_id", got.ID)
		assert.Equal(t, input.Description, got.Description)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("CreateMasterPurchase", ctx, mock.Anything).Return(sqlc.MasterPurchase{}, pgx.ErrNoRows)

		got, err := repo.Create(ctx, tenantID, input)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseRepository_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "mp_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("GetMasterPurchaseByID", ctx, sqlc.GetMasterPurchaseByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.MasterPurchase{
			ID:       id,
			TenantID: tenantID,
		}, nil)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("GetMasterPurchaseByID", ctx, mock.Anything).Return(sqlc.MasterPurchase{}, pgx.ErrNoRows)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("ListMasterPurchasesByTenant", ctx, tenantID).Return([]sqlc.MasterPurchase{
			{ID: "1", TenantID: tenantID},
			{ID: "2", TenantID: tenantID},
		}, nil)

		got, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("ListMasterPurchasesByTenant", ctx, tenantID).Return(nil, pgx.ErrNoRows)

		got, err := repo.ListByTenant(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseRepository_ListByAccount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	accountID := "acc_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("ListMasterPurchasesByAccount", ctx, sqlc.ListMasterPurchasesByAccountParams{
			TenantID:  tenantID,
			AccountID: accountID,
		}).Return([]sqlc.MasterPurchase{
			{ID: "1", TenantID: tenantID, AccountID: accountID},
		}, nil)

		got, err := repo.ListByAccount(ctx, tenantID, accountID)
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("ListMasterPurchasesByAccount", ctx, mock.Anything).Return(nil, pgx.ErrNoRows)

		got, err := repo.ListByAccount(ctx, tenantID, accountID)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseRepository_ListPendingClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	cutoff := time.Date(2023, 10, 10, 0, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("ListPendingMasterPurchasesByClosingDay", ctx, sqlc.ListPendingMasterPurchasesByClosingDayParams{
			TenantID:   tenantID,
			ClosingDay: 10,
		}).Return([]sqlc.MasterPurchase{
			{ID: "1", TenantID: tenantID, ClosingDay: 10},
		}, nil)

		got, err := repo.ListPendingClose(ctx, tenantID, cutoff)
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("ListPendingMasterPurchasesByClosingDay", ctx, mock.Anything).Return(nil, pgx.ErrNoRows)

		got, err := repo.ListPendingClose(ctx, tenantID, cutoff)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "mp_id"
	desc := "Updated iPhone"
	catID := "new_cat"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("UpdateMasterPurchase", ctx, mock.MatchedBy(func(p sqlc.UpdateMasterPurchaseParams) bool {
			return p.TenantID == tenantID && p.ID == id && p.Description.String == desc && p.CategoryID == catID
		})).Return(sqlc.MasterPurchase{
			ID:          id,
			TenantID:    tenantID,
			Description: desc,
			CategoryID:  catID,
		}, nil)

		got, err := repo.Update(ctx, tenantID, id, domain.UpdateMasterPurchaseInput{
			Description: &desc,
			CategoryID:  &catID,
		})
		require.NoError(t, err)
		assert.Equal(t, desc, got.Description)
		assert.Equal(t, catID, got.CategoryID)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("UpdateMasterPurchase", ctx, mock.Anything).Return(sqlc.MasterPurchase{}, pgx.ErrNoRows)

		got, err := repo.Update(ctx, tenantID, id, domain.UpdateMasterPurchaseInput{})
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseRepository_IncrementPaidInstallments(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "mp_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("IncrementPaidInstallments", ctx, sqlc.IncrementPaidInstallmentsParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.MasterPurchase{
			ID:               id,
			TenantID:         tenantID,
			PaidInstallments: 1,
		}, nil)

		err := repo.IncrementPaidInstallments(ctx, tenantID, id)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("IncrementPaidInstallments", ctx, mock.Anything).Return(sqlc.MasterPurchase{}, pgx.ErrNoRows)

		err := repo.IncrementPaidInstallments(ctx, tenantID, id)
		require.Error(t, err)
	})
}

func TestMasterPurchaseRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "mp_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("DeleteMasterPurchase", ctx, sqlc.DeleteMasterPurchaseParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(nil)

		err := repo.Delete(ctx, tenantID, id)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewMasterPurchaseRepository(mockQuerier)

		mockQuerier.On("DeleteMasterPurchase", ctx, mock.Anything).Return(pgx.ErrNoRows)

		err := repo.Delete(ctx, tenantID, id)
		require.Error(t, err)
	})
}
