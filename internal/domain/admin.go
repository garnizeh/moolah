package domain

import (
	"context"
)

// AdminTenantRepository defines cross-tenant operations for sysadmin use only.
// These methods bypass the standard tenant_id isolation for system-wide management.
type AdminTenantRepository interface {
	// ListAll returns every tenant in the system, including soft-deleted ones when withDeleted is true.
	ListAll(ctx context.Context, withDeleted bool) ([]Tenant, error)

	// GetByID retrieves a tenant without a tenant_id guard (sysadmin bypass).
	GetByID(ctx context.Context, id string) (*Tenant, error)

	// UpdatePlan changes a tenant's subscription plan.
	UpdatePlan(ctx context.Context, id string, plan TenantPlan) (*Tenant, error)

	// Suspend soft-deletes a tenant, blocking all logins for its users.
	Suspend(ctx context.Context, id string) error

	// Restore reverses a soft-delete on a tenant.
	Restore(ctx context.Context, id string) error

	// HardDelete permanently removes a tenant and all associated data.
	// WARNING: This is irreversible and will delete all household data via cascade.
	// Must only be called after explicit confirmation.
	HardDelete(ctx context.Context, id string) error
}

// AdminUserRepository defines cross-tenant user operations for sysadmin use only.
// These methods allow management of users across the entire system.
type AdminUserRepository interface {
	// ListAll returns every user in the system regardless of tenant.
	ListAll(ctx context.Context) ([]User, error)

	// GetByID retrieves a user without a tenant_id guard.
	GetByID(ctx context.Context, id string) (*User, error)

	// ForceDelete hard-deletes a user record.
	// WARNING: This is irreversible and should be used with extreme caution.
	ForceDelete(ctx context.Context, id string) error
}

// AdminAuditRepository defines global audit log queries without tenant isolation.
// This is used for system-wide compliance and debugging.
type AdminAuditRepository interface {
	// ListAll returns audit logs across all tenants with optional filters.
	ListAll(ctx context.Context, params ListAuditLogsParams) ([]AuditLog, error)
}

// AdminService defines the business-logic contract for cross-tenant sysadmin operations.
type AdminService interface {
	// ListAllTenants returns all tenants in the system, with an option to include soft-deleted ones.
	ListAllTenants(ctx context.Context, withDeleted bool) ([]Tenant, error)

	// GetTenantByID retrieves a tenant by its ID without tenant isolation.
	GetTenantByID(ctx context.Context, id string) (*Tenant, error)

	// UpdateTenantPlan changes the subscription plan for a tenant.
	UpdateTenantPlan(ctx context.Context, id string, plan TenantPlan) (*Tenant, error)

	// SuspendTenant performs a soft-delete on the tenant, blocking all user access.
	SuspendTenant(ctx context.Context, id string) error

	// RestoreTenant reverses a soft-delete on the tenant, restoring user access.
	RestoreTenant(ctx context.Context, id string) error

	// HardDeleteTenant permanently deletes a tenant and all associated data after confirmation.
	HardDeleteTenant(ctx context.Context, id, confirmationToken string) error

	// ListAllUsers returns all users across the system without tenant isolation.
	ListAllUsers(ctx context.Context) ([]User, error)

	// GetUserByID retrieves a user by ID without tenant isolation.
	GetUserByID(ctx context.Context, id string) (*User, error)

	// ForceDeleteUser permanently deletes a user record without tenant isolation.
	ForceDeleteUser(ctx context.Context, id string) error

	// ListAuditLogs returns audit logs across all tenants with optional filters, without tenant isolation.
	ListAuditLogs(ctx context.Context, params ListAuditLogsParams) ([]AuditLog, error)
}
