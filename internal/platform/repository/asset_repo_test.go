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

func TestAssetRepository_Unit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		ticker := "AAPL"
		name := "Apple"
		isin := "US123"
		input := domain.CreateAssetInput{
			Ticker:    ticker,
			Name:      name,
			AssetType: domain.AssetTypeStock,
			Currency:  "USD",
			ISIN:      &isin,
		}

		mQuerier.On("CreateAsset", ctx, mock.AnythingOfType("sqlc.CreateAssetParams")).
			Return(sqlc.Asset{
				ID:        "asset_1",
				Ticker:    ticker,
				Name:      name,
				Isin:      pgtype.Text{String: isin, Valid: true},
				AssetType: sqlc.AssetTypeStock,
				Currency:  "USD",
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil).Once()

		res, err := repo.Create(ctx, input)
		require.NoError(t, err)
		require.Equal(t, ticker, res.Ticker)
		require.Equal(t, isin, *res.ISIN)
	})

	t.Run("Create Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		mQuerier.On("CreateAsset", ctx, mock.Anything).
			Return(sqlc.Asset{}, errors.New("db error")).Once()

		res, err := repo.Create(ctx, domain.CreateAssetInput{})
		require.Error(t, err)
		require.Nil(t, res)
		require.Contains(t, err.Error(), "failed to create asset")
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		id := "asset_1"
		mQuerier.On("GetAssetByID", ctx, id).
			Return(sqlc.Asset{ID: id, Ticker: "AAPL"}, nil).Once()

		res, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, res.ID)
	})

	t.Run("GetByID Not Found", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		id := "missing"
		mQuerier.On("GetAssetByID", ctx, id).
			Return(sqlc.Asset{}, pgx.ErrNoRows).Once()

		res, err := repo.GetByID(ctx, id)
		require.ErrorIs(t, err, domain.ErrAssetNotFound)
		require.Nil(t, res)
	})

	t.Run("GetByID Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		id := "error"
		mQuerier.On("GetAssetByID", ctx, id).
			Return(sqlc.Asset{}, errors.New("db error")).Once()

		res, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get asset by id")
		require.Nil(t, res)
	})

	t.Run("GetByTicker", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		ticker := "AAPL"
		mQuerier.On("GetAssetByTicker", ctx, ticker).
			Return(sqlc.Asset{ID: "1", Ticker: ticker}, nil).Once()

		res, err := repo.GetByTicker(ctx, ticker)
		require.NoError(t, err)
		require.Equal(t, ticker, res.Ticker)
	})

	t.Run("GetByTicker Not Found", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		ticker := "NONE"
		mQuerier.On("GetAssetByTicker", ctx, ticker).
			Return(sqlc.Asset{}, pgx.ErrNoRows).Once()

		res, err := repo.GetByTicker(ctx, ticker)
		require.ErrorIs(t, err, domain.ErrAssetNotFound)
		require.Nil(t, res)
	})

	t.Run("GetByTicker Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		ticker := "ERR"
		mQuerier.On("GetAssetByTicker", ctx, ticker).
			Return(sqlc.Asset{}, errors.New("db error")).Once()

		res, err := repo.GetByTicker(ctx, ticker)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get asset by ticker")
		require.Nil(t, res)
	})

	t.Run("List", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		mQuerier.On("ListAssets", ctx).
			Return([]sqlc.Asset{{ID: "1"}, {ID: "2"}}, nil).Once()

		res, err := repo.List(ctx, domain.ListAssetsParams{})
		require.NoError(t, err)
		require.Len(t, res, 2)
	})

	t.Run("List Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		mQuerier.On("ListAssets", ctx).
			Return(([]sqlc.Asset)(nil), errors.New("db error")).Once()

		res, err := repo.List(ctx, domain.ListAssetsParams{})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		id := "1"
		mQuerier.On("DeleteAsset", ctx, id).Return(nil).Once()

		err := repo.Delete(ctx, id)
		require.NoError(t, err)
	})

	t.Run("Delete Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := repository.NewAssetRepository(mQuerier)
		id := "1"
		mQuerier.On("DeleteAsset", ctx, id).Return(errors.New("db error")).Once()

		err := repo.Delete(ctx, id)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete asset")
	})
}
