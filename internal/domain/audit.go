package domain

import (
	"context"
	"time"
)

// AuditAction mirrors the database enum for audit actions.
type AuditAction string

const (
	// AuditActionCreate representing many create operations.
	AuditActionCreate AuditAction = "create"
	// AuditActionUpdate representing many update operations.
	AuditActionUpdate AuditAction = "update"
	// AuditActionSoftDelete representing soft delete operations.
	AuditActionSoftDelete AuditAction = "soft_delete"
	// AuditActionRestore representing restore operations.
	AuditActionRestore AuditAction = "restore"
	// AuditActionLogin representing login operations.
	AuditActionLogin AuditAction = "login"
	// AuditActionLoginFailed representing login failure operations.
	AuditActionLoginFailed AuditAction = "login_failed"
	// AuditActionOTPRequested representing OTP request operations.
	AuditActionOTPRequested AuditAction = "otp_requested"
	// AuditActionOTPVerified representing OTP verification operations.
	AuditActionOTPVerified AuditAction = "otp_verified"
)

// AuditLog is an immutable record of a significant system event.
// It provides a tamper-evident trace of actor actions, IPs, and state changes.
type AuditLog struct {
	CreatedAt  time.Time   `json:"created_at"`
	ID         string      `json:"id"`
	TenantID   string      `json:"tenant_id"`
	ActorID    string      `json:"actor_id"` // User ULID or "SYSTEM" for automated actions
	EntityType string      `json:"entity_type"`
	EntityID   string      `json:"entity_id"`
	IPAddress  string      `json:"ip_address"`
	UserAgent  string      `json:"user_agent"`
	ActorRole  Role        `json:"actor_role"`
	Action     AuditAction `json:"action"`
	OldValues  []byte      `json:"old_values,omitempty"` // JSON snapshot before change
	NewValues  []byte      `json:"new_values,omitempty"` // JSON snapshot after change
}

// CreateAuditLogInput is the value object used for recording a new audit event.
type CreateAuditLogInput struct {
	TenantID   string      `validate:"required"`
	ActorID    string      `validate:"required"`
	Action     AuditAction `validate:"required"`
	EntityType string      `validate:"required"`
	EntityID   string      `validate:"omitempty"`
	IPAddress  string      `validate:"omitempty"`
	UserAgent  string      `validate:"omitempty"`
	ActorRole  Role        `validate:"required"`
	OldValues  []byte      `json:"-"`
	NewValues  []byte      `json:"-"`
}

// ListAuditLogsParams defines the filters and pagination for querying audit trails.
type ListAuditLogsParams struct {
	StartDate  *time.Time  `json:"start_date,omitempty"`
	EndDate    *time.Time  `json:"end_date,omitempty"`
	EntityType string      `json:"entity_type,omitempty"`
	EntityID   string      `json:"entity_id,omitempty"`
	ActorID    string      `json:"actor_id,omitempty"`
	Action     AuditAction `json:"action,omitempty"`
	Limit      int32       `json:"limit"`
	Offset     int32       `json:"offset"`
}

// AuditRepository defines persistence operations for audit logs.
// Audit logs are append-only — no update or delete methods are provided.
type AuditRepository interface {
	// Create appends a new audit log entry.
	Create(ctx context.Context, input CreateAuditLogInput) (*AuditLog, error)

	// ListByTenant returns audit logs for a specific tenant with optional filters.
	ListByTenant(ctx context.Context, tenantID string, params ListAuditLogsParams) ([]AuditLog, error)

	// ListByEntity returns audit logs for a specific entity (e.g. a single transaction).
	ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]AuditLog, error)
}
