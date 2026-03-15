//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/internal/testutil/seeds"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPositionRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)
	repo := repository.NewPositionRepository(db.Queries)

	// Seed dependencies
	tenant := seeds.SeedTenant(t, ctx, db.Queries)
	user := seeds.SeedUser(t, ctx, db.Queries, tenant.ID)
	account := seeds.SeedAccount(t, ctx, db.Queries, tenant.ID, user.ID)
	_ = seeds.SeedAsset(t, ctx, db.Queries)

	t.Run("Create & Get", func(t *testing.T) {
		t.Parallel()
		// Use a truly unique asset to avoid uq_assets_ticker collisions
		asset1, err := db.Queries.CreateAsset(ctx, sqlc.CreateAssetParams{
			ID:        ulid.New(),
			Ticker:    "POS-CG-" + ulid.New()[:12],
			Name:      "Position Create Get",
			AssetType: sqlc.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		in := domain.CreatePositionInput{
			AssetID:        asset1.ID,
			AccountID:      account.ID,
			Quantity:       decimal.NewFromInt(10),
			AvgCostCents:   1000,
			LastPriceCents: 1100,
			Currency:       "USD",
			PurchasedAt:    time.Now(),
			IncomeType:     domain.IncomeTypeDividend,
		}

		pos, err := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, err)
		require.NotEmpty(t, pos.ID)
		require.Equal(t, in.AssetID, pos.AssetID)

		fetched, err := repo.GetByID(ctx, tenant.ID, pos.ID)
		require.NoError(t, err)
		require.Equal(t, pos.ID, fetched.ID)
	})

	t.Run("List By Tenant & Account", func(t *testing.T) {
		t.Parallel()
		// Use a truly unique asset to avoid uq_assets_ticker collisions
		asset2, err := db.Queries.CreateAsset(ctx, sqlc.CreateAssetParams{
			ID:        ulid.New(),
			Ticker:    "POS-LT-" + ulid.New()[:12],
			Name:      "Position List Tenant",
			AssetType: sqlc.AssetTypeStock,
			Currency:  "BRL",
		})
		require.NoError(t, err)

		in := domain.CreatePositionInput{
			AssetID:        asset2.ID,
			AccountID:      account.ID,
			Quantity:       decimal.NewFromInt(1),
			AvgCostCents:   100,
			LastPriceCents: 100,
			Currency:       "BRL",
			PurchasedAt:    time.Now(),
			IncomeType:     domain.IncomeTypeNone,
		}
		_, err = repo.Create(ctx, tenant.ID, in)
		require.NoError(t, err)

		list, err := repo.ListByTenant(ctx, tenant.ID)
		require.NoError(t, err)
		require.NotEmpty(t, list)

		accList, err := repo.ListByAccount(ctx, tenant.ID, account.ID)
		require.NoError(t, err)
		require.NotEmpty(t, accList)
	})

	t.Run("List Due Income", func(t *testing.T) {
		t.Parallel()
		// Use a truly unique asset to avoid uq_assets_ticker collisions
		asset3, err := db.Queries.CreateAsset(ctx, sqlc.CreateAssetParams{
			ID:        ulid.New(),
			Ticker:    "POS-DI-" + ulid.New()[:12],
			Name:      "Position Due Income",
			AssetType: sqlc.AssetTypeStock,
			Currency:  "BRL",
		})
		require.NoError(t, err)

		next := time.Now().Add(-1 * time.Hour)
		in := domain.CreatePositionInput{
			AssetID:      asset3.ID,
			AccountID:    account.ID,
			Quantity:     decimal.NewFromInt(1),
			Currency:     "BRL",
			PurchasedAt:  time.Now(),
			IncomeType:   domain.IncomeTypeRent,
			NextIncomeAt: &next,
		}
		_, errIn := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, errIn)

		due, errDue := repo.ListDueIncome(ctx, time.Now())
		require.NoError(t, errDue)
		require.NotEmpty(t, due)
	})

	t.Run("Update & Delete", func(t *testing.T) {
		t.Parallel()
		// Use a truly unique asset to avoid uq_assets_ticker collisions
		asset4, err := db.Queries.CreateAsset(ctx, sqlc.CreateAssetParams{
			ID:        ulid.New(),
			Ticker:    "POS-UD-" + ulid.New()[:12],
			Name:      "Position Update Delete",
			AssetType: sqlc.AssetTypeStock,
			Currency:  "BRL",
		})
		require.NoError(t, err)

		in := domain.CreatePositionInput{
			AssetID:     asset4.ID,
			AccountID:   account.ID,
			Quantity:    decimal.NewFromInt(1),
			Currency:    "BRL",
			PurchasedAt: time.Now(),
			IncomeType:  domain.IncomeTypeNone,
		}
		pos, err := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, err)

		newQty := decimal.NewFromInt(50)
		updIn := domain.UpdatePositionInput{Quantity: &newQty}
		updated, errUpd := repo.Update(ctx, tenant.ID, pos.ID, updIn)
		require.NoError(t, errUpd)
		require.Equal(t, "50", updated.Quantity.String())

		errDel := repo.Delete(ctx, tenant.ID, pos.ID)
		require.NoError(t, errDel)

		_, errGet := repo.GetByID(ctx, tenant.ID, pos.ID)
		require.Error(t, errGet) // Should fail after soft delete
	})
}
