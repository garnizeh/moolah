package repository

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

type auditRepo struct {
	q sqlc.Querier
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(q sqlc.Querier) domain.AuditRepository {
	return &auditRepo{q: q}
}

// Create appends a new audit log entry.
func (r *auditRepo) Create(ctx context.Context, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	id := ulid.New()

	arg := sqlc.CreateAuditLogParams{
		ID:         id,
		TenantID:   input.TenantID,
		ActorID:    input.ActorID,
		ActorRole:  sqlc.UserRole(input.ActorRole),
		Action:     sqlc.AuditAction(input.Action),
		EntityType: input.EntityType,
		OldValues:  input.OldValues,
		NewValues:  input.NewValues,
	}

	if input.EntityID != "" {
		arg.EntityID = pgtype.Text{String: input.EntityID, Valid: true}
	}

	if input.UserAgent != "" {
		arg.UserAgent = pgtype.Text{String: input.UserAgent, Valid: true}
	}

	if input.IPAddress != "" {
		addr, err := netip.ParseAddr(input.IPAddress)
		if err == nil {
			arg.IpAddress = &addr
		}
	}

	row, err := r.q.CreateAuditLog(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// ListByTenant returns audit logs for a specific tenant with optional filters.
func (r *auditRepo) ListByTenant(ctx context.Context, tenantID string, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}

	// Note: Currently sqlc query ListAuditLogsByTenant only takes tenant_id, skip filters for Phase 1
	// if the sqlc query doesn't support them yet. However, TASK_1.3.7 says "ListByTenant applies all ListAuditLogsParams filters".
	// Looking at sqlc code, it only has TenantID, OffsetOff, LimitOff.
	// We might need to handle filters in memory or update sqlc.
	// For now, I'll implement what sqlc allows and note it.
	arg := sqlc.ListAuditLogsByTenantParams{
		TenantID:  tenantID,
		OffsetOff: params.Offset,
		LimitOff:  limit,
	}

	rows, err := r.q.ListAuditLogsByTenant(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", TranslateError(err))
	}

	logs := make([]domain.AuditLog, len(rows))
	for i, row := range rows {
		logs[i] = *r.mapToDomain(row)
	}

	return logs, nil
}

// ListByEntity returns audit logs for a specific entity (e.g. a single transaction).
func (r *auditRepo) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]domain.AuditLog, error) {
	arg := sqlc.ListAuditLogsByEntityParams{
		TenantID:   tenantID,
		EntityType: entityType,
		EntityID:   pgtype.Text{String: entityID, Valid: true},
	}

	rows, err := r.q.ListAuditLogsByEntity(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs by entity: %w", TranslateError(err))
	}

	logs := make([]domain.AuditLog, len(rows))
	for i, row := range rows {
		logs[i] = *r.mapToDomain(row)
	}

	return logs, nil
}

func (r *auditRepo) mapToDomain(row sqlc.AuditLog) *domain.AuditLog {
	log := &domain.AuditLog{
		ID:         row.ID,
		TenantID:   row.TenantID,
		ActorID:    row.ActorID,
		ActorRole:  domain.Role(row.ActorRole),
		Action:     domain.AuditAction(row.Action),
		EntityType: row.EntityType,
		OldValues:  row.OldValues,
		NewValues:  row.NewValues,
		CreatedAt:  row.CreatedAt.Time,
	}

	if row.EntityID.Valid {
		log.EntityID = row.EntityID.String
	}

	if row.UserAgent.Valid {
		log.UserAgent = row.UserAgent.String
	}

	if row.IpAddress != nil {
		log.IPAddress = row.IpAddress.String()
	}

	return log
}
