//go:build integration

package seeds

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

// SeedPosition inserts a test position for a tenant.
func SeedPosition(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID, accountID, assetID string) domain.Position {
	t.Helper()

	id := ulid.New()
	qty := decimal.NewFromInt(10)

	row, err := q.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:                 id,
		TenantID:           tenantID,
		AccountID:          accountID,
		AssetID:            assetID,
		Quantity:           qty,
		AvgCostCents:       5000,
		LastPriceCents:     5500,
		Currency:           "USD",
		PurchasedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		IncomeType:         sqlc.IncomeTypeDividend,
		IncomeIntervalDays: pgtype.Int4{Int32: 30, Valid: true},
		IncomeAmountCents:  pgtype.Int8{Int64: 100, Valid: true},
		NextIncomeAt:       pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	return domain.Position{
		ID:                 row.ID,
		TenantID:           row.TenantID,
		AccountID:          row.AccountID,
		AssetID:            row.AssetID,
		Quantity:           qty,
		AvgCostCents:       row.AvgCostCents,
		LastPriceCents:     row.LastPriceCents,
		Currency:           row.Currency,
		PurchasedAt:        row.PurchasedAt.Time,
		IncomeType:         domain.IncomeType(row.IncomeType),
		IncomeIntervalDays: ptr(int(row.IncomeIntervalDays.Int32)),
		IncomeAmountCents:  ptr(row.IncomeAmountCents.Int64),
		NextIncomeAt:       ptr(row.NextIncomeAt.Time),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// SeedPositionIncomeEvent inserts a test income event for a position.
func SeedPositionIncomeEvent(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID, positionID, accountID string) domain.PositionIncomeEvent {
	t.Helper()

	id := ulid.New()
	row, err := q.CreatePositionIncomeEvent(ctx, sqlc.CreatePositionIncomeEventParams{
		ID:          id,
		TenantID:    tenantID,
		PositionID:  positionID,
		AmountCents: 150,
		Currency:    "USD",
		IncomeType:  sqlc.IncomeTypeDividend,
		EventDate:   pgtype.Date{Time: time.Now(), Valid: true},
		Status:      sqlc.ReceivableStatusPending,
	})
	require.NoError(t, err)

	return domain.PositionIncomeEvent{
		ID:          row.ID,
		TenantID:    row.TenantID,
		PositionID:  row.PositionID,
		AccountID:   accountID,
		AmountCents: row.AmountCents,
		Currency:    row.Currency,
		IncomeType:  domain.IncomeType(row.IncomeType),
		DueAt:       row.EventDate.Time,
		Status:      domain.ReceivableStatus(row.Status),
		CreatedAt:   row.CreatedAt.Time,
	}
}

// SeedPositionSnapshot inserts a test snapshot for a position.
func SeedPositionSnapshot(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID, positionID string) domain.PositionSnapshot {
	t.Helper()

	id := ulid.New()
	qty := decimal.NewFromInt(10)
	row, err := q.CreatePositionSnapshot(ctx, sqlc.CreatePositionSnapshotParams{
		ID:             id,
		TenantID:       tenantID,
		PositionID:     positionID,
		Quantity:       qty,
		LastPriceCents: 5500,
		SnapshotDate:   pgtype.Date{Time: time.Now(), Valid: true},
	})
	require.NoError(t, err)

	return domain.PositionSnapshot{
		ID:             row.ID,
		TenantID:       row.TenantID,
		PositionID:     row.PositionID,
		Quantity:       row.Quantity,
		LastPriceCents: row.LastPriceCents,
		SnapshotDate:   row.SnapshotDate.Time,
		CreatedAt:      row.CreatedAt.Time,
	}
}

// SeedPortfolioSnapshot inserts a test portfolio snapshot.
func SeedPortfolioSnapshot(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID string) domain.PortfolioSnapshot {
	t.Helper()

	id := ulid.New()
	row, err := q.CreatePortfolioSnapshot(ctx, sqlc.CreatePortfolioSnapshotParams{
		ID:               id,
		TenantID:         tenantID,
		Details:          []byte("{}"),
		SnapshotDate:     pgtype.Date{Time: time.Now(), Valid: true},
		TotalValueCents:  1000,
		TotalIncomeCents: 0,
		Currency:         "USD",
	})
	require.NoError(t, err)

	return domain.PortfolioSnapshot{
		ID:           row.ID,
		TenantID:     row.TenantID,
		DetailsJSON:  row.Details,
		SnapshotDate: row.SnapshotDate.Time,
		CreatedAt:    row.CreatedAt.Time,
	}
}

func ptr[T any](v T) *T {
	return &v
}
