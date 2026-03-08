package domain

import (
	"context"
	"time"
)

// TenantPlan represents the subscription tier of a household.
type TenantPlan string

const (
	// TenantPlanFree is the default restricted tier.
	TenantPlanFree TenantPlan = "free"
	// TenantPlanBasic is the standard paid tier.
	TenantPlanBasic TenantPlan = "basic"
	// TenantPlanPremium is the highest tier with all features.
	TenantPlanPremium TenantPlan = "premium"
)

// Tenant represents a household, the root of the multi-tenancy hierarchy.
type Tenant struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	ID        string
	Name      string
	Plan      TenantPlan
}

// CreateTenantInput holds the data required to create a new household.
type CreateTenantInput struct {
	Name string `validate:"required,min=2,max=100"`
}

// UpdateTenantInput holds the data required to update an existing household.
type UpdateTenantInput struct {
	Name *string     `validate:"omitempty,min=2,max=100"`
	Plan *TenantPlan `validate:"omitempty"`
}

// TenantRepository defines the persistence contract for Tenant entities.
// These operations are typically restricted to the sysadmin role.
type TenantRepository interface {
	// Create persists a new tenant.
	Create(ctx context.Context, input CreateTenantInput) (*Tenant, error)

	// GetByID retrieves a tenant by its unique identifier.
	GetByID(ctx context.Context, id string) (*Tenant, error)

	// List returns all active (non-deleted) tenants.
	List(ctx context.Context) ([]Tenant, error)

	// Update modifies an existing tenant's attributes.
	Update(ctx context.Context, id string, input UpdateTenantInput) (*Tenant, error)

	// Delete performs a soft-delete on the tenant.
	Delete(ctx context.Context, id string) error
}

// TenantService defines the business-logic contract for tenant management.
type TenantService interface {
	// Create persists a new tenant and records an audit log.
	Create(ctx context.Context, input CreateTenantInput) (*Tenant, error)

	// GetByID retrieves a tenant by its unique identifier.
	GetByID(ctx context.Context, id string) (*Tenant, error)

	// List returns all active (non-deleted) tenants. Restricted to sysadmins.
	List(ctx context.Context) ([]Tenant, error)

	// Update modifies an existing tenant's attributes and records an audit log.
	Update(ctx context.Context, id string, input UpdateTenantInput) (*Tenant, error)

	// Delete performs a soft-delete on the tenant and records an audit log.
	Delete(ctx context.Context, id string) error

	// InviteUser creates a new user within the context of a tenant.
	InviteUser(ctx context.Context, tenantID string, input CreateUserInput) (*User, error)
}
