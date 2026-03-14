package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
)

// tenantService provides business logic for managing tenants, including creation, retrieval, updating, and deletion of tenants, as well as inviting users to a tenant. It also handles related audit logging.
type tenantService struct {
	tenantRepo domain.TenantRepository
	userRepo   domain.UserRepository
	auditRepo  domain.AuditRepository
}

// NewTenantService creates a new instance of TenantService.
func NewTenantService(
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
	auditRepo domain.AuditRepository,
) domain.TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		auditRepo:  auditRepo,
	}
}

// Create creates a new tenant with the specified input, and logs the creation action in the audit trail.
func (s *tenantService) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	newValues, err := json.Marshal(map[string]any{"name": tenant.Name, "plan": tenant.Plan})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tenant audit: %w", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenant.ID,
		ActorID:    "system", // Tenant creation is often a system/sysadmin action
		Action:     domain.AuditActionCreate,
		EntityType: "tenant",
		EntityID:   tenant.ID,
		ActorRole:  domain.RoleAdmin, // Setting a default role for the audit log
		NewValues:  newValues,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return tenant, nil
}

// GetByID retrieves a tenant by its ID. It returns the tenant or an error if the record cannot be found or retrieved.
func (s *tenantService) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// List returns all tenants. It retrieves the list of tenants or returns an error if the records cannot be retrieved.
func (s *tenantService) List(ctx context.Context) ([]domain.Tenant, error) {
	tenants, err := s.tenantRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// Update validates the input and updates an existing tenant record. It logs any changes in the tenant's name or plan in the audit trail. It returns the updated tenant or an error if validation fails or the record cannot be updated.
func (s *tenantService) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	oldTenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing tenant: %w", err)
	}

	tenant, err := s.tenantRepo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	newValuesMap := make(map[string]any)
	oldValuesMap := make(map[string]any)

	if input.Name != nil && *input.Name != oldTenant.Name {
		newValuesMap["name"] = *input.Name
		oldValuesMap["name"] = oldTenant.Name
	}
	if input.Plan != nil && *input.Plan != oldTenant.Plan {
		newValuesMap["plan"] = *input.Plan
		oldValuesMap["plan"] = oldTenant.Plan
	}

	if len(newValuesMap) > 0 {
		oldValues, err := json.Marshal(oldValuesMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tenant audit (old): %w", err)
		}
		newValues, err := json.Marshal(newValuesMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tenant audit (new): %w", err)
		}

		_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
			TenantID:   id,
			ActorID:    "system",
			Action:     domain.AuditActionUpdate,
			EntityType: "tenant",
			EntityID:   id,
			ActorRole:  domain.RoleAdmin,
			OldValues:  oldValues,
			NewValues:  newValues,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create audit log: %w", err)
		}
	}

	return tenant, nil
}

// Delete performs a soft delete of a tenant record by its ID. It logs the deletion action in the audit trail. It returns an error if the record cannot be deleted.
func (s *tenantService) Delete(ctx context.Context, id string) error {
	_, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   id,
		ActorID:    "system",
		Action:     domain.AuditActionSoftDelete,
		EntityType: "tenant",
		EntityID:   id,
		ActorRole:  domain.RoleAdmin,
	})
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// InviteUser invites a new user to the tenant by creating a user record with the specified input. It logs the invitation action in the audit trail. It returns the created user or an error if the record cannot be created.
func (s *tenantService) InviteUser(ctx context.Context, tenantID string, input domain.CreateUserInput) (*domain.User, error) {
	// Enforcement: logic requires tenant match.
	input.TenantID = tenantID

	user, err := s.userRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to invite user: %w", err)
	}

	newValues, err := json.Marshal(map[string]any{"email": user.Email, "role": user.Role})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user audit: %w", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    "system", // In a real scenario, this would be the admin user's ID from context
		Action:     domain.AuditActionCreate,
		EntityType: "user",
		EntityID:   user.ID,
		ActorRole:  domain.RoleAdmin,
		NewValues:  newValues,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return user, nil
}
