//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/stretchr/testify/require"
)

func TestTenantAssetConfigRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)
	tenantRepo := repository.NewTenantRepository(db.Queries)
	assetRepo := repository.NewAssetRepository(db.Queries)
	repo := repository.NewTenantAssetConfigRepository(db.Queries)

	// Setup: Create tenant and asset
	tenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Config Tenant"})
	require.NoError(t, err)

	asset, err := assetRepo.Create(ctx, domain.CreateAssetInput{
		Ticker:    "AMZN",
		Name:      "Amazon",
		AssetType: domain.AssetTypeStock,
		Currency:  "USD",
	})
	require.NoError(t, err)

	t.Run("Upsert and Get", func(t *testing.T) {
		t.Parallel()

		nameOverride := "Amazon (Tenant Override)"
		currencyOverride := "USD"
		detailsOverride := "Override details"

		input := domain.UpsertTenantAssetConfigInput{
			AssetID:  asset.ID,
			Name:     &nameOverride,
			Currency: &currencyOverride,
			Details:  &detailsOverride,
		}

		// Initial Upsert
		created, err := repo.Upsert(ctx, tenant.ID, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, tenant.ID, created.TenantID)
		require.Equal(t, asset.ID, created.AssetID)
		require.Equal(t, nameOverride, *created.Name)

		// GetByAssetID
		got, err := repo.GetByAssetID(ctx, tenant.ID, asset.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, nameOverride, *got.Name)

		// Duplicate Upsert (Update)
		newNameOverride := "Amazon (Updated Override)"
		input.Name = &newNameOverride
		updated, err := repo.Upsert(ctx, tenant.ID, input)
		require.NoError(t, err)
		require.Equal(t, created.ID, updated.ID)
		require.Equal(t, newNameOverride, *updated.Name)

		// Verify get returns updated
		gotUpdated, err := repo.GetByAssetID(ctx, tenant.ID, asset.ID)
		require.NoError(t, err)
		require.Equal(t, newNameOverride, *gotUpdated.Name)
	})

	t.Run("Get Non-Existent Conf", func(t *testing.T) {
		t.Parallel()

		_, err := repo.GetByAssetID(ctx, tenant.ID, "non-existent")
		require.ErrorIs(t, err, domain.ErrAssetConfigNotFound)
	})

	t.Run("List By Tenant", func(t *testing.T) {
		t.Parallel()

		asset2, err := assetRepo.Create(ctx, domain.CreateAssetInput{
			Ticker:    "NFLX",
			Name:      "Netflix",
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		name2 := "Netflix (Tenant Override)"
		_, err = repo.Upsert(ctx, tenant.ID, domain.UpsertTenantAssetConfigInput{
			AssetID: asset2.ID,
			Name:    &name2,
		})
		require.NoError(t, err)

		configs, err := repo.ListByTenant(ctx, tenant.ID)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(configs), 2)
	})

	t.Run("Delete (Soft)", func(t *testing.T) {
		t.Parallel()

		asset3, err := assetRepo.Create(ctx, domain.CreateAssetInput{
			Ticker:    "DELCONFIG",
			Name:      "To Delete Config",
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		name3 := "Config to Delete"
		conf, err := repo.Upsert(ctx, tenant.ID, domain.UpsertTenantAssetConfigInput{
			AssetID: asset3.ID,
			Name:    &name3,
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, tenant.ID, asset3.ID)
		require.NoError(t, err)

		_, err = repo.GetByAssetID(ctx, tenant.ID, asset3.ID)
		require.ErrorIs(t, err, domain.ErrAssetConfigNotFound)

		// Attempting to upsert again should create a new record or update the soft-deleted one?
		// Note from ADR: ON CONFLICT (tenant_id, asset_id) WHERE deleted_at IS NULL DO UPDATE...
		// This means a new record can be created if the previous one was deleted.
		recreated, err := repo.Upsert(ctx, tenant.ID, domain.UpsertTenantAssetConfigInput{
			AssetID: asset3.ID,
			Name:    &name3,
		})
		require.NoError(t, err)
		require.NotEqual(t, conf.ID, recreated.ID)
	})
}
