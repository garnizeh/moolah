package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5"
)

// AssetRepo provides methods to manage financial assets in the database.
type AssetRepo struct {
	q sqlc.Querier
}

// NewAssetRepository creates a new instance of AssetRepository with the provided sqlc.Querier.
func NewAssetRepository(q sqlc.Querier) *AssetRepo {
	return &AssetRepo{q: q}
}

// Create adds a new asset to the database based on the provided input and returns the created asset.
func (r *AssetRepo) Create(ctx context.Context, input domain.CreateAssetInput) (*domain.Asset, error) {
	row, err := r.q.CreateAsset(ctx, sqlc.CreateAssetParams{
		ID:        ulid.New(),
		Ticker:    input.Ticker,
		Isin:      toText(input.ISIN),
		Name:      input.Name,
		AssetType: sqlc.AssetType(input.AssetType),
		Currency:  input.Currency,
		Details:   toText(input.Details),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return mapAsset(row), nil
}

// GetByID retrieves an asset by its unique ID. If the asset does not exist, it returns a domain.ErrAssetNotFound error.
func (r *AssetRepo) GetByID(ctx context.Context, id string) (*domain.Asset, error) {
	row, err := r.q.GetAssetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", domain.ErrAssetNotFound, id)
		}
		return nil, fmt.Errorf("failed to get asset by id: %w", err)
	}

	return mapAsset(row), nil
}

// GetByTicker retrieves an asset by its ticker symbol. If the asset does not exist, it returns a domain.ErrAssetNotFound error.
func (r *AssetRepo) GetByTicker(ctx context.Context, ticker string) (*domain.Asset, error) {
	row, err := r.q.GetAssetByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", domain.ErrAssetNotFound, ticker)
		}
		return nil, fmt.Errorf("failed to get asset by ticker: %w", err)
	}

	return mapAsset(row), nil
}

// List returns a list of assets based on the provided filtering and pagination parameters. Currently,
// the sqlc ListAssets query does not support filtering/pagination, so this method returns all assets.
// If Phase 3 requires filtering/pagination, the SQL query and this method would need to be updated accordingly.
func (r *AssetRepo) List(ctx context.Context, params domain.ListAssetsParams) ([]domain.Asset, error) {
	// Note: The sqlc ListAssets query in assets.sql.go currently doesn't take params.
	// If Phase 3 requires filtering/pagination in the repo layer, we would need to update the SQL.
	// For now, listing all as per generated sqlc.
	rows, err := r.q.ListAssets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	assets := make([]domain.Asset, len(rows))
	for i, row := range rows {
		assets[i] = *mapAsset(row)
	}

	return assets, nil
}

// Delete removes an asset from the database by its ID. If the deletion fails, it returns an error.
func (r *AssetRepo) Delete(ctx context.Context, id string) error {
	err := r.q.DeleteAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}
	return nil
}

// mapAsset converts a sqlc.Asset to a domain.Asset, handling nullable fields appropriately.
func mapAsset(row sqlc.Asset) *domain.Asset {
	return &domain.Asset{
		ID:        row.ID,
		Ticker:    row.Ticker,
		ISIN:      fromText(row.Isin),
		Name:      row.Name,
		AssetType: domain.AssetType(row.AssetType),
		Currency:  row.Currency,
		Details:   fromText(row.Details),
		CreatedAt: row.CreatedAt.Time,
	}
}

// GetLastPrice returns the most recent price for an asset. In a real app, this might call an external API.
func (r *AssetRepo) GetLastPrice(ctx context.Context, id string) (int64, error) {
	// For MVP, return a static value or 0 if not found
	return 0, nil
}

// GetAssetWithTenantConfig retrieves a global asset and applies any tenant-specific overrides from the database.
func (r *AssetRepo) GetAssetWithTenantConfig(ctx context.Context, tenantID, id string) (*domain.Asset, error) {
	row, err := r.q.GetAssetWithTenantConfig(ctx, sqlc.GetAssetWithTenantConfigParams{
		TenantID: tenantID,
		AssetID:  id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", domain.ErrAssetNotFound, id)
		}
		return nil, fmt.Errorf("failed to get asset with tenant config: %w", err)
	}

	return &domain.Asset{
		ID:        row.ID,
		Ticker:    row.Ticker,
		ISIN:      fromText(row.Isin),
		Name:      row.Name,
		AssetType: domain.AssetType(row.AssetType),
		Currency:  row.Currency,
		Details:   fromText(row.Details),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
