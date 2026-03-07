package repository

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	input := domain.CreateAccountInput{
		UserID:       "user_id",
		Name:         "Checking",
		Type:         domain.AccountTypeChecking,
		Currency:     "USD",
		InitialCents: 1000,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("CreateAccount", ctx, mock.MatchedBy(func(p sqlc.CreateAccountParams) bool {
			return p.TenantID == tenantID && p.Name == input.Name && p.Type == sqlc.AccountType(input.Type)
		})).Return(sqlc.Account{
			ID:           "acc_id",
			TenantID:     tenantID,
			UserID:       input.UserID,
			Name:         input.Name,
			Type:         sqlc.AccountType(input.Type),
			Currency:     input.Currency,
			BalanceCents: input.InitialCents,
		}, nil)

		got, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		assert.Equal(t, "acc_id", got.ID)
		assert.Equal(t, input.Name, got.Name)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("CreateAccount", ctx, mock.Anything).Return(sqlc.Account{}, pgx.ErrNoRows)

		got, err := repo.Create(ctx, tenantID, input)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestAccountRepository_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "acc_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("GetAccountByID", ctx, sqlc.GetAccountByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.Account{
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
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{}, pgx.ErrNoRows)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestAccountRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("ListAccountsByTenant", ctx, tenantID).Return([]sqlc.Account{
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
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("ListAccountsByTenant", ctx, mock.Anything).Return(nil, pgx.ErrNoRows)

		got, err := repo.ListByTenant(ctx, tenantID)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestAccountRepository_ListByUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	userID := "user_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("ListAccountsByUser", ctx, sqlc.ListAccountsByUserParams{
			TenantID: tenantID,
			UserID:   userID,
		}).Return([]sqlc.Account{
			{ID: "1", TenantID: tenantID, UserID: userID},
		}, nil)

		got, err := repo.ListByUser(ctx, tenantID, userID)
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})
}

func TestAccountRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "acc_id"
	name := "Updated Name"
	input := domain.UpdateAccountInput{
		Name: &name,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("GetAccountByID", ctx, sqlc.GetAccountByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.Account{
			ID:   id,
			Name: "Old Name",
		}, nil)

		mockQuerier.On("UpdateAccount", ctx, mock.MatchedBy(func(p sqlc.UpdateAccountParams) bool {
			return p.Name == name
		})).Return(sqlc.Account{
			ID:   id,
			Name: name,
		}, nil)

		got, err := repo.Update(ctx, tenantID, id, input)
		require.NoError(t, err)
		assert.Equal(t, name, got.Name)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("GetAccountByID", ctx, mock.Anything).Return(sqlc.Account{}, nil)
		mockQuerier.On("UpdateAccount", ctx, mock.Anything).Return(sqlc.Account{}, &pgconn.PgError{Code: "23505"})

		got, err := repo.Update(ctx, tenantID, id, input)
		require.ErrorIs(t, err, domain.ErrConflict)
		assert.Nil(t, got)
	})
}

func TestAccountRepository_UpdateBalance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "acc_id"
	delta := int64(500)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("UpdateAccountBalance", ctx, sqlc.UpdateAccountBalanceParams{
			TenantID:     tenantID,
			ID:           id,
			BalanceCents: delta,
		}).Return(nil)

		err := repo.UpdateBalance(ctx, tenantID, id, delta)
		require.NoError(t, err)
	})
}

func TestAccountRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "acc_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("SoftDeleteAccount", ctx, sqlc.SoftDeleteAccountParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(nil)

		err := repo.Delete(ctx, tenantID, id)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAccountRepository(mockQuerier)

		mockQuerier.On("SoftDeleteAccount", ctx, mock.Anything).Return(pgx.ErrNoRows)

		err := repo.Delete(ctx, tenantID, id)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}
