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

func TestAssetRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)
	repo := repository.NewAssetRepository(db.Queries)

	t.Run("Create and Get", func(t *testing.T) {
		t.Parallel()

		isin := "US0378331005"
		details := "Apple Inc. common stock"
		input := domain.CreateAssetInput{
			Ticker:    "AAPL",
			Name:      "Apple Inc.",
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
			ISIN:      &isin,
			Details:   &details,
		}

		created, err := repo.Create(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.NotEmpty(t, created.ID)
		require.Equal(t, input.Ticker, created.Ticker)
		require.Equal(t, input.Name, created.Name)
		require.Equal(t, input.AssetType, created.AssetType)
		require.Equal(t, input.Currency, created.Currency)
		require.Equal(t, isin, *created.ISIN)
		require.Equal(t, details, *created.Details)

		// GetByID
		got, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)

		// GetByTicker
		gotTicker, err := repo.GetByTicker(ctx, "AAPL")
		require.NoError(t, err)
		require.Equal(t, created.ID, gotTicker.ID)
	})

	t.Run("Get Non-Existent", func(t *testing.T) {
		t.Parallel()

		_, err := repo.GetByID(ctx, "non-existent")
		require.ErrorIs(t, err, domain.ErrAssetNotFound)

		_, err = repo.GetByTicker(ctx, "NONE")
		require.ErrorIs(t, err, domain.ErrAssetNotFound)
	})

	t.Run("List Assets", func(t *testing.T) {
		t.Parallel()

		_, err := repo.Create(ctx, domain.CreateAssetInput{
			Ticker:    "MSFT",
			Name:      "Microsoft",
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		_, err = repo.Create(ctx, domain.CreateAssetInput{
			Ticker:    "GOOGL",
			Name:      "Alphabet",
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		assets, err := repo.List(ctx, domain.ListAssetsParams{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(assets), 2)
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()

		asset, err := repo.Create(ctx, domain.CreateAssetInput{
			Ticker:    "DEL",
			Name:      "To Delete",
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, asset.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, asset.ID)
		require.ErrorIs(t, err, domain.ErrAssetNotFound)
	})
}
