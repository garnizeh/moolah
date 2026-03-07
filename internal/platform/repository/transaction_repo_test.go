package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	now := time.Now()
	input := domain.CreateTransactionInput{
		OccurredAt:  now,
		AccountID:   "account_id",
		CategoryID:  "category_id",
		Description: "Lunch",
		AmountCents: 1500,
		Type:        domain.TransactionTypeExpense,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("CreateTransaction", ctx, mock.MatchedBy(func(p sqlc.CreateTransactionParams) bool {
			return p.TenantID == tenantID &&
				p.AccountID == input.AccountID &&
				p.CategoryID == input.CategoryID &&
				p.Description == input.Description &&
				p.AmountCents == input.AmountCents &&
				p.Type == sqlc.TransactionType(input.Type) &&
				p.OccurredAt.Time.Equal(now)
		})).Return(sqlc.Transaction{
			ID:          "trans_id",
			TenantID:    tenantID,
			AccountID:   input.AccountID,
			CategoryID:  input.CategoryID,
			Description: input.Description,
			AmountCents: input.AmountCents,
			Type:        sqlc.TransactionType(input.Type),
			OccurredAt:  pgtype.Timestamptz{Time: now, Valid: true},
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		got, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		assert.Equal(t, "trans_id", got.ID)
		assert.Equal(t, input.Description, got.Description)
	})

	t.Run("with master purchase id", func(t *testing.T) {
		t.Parallel()
		inputWithMP := input
		inputWithMP.MasterPurchaseID = "mp_id"
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("CreateTransaction", ctx, mock.MatchedBy(func(p sqlc.CreateTransactionParams) bool {
			return p.MasterPurchaseID.String == "mp_id" && p.MasterPurchaseID.Valid
		})).Return(sqlc.Transaction{
			ID:               "trans_id",
			MasterPurchaseID: pgtype.Text{String: "mp_id", Valid: true},
		}, nil)

		got, err := repo.Create(ctx, tenantID, inputWithMP)
		require.NoError(t, err)
		assert.Equal(t, "mp_id", got.MasterPurchaseID)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("CreateTransaction", ctx, mock.Anything).Return(sqlc.Transaction{}, pgx.ErrNoRows)

		got, err := repo.Create(ctx, tenantID, input)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestTransactionRepository_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "trans_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("GetTransactionByID", ctx, sqlc.GetTransactionByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.Transaction{
			ID:       id,
			TenantID: tenantID,
		}, nil)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("GetTransactionByID", ctx, mock.Anything).Return(sqlc.Transaction{}, pgx.ErrNoRows)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestTransactionRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	params := domain.ListTransactionsParams{
		Limit:  10,
		Offset: 0,
	}

	t.Run("success simple", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("ListTransactionsByTenant", ctx, tenantID).Return([]sqlc.Transaction{
			{ID: "1", TenantID: tenantID},
			{ID: "2", TenantID: tenantID},
		}, nil)

		got, err := repo.List(ctx, tenantID, params)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("ListTransactionsByTenant", ctx, mock.Anything).Return(nil, pgx.ErrNoRows)

		got, err := repo.List(ctx, tenantID, params)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestTransactionRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "trans_id"
	now := time.Now()
	amount := int64(2000)

	input := domain.UpdateTransactionInput{
		AmountCents: &amount,
		OccurredAt:  &now,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		// Get current
		mockQuerier.On("GetTransactionByID", ctx, sqlc.GetTransactionByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.Transaction{
			ID:          id,
			TenantID:    tenantID,
			AmountCents: 1500,
			Description: "Old Desc",
		}, nil)

		// Update
		mockQuerier.On("UpdateTransaction", ctx, mock.MatchedBy(func(p sqlc.UpdateTransactionParams) bool {
			return p.AmountCents == 2000 &&
				p.Description == "Old Desc" &&
				p.OccurredAt.Time.Equal(now)
		})).Return(sqlc.Transaction{
			ID:          id,
			TenantID:    tenantID,
			AmountCents: 2000,
			Description: "Old Desc",
		}, nil)

		got, err := repo.Update(ctx, tenantID, id, input)
		require.NoError(t, err)
		assert.Equal(t, int64(2000), got.AmountCents)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("GetTransactionByID", ctx, mock.Anything).Return(sqlc.Transaction{}, pgx.ErrNoRows)

		got, err := repo.Update(ctx, tenantID, id, input)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("GetTransactionByID", ctx, mock.Anything).Return(sqlc.Transaction{}, nil)
		mockQuerier.On("UpdateTransaction", ctx, mock.Anything).Return(sqlc.Transaction{}, &pgconn.PgError{Code: "23503"})

		got, err := repo.Update(ctx, tenantID, id, input)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("partial update", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		desc := "New Description"
		inputPartial := domain.UpdateTransactionInput{
			Description: &desc,
		}

		mockQuerier.On("GetTransactionByID", ctx, mock.Anything).Return(sqlc.Transaction{
			ID:          id,
			TenantID:    tenantID,
			AmountCents: 1500,
			Description: "Old Desc",
		}, nil)

		mockQuerier.On("UpdateTransaction", ctx, mock.MatchedBy(func(p sqlc.UpdateTransactionParams) bool {
			return p.Description == desc && p.AmountCents == 1500
		})).Return(sqlc.Transaction{
			ID:          id,
			Description: desc,
			AmountCents: 1500,
		}, nil)

		got, err := repo.Update(ctx, tenantID, id, inputPartial)
		require.NoError(t, err)
		assert.Equal(t, desc, got.Description)
	})

	t.Run("full partial update", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		catID := "new_cat"
		amount := int64(3000)
		now := time.Now()
		inputFull := domain.UpdateTransactionInput{
			CategoryID:  &catID,
			AmountCents: &amount,
			OccurredAt:  &now,
		}

		mockQuerier.On("GetTransactionByID", ctx, mock.Anything).Return(sqlc.Transaction{
			ID: id,
		}, nil)

		mockQuerier.On("UpdateTransaction", ctx, mock.MatchedBy(func(p sqlc.UpdateTransactionParams) bool {
			return p.CategoryID == catID && p.AmountCents == amount && p.OccurredAt.Time.Equal(now)
		})).Return(sqlc.Transaction{}, nil)

		_, err := repo.Update(ctx, tenantID, id, inputFull)
		require.NoError(t, err)
	})
}

func TestTransactionRepository_OtherBranches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tenantID := "tenant_id"

	t.Run("translate unique violation", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)
		mockQuerier.On("SoftDeleteTransaction", ctx, mock.Anything).Return(&pgconn.PgError{Code: "23505"})
		err := repo.Delete(ctx, tenantID, "id")
		require.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("translate generic error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)
		genericErr := errors.New("generic error")
		mockQuerier.On("SoftDeleteTransaction", ctx, mock.Anything).Return(genericErr)
		err := repo.Delete(ctx, tenantID, "id")
		assert.ErrorIs(t, err, genericErr)
	})
}

func TestTransactionRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "trans_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("SoftDeleteTransaction", ctx, sqlc.SoftDeleteTransactionParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(nil)

		err := repo.Delete(ctx, tenantID, id)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTransactionRepository(mockQuerier)

		mockQuerier.On("SoftDeleteTransaction", ctx, mock.Anything).Return(pgx.ErrNoRows)

		err := repo.Delete(ctx, tenantID, id)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}
