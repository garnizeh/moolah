package domain

import (
	"context"
	"time"
)

// User represents a person belonging to a Tenant (household).
// All operational field accesses must be isolated by TenantID.
type User struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	LastLoginAt *time.Time
	ID          string
	TenantID    string
	Email       string
	Name        string
	Role        Role
}

// CreateUserInput holds the data required to enroll a new user.
type CreateUserInput struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Role     Role   `json:"role" validate:"required"`
}

// UpdateUserInput holds the data that can be updated for a user.
type UpdateUserInput struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Role *Role   `json:"role,omitempty" validate:"omitempty"`
}

// UserRepository defines the persistence contract for User entities.
// Every method (except GetByEmail for login) MUST enforce TenantID isolation.
type UserRepository interface {
	// Create persists a new user within a tenant.
	Create(ctx context.Context, input CreateUserInput) (*User, error)

	// GetByID retrieves a user by ID, scoped to a specific tenant.
	GetByID(ctx context.Context, tenantID, id string) (*User, error)

	// GetByEmail retrieves a user by their unique email across all tenants.
	// Used primarily during the authentication flow.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// ListByTenant returns all active users belonging to a household.
	ListByTenant(ctx context.Context, tenantID string) ([]User, error)

	// Update modifies a user's attributes within their tenant.
	Update(ctx context.Context, tenantID, id string, input UpdateUserInput) (*User, error)

	// UpdateLastLogin updates the login timestamp for a user.
	UpdateLastLogin(ctx context.Context, id string) error

	// Delete performs a soft-delete on a user.
	Delete(ctx context.Context, tenantID, id string) error
}
