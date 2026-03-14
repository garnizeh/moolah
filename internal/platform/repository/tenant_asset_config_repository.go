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

// TenantAssetConfigRepository provides methods to manage tenant-specific asset configurations in the database.
type TenantAssetConfigRepository struct {
	q sqlc.Querier
}

// NewTenantAssetConfigRepository creates a new instance of TenantAssetConfigRepository with the provided sqlc.Querier.
func NewTenantAssetConfigRepository(q sqlc.Querier) *TenantAssetConfigRepository {
	return &TenantAssetConfigRepository{q: q}
}

// Upsert creates or updates a tenant asset configuration based on the provided input and returns the resulting configuration.
func (r *TenantAssetConfigRepository) Upsert(ctx context.Context, tenantID string, input domain.UpsertTenantAssetConfigInput) (*domain.TenantAssetConfig, error) {
	row, err := r.q.UpsertTenantAssetConfig(ctx, sqlc.UpsertTenantAssetConfigParams{
		ID:       ulid.New(),
		TenantID: tenantID,
		AssetID:  input.AssetID,
		Name:     toText(input.Name),
		Currency: toText(input.Currency),
		Details:  toText(input.Details),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert tenant asset config: %w", err)
	}

	return mapTenantAssetConfig(row), nil
}

// GetByAssetID retrieves a tenant asset configuration by its asset ID and tenant ID.
// If the configuration does not exist, it returns a domain.ErrAssetConfigNotFound error.
func (r *TenantAssetConfigRepository) GetByAssetID(ctx context.Context, tenantID, assetID string) (*domain.TenantAssetConfig, error) {
	row, err := r.q.GetTenantAssetConfigByAssetID(ctx, sqlc.GetTenantAssetConfigByAssetIDParams{
		TenantID: tenantID,
		AssetID:  assetID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: (tenant: %s, asset: %s)", domain.ErrAssetConfigNotFound, tenantID, assetID)
		}
		return nil, fmt.Errorf("failed to get tenant asset config by asset id: %w", err)
	}

	return mapTenantAssetConfig(row), nil
}

// ListByTenant returns all tenant asset configurations for the given tenant ID.
func (r *TenantAssetConfigRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.TenantAssetConfig, error) {
	rows, err := r.q.ListTenantAssetConfigs(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant asset configs: %w", err)
	}

	configs := make([]domain.TenantAssetConfig, len(rows))
	for i, row := range rows {
		configs[i] = *mapTenantAssetConfig(row)
	}

	return configs, nil
}

// Delete performs a soft delete of the tenant asset configuration by its asset ID and tenant ID.
func (r *TenantAssetConfigRepository) Delete(ctx context.Context, tenantID, assetID string) error {
	err := r.q.SoftDeleteTenantAssetConfig(ctx, sqlc.SoftDeleteTenantAssetConfigParams{
		TenantID: tenantID,
		AssetID:  assetID,
	})
	if err != nil {
		return fmt.Errorf("failed to soft delete tenant asset config: %w", err)
	}
	return nil
}

// mapTenantAssetConfig converts a sqlc.TenantAssetConfig to a domain.TenantAssetConfig, handling nullable fields appropriately.
func mapTenantAssetConfig(row sqlc.TenantAssetConfig) *domain.TenantAssetConfig {
	return &domain.TenantAssetConfig{
		ID:        row.ID,
		TenantID:  row.TenantID,
		AssetID:   row.AssetID,
		Name:      fromText(row.Name),
		Currency:  fromText(row.Currency),
		Details:   fromText(row.Details),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: fromTime(row.DeletedAt),
	}
}
