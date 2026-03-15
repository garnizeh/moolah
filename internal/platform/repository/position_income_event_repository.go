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

var _ domain.PositionIncomeEventRepository = (*PositionIncomeEventRepository)(nil)

// PositionIncomeEventRepository handles receivables.
type PositionIncomeEventRepository struct {
	q sqlc.Querier
}

// NewPositionIncomeEventRepository returns a new PositionIncomeEventRepository instance.
func NewPositionIncomeEventRepository(q sqlc.Querier) *PositionIncomeEventRepository {
	return &PositionIncomeEventRepository{q: q}
}

// Create persists a new income event.
func (r *PositionIncomeEventRepository) Create(ctx context.Context, tenantID string, in domain.CreatePositionIncomeEventInput) (*domain.PositionIncomeEvent, error) {
	arg := sqlc.CreatePositionIncomeEventParams{
		ID:          ulid.New(),
		TenantID:    tenantID,
		PositionID:  in.PositionID,
		IncomeType:  sqlc.IncomeType(in.IncomeType),
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		EventDate:   pgtype.Date{Time: in.DueAt, Valid: true},
		Status:      sqlc.ReceivableStatusPending,
	}

	event, err := r.q.CreatePositionIncomeEvent(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create income event: %w", err)
	}

	return r.mapToDomain(event), nil
}

// GetByID retrieves an income event by ID.
func (r *PositionIncomeEventRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.PositionIncomeEvent, error) {
	event, err := r.q.GetPositionIncomeEventByID(ctx, sqlc.GetPositionIncomeEventByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get income event: %w", err)
	}

	return r.mapToDomain(event), nil
}

// ListByTenant lists all events for a tenant.
func (r *PositionIncomeEventRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.PositionIncomeEvent, error) {
	list, err := r.q.ListPositionIncomeEventsByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant income events: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// ListPending lists pending events.
func (r *PositionIncomeEventRepository) ListPending(ctx context.Context, tenantID string) ([]domain.PositionIncomeEvent, error) {
	list, err := r.q.ListPendingIncomeEvents(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending income events: %w", err)
	}

	return r.mapSliceToDomain(list), nil
}

// UpdateStatus transitions status.
func (r *PositionIncomeEventRepository) UpdateStatus(
	ctx context.Context,
	tenantID, id string,
	status domain.ReceivableStatus,
	receivedAt *time.Time,
) (*domain.PositionIncomeEvent, error) {
	// 1. Get current status to enforce state machine
	curr, err := r.q.GetPositionIncomeEventByID(ctx, sqlc.GetPositionIncomeEventByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find income event for status update: %w", err)
	}

	// 2. Enforce transition: only from pending
	if curr.Status != sqlc.ReceivableStatusPending {
		return nil, domain.ErrConflict
	}

	// 3. Update status
	event, err := r.q.UpdateIncomeEventStatus(ctx, sqlc.UpdateIncomeEventStatusParams{
		ID:         id,
		TenantID:   tenantID,
		Status:     sqlc.ReceivableStatus(status),
		RealizedAt: pgtype.Timestamptz{Time: safeTime(receivedAt), Valid: receivedAt != nil},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update income event status: %w", err)
	}

	return r.mapToDomain(event), nil
}

// mapToDomain converts a sqlc PositionIncomeEvent to a domain PositionIncomeEvent.
func (r *PositionIncomeEventRepository) mapToDomain(e sqlc.PositionIncomeEvent) *domain.PositionIncomeEvent {
	return &domain.PositionIncomeEvent{
		ID:          e.ID,
		TenantID:    e.TenantID,
		PositionID:  e.PositionID,
		IncomeType:  domain.IncomeType(e.IncomeType),
		AmountCents: e.AmountCents,
		Currency:    e.Currency,
		DueAt:       e.EventDate.Time,
		Status:      domain.ReceivableStatus(e.Status),
		ReceivedAt:  fromPgTimestamptz(e.RealizedAt),
		CreatedAt:   e.CreatedAt.Time,
	}
}

// mapSliceToDomain converts a slice of sqlc PositionIncomeEvents to a slice of domain PositionIncomeEvents.
func (r *PositionIncomeEventRepository) mapSliceToDomain(list []sqlc.PositionIncomeEvent) []domain.PositionIncomeEvent {
	res := make([]domain.PositionIncomeEvent, len(list))
	for i, e := range list {
		res[i] = *r.mapToDomain(e)
	}
	return res
}
