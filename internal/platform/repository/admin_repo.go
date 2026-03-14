package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// AdminRepository provides methods for administrative operations on tenants, users, and audit logs.
// It allows listing, retrieving, updating, suspending, restoring, and hard-deleting tenants and users,
// as well as listing audit logs with various filters.
type adminTenantRepo struct {
	q sqlc.Querier
}

// NewAdminTenantRepository creates a new concrete implementation of domain.AdminTenantRepository.
func NewAdminTenantRepository(q sqlc.Querier) domain.AdminTenantRepository {
	return &adminTenantRepo{q: q}
}

// ListAll returns all tenants, including suspended ones. If withDeleted is true, it also includes hard-deleted tenants.
func (r *adminTenantRepo) ListAll(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	tenants, err := r.q.AdminListAllTenants(ctx, withDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to list all tenants: %w", TranslateError(err))
	}

	result := make([]domain.Tenant, len(tenants))
	for i, t := range tenants {
		result[i] = *mapTenant(t)
	}
	return result, nil
}

// GetByID retrieves a tenant by its ID, regardless of its status (active, suspended, or deleted).
func (r *adminTenantRepo) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t, err := r.q.AdminGetTenantByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by id: %w", TranslateError(err))
	}
	return mapTenant(t), nil
}

// UpdatePlan changes the subscription plan of a tenant. It can be used to upgrade, downgrade, or set a custom plan.
func (r *adminTenantRepo) UpdatePlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	arg := sqlc.AdminUpdateTenantPlanParams{
		ID:   id,
		Plan: sqlc.TenantPlan(plan),
	}
	t, err := r.q.AdminUpdateTenantPlan(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant plan: %w", TranslateError(err))
	}
	return mapTenant(t), nil
}

// Suspend marks a tenant as suspended, which typically disables access to the application for all users under that tenant. Suspended tenants can be restored later.
func (r *adminTenantRepo) Suspend(ctx context.Context, id string) error {
	err := r.q.AdminSuspendTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to suspend tenant: %w", TranslateError(err))
	}
	return nil
}

// Restore reactivates a suspended tenant, allowing users under that tenant to access the application again. It cannot restore hard-deleted tenants.
func (r *adminTenantRepo) Restore(ctx context.Context, id string) error {
	err := r.q.AdminRestoreTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore tenant: %w", TranslateError(err))
	}
	return nil
}

// HardDelete permanently removes a tenant from the database. This action is irreversible and should be used with caution, typically only for tenants that were previously soft-deleted.
func (r *adminTenantRepo) HardDelete(ctx context.Context, id string) error {
	err := r.q.AdminHardDeleteTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete tenant: %w", TranslateError(err))
	}
	return nil
}

// mapTenant converts a sqlc.Tenant to a domain.Tenant, handling nullable fields appropriately.
type adminUserRepo struct {
	q sqlc.Querier
}

// NewAdminUserRepository creates a new concrete implementation of domain.AdminUserRepository.
func NewAdminUserRepository(q sqlc.Querier) domain.AdminUserRepository {
	return &adminUserRepo{q: q}
}

// ListAll returns all users across all tenants, including those that are soft-deleted. It does not return hard-deleted users.
func (r *adminUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	users, err := r.q.AdminListAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all users: %w", TranslateError(err))
	}

	result := make([]domain.User, len(users))
	for i, u := range users {
		result[i] = *mapUser(u)
	}
	return result, nil
}

// GetByID retrieves a user by their ID, regardless of their status (active or soft-deleted). It does not return hard-deleted users.
func (r *adminUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := r.q.AdminGetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", TranslateError(err))
	}
	return mapUser(u), nil
}

// ForceDelete permanently removes a user from the database. This action is irreversible and should be used with caution, typically only for users that were previously soft-deleted.
func (r *adminUserRepo) ForceDelete(ctx context.Context, id string) error {
	err := r.q.AdminForceDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to force delete user: %w", TranslateError(err))
	}
	return nil
}

