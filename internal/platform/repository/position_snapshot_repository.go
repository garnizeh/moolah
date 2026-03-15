package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ domain.PositionSnapshotRepository = (*PositionSnapshotRepository)(nil)

// PositionSnapshotRepository handles holding state history.
type PositionSnapshotRepository struct {
	q sqlc.Querier
}

// NewPositionSnapshotRepository returns a new PositionSnapshotRepository instance.
func NewPositionSnapshotRepository(q sqlc.Querier) *PositionSnapshotRepository {
	return &PositionSnapshotRepository{q: q}
}

// Create persists a new position snapshot.
func (r *PositionSnapshotRepository) Create(ctx context.Context, tenantID string, in domain.CreatePositionSnapshotInput) (*domain.PositionSnapshot, error) {
	arg := sqlc.CreatePositionSnapshotParams{
		ID:             ulid.New(),
		TenantID:       tenantID,
		PositionID:     in.PositionID,
		SnapshotDate:   pgtype.Date{Time: in.SnapshotDate, Valid: true},
		Quantity:       in.Quantity,
		LastPriceCents: in.LastPriceCents,
		Currency:       in.Currency,
		AvgCostCents:   0, // Not in input yet, but required by schema if needed
	}

	snap, err := r.q.CreatePositionSnapshot(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create position snapshot: %w", err)
	}

	return r.mapToDomain(snap), nil
}

// ListByPosition retrieves historical snapshots for a position.
func (r *PositionSnapshotRepository) ListByPosition(ctx context.Context, tenantID, positionID string) ([]domain.PositionSnapshot, error) {
	list, err := r.q.ListPositionSnapshotsByPosition(ctx, sqlc.ListPositionSnapshotsByPositionParams{
		PositionID: positionID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list position snapshots: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// ListByTenantSince retrieves snapshots for a tenant since a date.
func (r *PositionSnapshotRepository) ListByTenantSince(ctx context.Context, tenantID string, since time.Time) ([]domain.PositionSnapshot, error) {
	list, err := r.q.ListPositionSnapshotsByTenantSince(ctx, sqlc.ListPositionSnapshotsByTenantSinceParams{
		TenantID:     tenantID,
		SnapshotDate: pgtype.Date{Time: since, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant snapshots: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// mapToDomain converts a sqlc PositionSnapshot to a domain PositionSnapshot.
func (r *PositionSnapshotRepository) mapToDomain(s sqlc.PositionSnapshot) *domain.PositionSnapshot {
	return &domain.PositionSnapshot{
		ID:             s.ID,
		TenantID:       s.TenantID,
		PositionID:     s.PositionID,
		SnapshotDate:   s.SnapshotDate.Time,
		Quantity:       s.Quantity,
		LastPriceCents: s.LastPriceCents,
		Currency:       s.Currency,
		CreatedAt:      s.CreatedAt.Time,
	}
}

// mapSliceToDomain converts a slice of sqlc PositionSnapshots to a slice of domain PositionSnapshots.
func (r *PositionSnapshotRepository) mapSliceToDomain(list []sqlc.PositionSnapshot) []domain.PositionSnapshot {
	res := make([]domain.PositionSnapshot, len(list))
	for i, s := range list {
		res[i] = *r.mapToDomain(s)
	}
	return res
}
