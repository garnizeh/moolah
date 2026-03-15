//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/internal/testutil/seeds"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPositionIncomeEventRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)
	repo := repository.NewPositionIncomeEventRepository(db.Queries)

	// Seed dependencies
	tenant := seeds.SeedTenant(t, ctx, db.Queries)
	user := seeds.SeedUser(t, ctx, db.Queries, tenant.ID)
	account := seeds.SeedAccount(t, ctx, db.Queries, tenant.ID, user.ID)
	asset := seeds.SeedAsset(t, ctx, db.Queries)

	// Create position for income events
	posRepo := repository.NewPositionRepository(db.Queries)
	pos, err := posRepo.Create(ctx, tenant.ID, domain.CreatePositionInput{
		AssetID:     asset.ID,
		AccountID:   account.ID,
		Quantity:    decimal.NewFromInt(1),
		Currency:    "BRL",
		PurchasedAt: time.Now(),
		IncomeType:  domain.IncomeTypeRent,
	})
	require.NoError(t, err)

	t.Run("Create & Get", func(t *testing.T) {
		t.Parallel()
		in := domain.CreatePositionIncomeEventInput{
			PositionID:  pos.ID,
			AccountID:   account.ID,
			DueAt:       time.Now().Add(24 * time.Hour),
			AmountCents: 500,
			Currency:    "BRL",
			IncomeType:  domain.IncomeTypeRent,
		}
		event, errEvent := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, errEvent)
		require.NotEmpty(t, event.ID)

		var errGet error
		fetched, errGet := repo.GetByID(ctx, tenant.ID, event.ID)
		require.NoError(t, errGet)
		require.Equal(t, event.ID, fetched.ID)
	})

	t.Run("List By Tenant & Pending", func(t *testing.T) {
		t.Parallel()
		in := domain.CreatePositionIncomeEventInput{
			PositionID:  pos.ID,
			AccountID:   account.ID,
			DueAt:       time.Now().Add(1 * time.Hour),
			AmountCents: 100,
			Currency:    "BRL",
			IncomeType:  domain.IncomeTypeRent,
		}
		_, err = repo.Create(ctx, tenant.ID, in)
		require.NoError(t, err)

		list, err := repo.ListByTenant(ctx, tenant.ID)
		require.NoError(t, err)
		require.NotEmpty(t, list)

		pending, err := repo.ListPending(ctx, tenant.ID)
		require.NoError(t, err)
		require.NotEmpty(t, pending)
	})

	t.Run("Update Status - State Machine Enforcement", func(t *testing.T) {
		t.Parallel()
		in := domain.CreatePositionIncomeEventInput{
			PositionID:  pos.ID,
			AccountID:   account.ID,
			DueAt:       time.Now().Add(24 * time.Hour),
			AmountCents: 1000,
			Currency:    "USD",
			IncomeType:  domain.IncomeTypeDividend,
		}
		event, errUpd := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, errUpd)

		// Transition from Pending to Received
		receivedAt := time.Now()
		updated, err := repo.UpdateStatus(ctx, tenant.ID, event.ID, domain.ReceivableStatusReceived, &receivedAt)
		require.NoError(t, err)
		require.Equal(t, domain.ReceivableStatusReceived, updated.Status)

		// Attempt second transition should fail (ErrConflict)
		_, err = repo.UpdateStatus(ctx, tenant.ID, event.ID, domain.ReceivableStatusCancelled, nil)
		require.ErrorIs(t, err, domain.ErrConflict)
	})
}