// mapUser converts a sqlc.User to a domain.User, handling nullable fields appropriately.
type adminAuditRepo struct {
	q sqlc.Querier
}

// NewAdminAuditRepository creates a new concrete implementation of domain.AdminAuditRepository.
func NewAdminAuditRepository(q sqlc.Querier) domain.AdminAuditRepository {
	return &adminAuditRepo{q: q}
}

// ListAll returns all audit logs across all tenants, with optional filtering by actor ID, entity type, entity ID, action, and date range. It supports pagination through limit and offset parameters.
func (r *adminAuditRepo) ListAll(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	arg := sqlc.AdminListAllAuditLogsParams{
		LimitOff:  params.Limit,
		OffsetOff: params.Offset,
		ActorID:   params.ActorID,
	}

	if params.EntityType != "" {
		arg.EntityType = pgtype.Text{String: params.EntityType, Valid: true}
	}
	if params.EntityID != "" {
		arg.EntityID = pgtype.Text{String: params.EntityID, Valid: true}
	}
	if params.Action != "" {
		arg.Action = sqlc.NullAuditAction{AuditAction: sqlc.AuditAction(params.Action), Valid: true}
	}
	if params.StartDate != nil {
		arg.StartDate = pgtype.Timestamptz{Time: *params.StartDate, Valid: true}
	}
	if params.EndDate != nil {
		arg.EndDate = pgtype.Timestamptz{Time: *params.EndDate, Valid: true}
	}

	logs, err := r.q.AdminListAllAuditLogs(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list all audit logs: %w", TranslateError(err))
	}

	result := make([]domain.AuditLog, len(logs))
	for i, l := range logs {
		result[i] = *mapAuditLog(l)
	}
	return result, nil
}

// mapTenant converts a sqlc.Tenant to a domain.Tenant, handling nullable fields appropriately.
func mapTenant(t sqlc.Tenant) *domain.Tenant {
	var deletedAt *time.Time
	if t.DeletedAt.Valid {
		deletedAt = &t.DeletedAt.Time
	}

	return &domain.Tenant{
		ID:        t.ID,
		Name:      t.Name,
		Plan:      domain.TenantPlan(t.Plan),
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
		DeletedAt: deletedAt,
	}
}

// mapUser converts a sqlc.User to a domain.User, handling nullable fields appropriately.
func mapUser(u sqlc.User) *domain.User {
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		deletedAt = &u.DeletedAt.Time
	}

	var lastLoginAt *time.Time
	if u.LastLoginAt.Valid {
		lastLoginAt = &u.LastLoginAt.Time
	}

	return &domain.User{
		ID:          u.ID,
		TenantID:    u.TenantID,
		Email:       u.Email,
		Name:        u.Name,
		Role:        domain.Role(u.Role),
		CreatedAt:   u.CreatedAt.Time,
		UpdatedAt:   u.UpdatedAt.Time,
		DeletedAt:   deletedAt,
		LastLoginAt: lastLoginAt,
	}
}

// mapAuditLog converts a sqlc.AuditLog to a domain.AuditLog, handling nullable fields appropriately.
func mapAuditLog(l sqlc.AuditLog) *domain.AuditLog {
	var ip string
	if l.IpAddress != nil {
		ip = l.IpAddress.String()
	}

	return &domain.AuditLog{
		ID:         l.ID,
		TenantID:   l.TenantID,
		ActorID:    l.ActorID,
		ActorRole:  domain.Role(l.ActorRole),
		Action:     domain.AuditAction(l.Action),
		EntityType: l.EntityType,
		EntityID:   l.EntityID.String,
		OldValues:  l.OldValues,
		NewValues:  l.NewValues,
		IPAddress:  ip,
		UserAgent:  l.UserAgent.String,
		CreatedAt:  l.CreatedAt.Time,
	}
}
