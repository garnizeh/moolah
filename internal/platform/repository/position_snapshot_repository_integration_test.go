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

func TestPositionSnapshotRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)
	repo := repository.NewPositionSnapshotRepository(db.Queries)

	// Seed dependencies
	tenant := seeds.SeedTenant(t, ctx, db.Queries)
	user := seeds.SeedUser(t, ctx, db.Queries, tenant.ID)
	account := seeds.SeedAccount(t, ctx, db.Queries, tenant.ID, user.ID)
	asset := seeds.SeedAsset(t, ctx, db.Queries)

	// Create a position for snapshots
	posRepo := repository.NewPositionRepository(db.Queries)
	_, err := posRepo.Create(ctx, tenant.ID, domain.CreatePositionInput{
		AssetID:     asset.ID,
		AccountID:   account.ID,
		Quantity:    decimal.NewFromInt(10),
		Currency:    "USD",
		PurchasedAt: time.Now(),
		IncomeType:  domain.IncomeTypeNone,
	})
	require.NoError(t, err)

	t.Run("Create & List By Position", func(t *testing.T) {
		t.Parallel()
		// Use a truly unique asset to avoid uq_assets_ticker collisions
		asset1, err := db.Queries.CreateAsset(ctx, sqlc.CreateAssetParams{
			ID:        ulid.New(),
			Ticker:    "SN-PS-" + ulid.New()[:12],
			Name:      "Snapshot Pos 1",
			AssetType: sqlc.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		pos1, err := posRepo.Create(ctx, tenant.ID, domain.CreatePositionInput{
			AssetID:     asset1.ID,
			AccountID:   account.ID,
			Quantity:    decimal.NewFromInt(10),
			Currency:    "USD",
			PurchasedAt: time.Now(),
			IncomeType:  domain.IncomeTypeNone,
		})
		require.NoError(t, err)

		in := domain.CreatePositionSnapshotInput{
			PositionID:     pos1.ID,
			SnapshotDate:   time.Now().AddDate(0, 0, -2), // Use distinct date
			Quantity:       decimal.NewFromInt(10),
			LastPriceCents: 1200,
			Currency:       "USD",
		}
		snap, errSnap := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, errSnap)
		require.NotEmpty(t, snap.ID)

		list, errList := repo.ListByPosition(ctx, tenant.ID, pos1.ID)
		require.NoError(t, errList)
		require.NotEmpty(t, list)
		require.Equal(t, snap.ID, list[0].ID)
	})

	t.Run("List By Tenant Since", func(t *testing.T) {
		t.Parallel()
		// Use a truly unique asset to avoid uq_assets_ticker collisions
		asset2, err := db.Queries.CreateAsset(ctx, sqlc.CreateAssetParams{
			ID:        ulid.New(),
			Ticker:    "SN-TS-" + ulid.New()[:12],
			Name:      "Snapshot Since 1",
			AssetType: sqlc.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		pos2, err := posRepo.Create(ctx, tenant.ID, domain.CreatePositionInput{
			AssetID:     asset2.ID,
			AccountID:   account.ID,
			Quantity:    decimal.NewFromInt(10),
			Currency:    "USD",
			PurchasedAt: time.Now(),
			IncomeType:  domain.IncomeTypeNone,
		})
		require.NoError(t, err)

		since := time.Now().Add(-24 * time.Hour)
		in := domain.CreatePositionSnapshotInput{
			PositionID:     pos2.ID,
			SnapshotDate:   time.Now(),
			Quantity:       decimal.NewFromInt(10),
			LastPriceCents: 1250,
			Currency:       "USD",
		}
		_, errSnap := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, errSnap)

		list, errList := repo.ListByTenantSince(ctx, tenant.ID, since.AddDate(0, 0, -1))
		require.NoError(t, errList)
		require.NotEmpty(t, list)
	})
}
