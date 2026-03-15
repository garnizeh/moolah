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

var _ domain.PositionRepository = (*PositionRepository)(nil)

// PositionRepository handles holding persistence.
type PositionRepository struct {
	q sqlc.Querier
}

// NewPositionRepository returns a new PositionRepository instance.
func NewPositionRepository(q sqlc.Querier) *PositionRepository {
	return &PositionRepository{q: q}
}

// Create persists a new position.
func (r *PositionRepository) Create(ctx context.Context, tenantID string, in domain.CreatePositionInput) (*domain.Position, error) {
	arg := sqlc.CreatePositionParams{
		ID:                 ulid.New(),
		TenantID:           tenantID,
		AssetID:            in.AssetID,
		AccountID:          in.AccountID,
		Quantity:           in.Quantity,
		AvgCostCents:       in.AvgCostCents,
		LastPriceCents:     in.LastPriceCents,
		Currency:           in.Currency,
		PurchasedAt:        pgtype.Timestamptz{Time: in.PurchasedAt, Valid: true},
		IncomeType:         sqlc.IncomeType(in.IncomeType),
		IncomeIntervalDays: valOrNil(in.IncomeIntervalDays),
		IncomeAmountCents:  valOrNil64(in.IncomeAmountCents),
		IncomeRateBps:      valOrNil(in.IncomeRateBps),
		NextIncomeAt:       pgtype.Timestamptz{Time: safeTime(in.NextIncomeAt), Valid: in.NextIncomeAt != nil},
		MaturityAt:         pgtype.Timestamptz{Time: safeTime(in.MaturityAt), Valid: in.MaturityAt != nil},
	}

	p, err := r.q.CreatePosition(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	return r.mapToDomain(p), nil
}

// GetByID retrieves a position by ID.
func (r *PositionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Position, error) {
	p, err := r.q.GetPositionByID(ctx, sqlc.GetPositionByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	return r.mapToDomain(p), nil
}

// ListByTenant lists positions for a tenant.
func (r *PositionRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Position, error) {
	list, err := r.q.ListPositionsByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// ListByAccount lists positions for an account.
func (r *PositionRepository) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.Position, error) {
	list, err := r.q.ListPositionsByAccount(ctx, sqlc.ListPositionsByAccountParams{
		TenantID:  tenantID,
		AccountID: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list account positions: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// ListDueIncome lists positions due for income globally (for scheduler).
func (r *PositionRepository) ListDueIncome(ctx context.Context, before time.Time) ([]domain.Position, error) {
	list, err := r.q.ListPositionsDueIncome(ctx, pgtype.Timestamptz{Time: before, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list due income: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// Update modifies a position.
func (r *PositionRepository) Update(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error) {
	// First get current to merge or just use params
	curr, err := r.q.GetPositionByID(ctx, sqlc.GetPositionByIDParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("failed to find position for update: %w", err)
	}

	arg := sqlc.UpdatePositionParams{
		ID:                 id,
		TenantID:           tenantID,
		Quantity:           valOrDefault(in.Quantity, curr.Quantity),
		AvgCostCents:       valOrDefault(in.AvgCostCents, curr.AvgCostCents),
		LastPriceCents:     valOrDefault(in.LastPriceCents, curr.LastPriceCents),
		IncomeType:         sqlc.IncomeType(valOrDefault(in.IncomeType, domain.IncomeType(curr.IncomeType))),
		IncomeIntervalDays: toPgInt4(in.IncomeIntervalDays, curr.IncomeIntervalDays),
		IncomeAmountCents:  toPgInt8(in.IncomeAmountCents, curr.IncomeAmountCents),
		IncomeRateBps:      toPgInt4(in.IncomeRateBps, curr.IncomeRateBps),
		NextIncomeAt:       toPgTimestamptz(in.NextIncomeAt, curr.NextIncomeAt),
		MaturityAt:         toPgTimestamptz(in.MaturityAt, curr.MaturityAt),
	}

	p, err := r.q.UpdatePosition(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return r.mapToDomain(p), nil
}

// Delete soft deletes a position.
func (r *PositionRepository) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.SoftDeletePosition(ctx, sqlc.SoftDeletePositionParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}
	return nil
}

// mapToDomain converts a sqlc Position to a domain Position.
func (r *PositionRepository) mapToDomain(p sqlc.Position) *domain.Position {
	return &domain.Position{
		ID:                 p.ID,
		TenantID:           p.TenantID,
		AssetID:            p.AssetID,
		AccountID:          p.AccountID,
		Quantity:           p.Quantity,
		AvgCostCents:       p.AvgCostCents,
		LastPriceCents:     p.LastPriceCents,
		Currency:           p.Currency,
		PurchasedAt:        p.PurchasedAt.Time,
		IncomeType:         domain.IncomeType(p.IncomeType),
		IncomeIntervalDays: fromPgInt4(p.IncomeIntervalDays),
		IncomeAmountCents:  fromPgInt8(p.IncomeAmountCents),
		IncomeRateBps:      fromPgInt4(p.IncomeRateBps),
		NextIncomeAt:       fromPgTimestamptz(p.NextIncomeAt),
		MaturityAt:         fromPgTimestamptz(p.MaturityAt),
		CreatedAt:          p.CreatedAt.Time,
		UpdatedAt:          p.UpdatedAt.Time,
		DeletedAt:          fromPgTimestamptz(p.DeletedAt),
	}
}

// mapSliceToDomain converts a slice of sqlc Positions to a slice of domain Positions.
func (r *PositionRepository) mapSliceToDomain(list []sqlc.Position) []domain.Position {
	res := make([]domain.Position, len(list))
	for i, p := range list {
		res[i] = *r.mapToDomain(p)
	}
	return res
}
