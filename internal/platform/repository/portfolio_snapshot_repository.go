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

var _ domain.PortfolioSnapshotRepository = (*PortfolioSnapshotRepository)(nil)

// PortfolioSnapshotRepository handles top-level portfolio aggregate state.
type PortfolioSnapshotRepository struct {
	q sqlc.Querier
}

// NewPortfolioSnapshotRepository returns a new PortfolioSnapshotRepository instance.
func NewPortfolioSnapshotRepository(q sqlc.Querier) *PortfolioSnapshotRepository {
	return &PortfolioSnapshotRepository{q: q}
}

// Create persists a new portfolio snapshot.
func (r *PortfolioSnapshotRepository) Create(ctx context.Context, tenantID string, in domain.CreatePortfolioSnapshotInput) (*domain.PortfolioSnapshot, error) {
	arg := sqlc.CreatePortfolioSnapshotParams{
		ID:           ulid.New(),
		TenantID:     tenantID,
		SnapshotDate: pgtype.Date{Time: in.SnapshotDate, Valid: true},
		Details:      in.DetailsJSON,
		// Schema has other fields but if not in input, we use defaults
		Currency: "BRL", // Default currency
	}

	snap, err := r.q.CreatePortfolioSnapshot(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio snapshot: %w", err)
	}

	return r.mapToDomain(snap), nil
}

// GetByDate retrieves a snapshot for a tenant on a date.
func (r *PortfolioSnapshotRepository) GetByDate(ctx context.Context, tenantID string, date time.Time) (*domain.PortfolioSnapshot, error) {
	snap, err := r.q.GetPortfolioSnapshotByDate(ctx, sqlc.GetPortfolioSnapshotByDateParams{
		SnapshotDate: pgtype.Date{Time: date, Valid: true},
		TenantID:     tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio snapshot: %w", err)
	}

	return r.mapToDomain(snap), nil
}

// ListByTenant lists all aggregate snapshots for a tenant.
func (r *PortfolioSnapshotRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.PortfolioSnapshot, error) {
	list, err := r.q.ListPortfolioSnapshots(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list portfolio snapshots: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// mapToDomain converts a sqlc PortfolioSnapshot to a domain PortfolioSnapshot.
func (r *PortfolioSnapshotRepository) mapToDomain(s sqlc.PortfolioSnapshot) *domain.PortfolioSnapshot {
	return &domain.PortfolioSnapshot{
		ID:           s.ID,
		TenantID:     s.TenantID,
		SnapshotDate: s.SnapshotDate.Time,
		DetailsJSON:  s.Details,
		CreatedAt:    s.CreatedAt.Time,
	}
}

// mapSliceToDomain converts a slice of sqlc PortfolioSnapshots to a slice of domain PortfolioSnapshots.
func (r *PortfolioSnapshotRepository) mapSliceToDomain(list []sqlc.PortfolioSnapshot) []domain.PortfolioSnapshot {
	res := make([]domain.PortfolioSnapshot, len(list))
	for i, s := range list {
		res[i] = *r.mapToDomain(s)
	}
	return res
}
