package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type adminTenantRepo struct {
	q sqlc.Querier
}

// NewAdminTenantRepository creates a new concrete implementation of domain.AdminTenantRepository.
func NewAdminTenantRepository(q sqlc.Querier) domain.AdminTenantRepository {
	return &adminTenantRepo{q: q}
}

func (r *adminTenantRepo) ListAll(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	tenants, err := r.q.AdminListAllTenants(ctx, withDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to list all tenants: %w", err)
	}

	result := make([]domain.Tenant, len(tenants))
	for i, t := range tenants {
		result[i] = *mapTenant(t)
	}
	return result, nil
}

func (r *adminTenantRepo) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t, err := r.q.AdminGetTenantByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by id: %w", err)
	}
	return mapTenant(t), nil
}

func (r *adminTenantRepo) UpdatePlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	arg := sqlc.AdminUpdateTenantPlanParams{
		ID:   id,
		Plan: sqlc.TenantPlan(plan),
	}
	t, err := r.q.AdminUpdateTenantPlan(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update tenant plan: %w", err)
	}
	return mapTenant(t), nil
}

func (r *adminTenantRepo) Suspend(ctx context.Context, id string) error {
	err := r.q.AdminSuspendTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to suspend tenant: %w", err)
	}
	return nil
}

func (r *adminTenantRepo) Restore(ctx context.Context, id string) error {
	err := r.q.AdminRestoreTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore tenant: %w", err)
	}
	return nil
}

func (r *adminTenantRepo) HardDelete(ctx context.Context, id string) error {
	err := r.q.AdminHardDeleteTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete tenant: %w", err)
	}
	return nil
}

type adminUserRepo struct {
	q sqlc.Querier
}

// NewAdminUserRepository creates a new concrete implementation of domain.AdminUserRepository.
func NewAdminUserRepository(q sqlc.Querier) domain.AdminUserRepository {
	return &adminUserRepo{q: q}
}

func (r *adminUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	users, err := r.q.AdminListAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all users: %w", err)
	}

	result := make([]domain.User, len(users))
	for i, u := range users {
		result[i] = *mapUser(u)
	}
	return result, nil
}

func (r *adminUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := r.q.AdminGetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return mapUser(u), nil
}

func (r *adminUserRepo) ForceDelete(ctx context.Context, id string) error {
	err := r.q.AdminForceDeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to force delete user: %w", err)
	}
	return nil
}

type adminAuditRepo struct {
	q sqlc.Querier
}

// NewAdminAuditRepository creates a new concrete implementation of domain.AdminAuditRepository.
func NewAdminAuditRepository(q sqlc.Querier) domain.AdminAuditRepository {
	return &adminAuditRepo{q: q}
}

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
		return nil, fmt.Errorf("failed to list all audit logs: %w", err)
	}

	result := make([]domain.AuditLog, len(logs))
	for i, l := range logs {
		result[i] = *mapAuditLog(l)
	}
	return result, nil
}

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
