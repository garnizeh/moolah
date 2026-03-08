package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/garnizeh/moolah/internal/domain"
)

type adminService struct {
	tenantRepo domain.AdminTenantRepository
	userRepo   domain.AdminUserRepository
	adminAudit domain.AdminAuditRepository
	auditRepo  domain.AuditRepository
	logger     *slog.Logger
}

// NewAdminService creates a new system-wide administrative service.
func NewAdminService(
	tenantRepo domain.AdminTenantRepository,
	userRepo domain.AdminUserRepository,
	adminAudit domain.AdminAuditRepository,
	auditRepo domain.AuditRepository,
	logger *slog.Logger,
) domain.AdminService {
	return &adminService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		adminAudit: adminAudit,
		auditRepo:  auditRepo,
		logger:     logger,
	}
}

func (s *adminService) ListAllTenants(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	tenants, err := s.tenantRepo.ListAll(ctx, withDeleted)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to list all tenants: %w", err)
	}
	return tenants, nil
}

func (s *adminService) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to get tenant by id: %w", err)
	}
	if tenant == nil {
		return nil, domain.ErrNotFound
	}
	return tenant, nil
}

func (s *adminService) UpdateTenantPlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.UpdatePlan(ctx, id, plan)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to update tenant plan: %w", err)
	}

	// Global audit log.
	_, auditErr := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   id,
		ActorID:    "SYSTEM",
		EntityType: "tenant",
		EntityID:   id,
		Action:     domain.AuditActionUpdate,
	})
	if auditErr != nil {
		s.logger.ErrorContext(ctx, "failed to create audit log for tenant plan update", "error", auditErr)
	}

	return tenant, nil
}

func (s *adminService) SuspendTenant(ctx context.Context, id string) error {
	err := s.tenantRepo.Suspend(ctx, id)
	if err != nil {
		return fmt.Errorf("admin service: failed to suspend tenant: %w", err)
	}

	_, auditErr := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   id,
		ActorID:    "SYSTEM",
		EntityType: "tenant",
		EntityID:   id,
		Action:     domain.AuditActionSoftDelete,
	})
	if auditErr != nil {
		s.logger.ErrorContext(ctx, "failed to create audit log for tenant suspension", "error", auditErr)
	}

	return nil
}

func (s *adminService) RestoreTenant(ctx context.Context, id string) error {
	err := s.tenantRepo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("admin service: failed to restore tenant: %w", err)
	}

	_, auditErr := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   id,
		ActorID:    "SYSTEM",
		EntityType: "tenant",
		EntityID:   id,
		Action:     domain.AuditActionRestore,
	})
	if auditErr != nil {
		s.logger.ErrorContext(ctx, "failed to create audit log for tenant restoration", "error", auditErr)
	}

	return nil
}

func (s *adminService) HardDeleteTenant(ctx context.Context, id, confirmationToken string) error {
	if confirmationToken != id {
		return domain.ErrInvalidInput
	}

	// Write audit before hard deletion since the tenant ID won't exist soon.
	_, auditErr := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   id,
		ActorID:    "SYSTEM",
		EntityType: "tenant",
		EntityID:   id,
		Action:     domain.AuditActionSoftDelete, // Record intent
	})
	if auditErr != nil {
		s.logger.ErrorContext(ctx, "failed to create audit log for tenant hard deletion", "error", auditErr)
	}

	err := s.tenantRepo.HardDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("admin service: failed to hard delete tenant: %w", err)
	}

	return nil
}

func (s *adminService) ListAllUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.userRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to list all users: %w", err)
	}
	return users, nil
}

func (s *adminService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to get user by id: %w", err)
	}
	if user == nil {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (s *adminService) ForceDeleteUser(ctx context.Context, id string) error {
	// Write audit record.
	_, auditErr := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   "SYSTEM", // Global context
		ActorID:    "SYSTEM",
		EntityType: "user",
		EntityID:   id,
		Action:     domain.AuditActionSoftDelete,
	})
	if auditErr != nil {
		s.logger.ErrorContext(ctx, "failed to create audit log for user force delete", "error", auditErr)
	}

	err := s.userRepo.ForceDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("admin service: failed to force delete user: %w", err)
	}

	return nil
}

func (s *adminService) ListAuditLogs(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	logs, err := s.adminAudit.ListAll(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("admin service: failed to list all audit logs: %w", err)
	}
	return logs, nil
}
