package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTenantAssetConfigRepository_Unit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tenantID := "tenant_abc"
	assetID := "asset_123"

	t.Run("Upsert", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		name := "My Asset"
		input := domain.UpsertTenantAssetConfigInput{
			AssetID: assetID,
			Name:    &name,
		}

		mQuerier.On("UpsertTenantAssetConfig", ctx, mock.AnythingOfType("sqlc.UpsertTenantAssetConfigParams")).
			Return(sqlc.TenantAssetConfig{
				ID:        "cfg_1",
				AssetID:   assetID,
				TenantID:  tenantID,
				Name:      pgtype.Text{String: name, Valid: true},
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil).Once()

		res, err := repo.Upsert(ctx, tenantID, input)
		require.NoError(t, err)
		require.Equal(t, assetID, res.AssetID)
		require.Equal(t, name, *res.Name)
	})

	t.Run("Upsert Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("UpsertTenantAssetConfig", ctx, mock.Anything).
			Return(sqlc.TenantAssetConfig{}, errors.New("db error")).Once()

		res, err := repo.Upsert(ctx, tenantID, domain.UpsertTenantAssetConfigInput{AssetID: assetID})
		require.Error(t, err)
		require.Nil(t, res)
		require.Contains(t, err.Error(), "failed to upsert tenant asset config")
	})

	t.Run("GetByAssetID", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("GetTenantAssetConfigByAssetID", ctx, sqlc.GetTenantAssetConfigByAssetIDParams{
			TenantID: tenantID,
			AssetID:  assetID,
		}).Return(sqlc.TenantAssetConfig{ID: "cfg_1", AssetID: assetID}, nil).Once()

		res, err := repo.GetByAssetID(ctx, tenantID, assetID)
		require.NoError(t, err)
		require.Equal(t, "cfg_1", res.ID)
	})

	t.Run("GetByAssetID Not Found", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("GetTenantAssetConfigByAssetID", ctx, mock.Anything).
			Return(sqlc.TenantAssetConfig{}, pgx.ErrNoRows).Once()

		res, err := repo.GetByAssetID(ctx, tenantID, assetID)
		require.ErrorIs(t, err, domain.ErrAssetConfigNotFound)
		require.Nil(t, res)
	})

	t.Run("GetByAssetID Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("GetTenantAssetConfigByAssetID", ctx, mock.Anything).
			Return(sqlc.TenantAssetConfig{}, errors.New("db error")).Once()

		res, err := repo.GetByAssetID(ctx, tenantID, assetID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get tenant asset config by asset id")
		require.Nil(t, res)
	})

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("ListTenantAssetConfigs", ctx, tenantID).
			Return([]sqlc.TenantAssetConfig{{ID: "1"}, {ID: "2"}}, nil).Once()

		res, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		require.Len(t, res, 2)
	})

	t.Run("ListByTenant Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("ListTenantAssetConfigs", ctx, tenantID).
			Return(([]sqlc.TenantAssetConfig)(nil), errors.New("db error")).Once()

		res, err := repo.ListByTenant(ctx, tenantID)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("SoftDeleteTenantAssetConfig", ctx, sqlc.SoftDeleteTenantAssetConfigParams{
			TenantID: tenantID,
			AssetID:  assetID,
		}).Return(nil).Once()

		err := repo.Delete(ctx, tenantID, assetID)
		require.NoError(t, err)
	})

	t.Run("Delete Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewTenantAssetConfigRepository(mQuerier)
		mQuerier.On("SoftDeleteTenantAssetConfig", ctx, mock.Anything).
			Return(errors.New("db error")).Once()

		err := repo.Delete(ctx, tenantID, assetID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to soft delete tenant asset config")
	})
}
