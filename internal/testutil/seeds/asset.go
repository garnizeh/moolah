//go:build integration

package seeds

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

// SeedAsset inserts a test asset.
func SeedAsset(t *testing.T, ctx context.Context, q sqlc.Querier) domain.Asset {
	t.Helper()

	id := ulid.New()
	ticker := "TICK-" + id[:8]
	row, err := q.CreateAsset(ctx, sqlc.CreateAssetParams{
		ID:        id,
		Ticker:    ticker,
		Name:      "Test Asset " + id,
		AssetType: sqlc.AssetTypeStock,
		Currency:  "USD",
	})
	require.NoError(t, err)

	return domain.Asset{
		ID:        row.ID,
		Ticker:    row.Ticker,
		Name:      row.Name,
		AssetType: domain.AssetType(row.AssetType),
		Currency:  row.Currency,
		CreatedAt: row.CreatedAt.Time,
	}
}

// SeedTenantAssetConfig inserts a test tenant asset configuration.
func SeedTenantAssetConfig(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID, assetID string) domain.TenantAssetConfig {
	t.Helper()

	id := ulid.New()
	nameOverride := "Overridden Name"
	row, err := q.UpsertTenantAssetConfig(ctx, sqlc.UpsertTenantAssetConfigParams{
		ID:       id,
		TenantID: tenantID,
		AssetID:  assetID,
		Name:     pgtype.Text{String: nameOverride, Valid: true},
	})
	require.NoError(t, err)

	return domain.TenantAssetConfig{
		ID:        row.ID,
		TenantID:  row.TenantID,
		AssetID:   row.AssetID,
		Name:      &nameOverride,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
