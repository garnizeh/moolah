# Task 1.2.10 — Domain AuditLog Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `AuditLog` domain entity and `AuditRepository` interface in `internal/domain/audit.go`. Audit logs provide an immutable trace of all create, update, delete, and authentication events, which is essential for compliance, debugging, and dispute resolution in a financial application.

---

## 2. Context & Motivation

Financial applications require a tamper-evident audit trail. Every significant action (transaction created, account deleted, OTP verified) must be recorded with actor identity, timestamps, IP address, and before/after state. The `AuditRepository` interface enables the service layer to write audit entries without coupling to the database.

---

## 3. Scope

### In scope

- [x] `AuditLog` struct and `AuditAction` constants.
- [x] `CreateAuditLogInput` value object.
- [x] `AuditRepository` interface: `Create` + `ListByTenant` + `ListByEntity`.
- [x] Audit logs are **append-only** — no update or delete methods on the repository.

### Out of scope

- Concrete repository implementation (Task 1.3.7).
- Service-layer helpers for emitting audit events (Task 1.4.5).
- HTTP handler for querying audit logs (Task 1.5.9).
- Long-term archival / purge strategy (Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                       | Purpose                              |
| ------ | -------------------------- | ------------------------------------ |
| CREATE | `internal/domain/audit.go` | Entity, input types, repo interface  |

### Key interfaces / types

```go
// AuditAction mirrors the database enum for audit actions.
type AuditAction string

const (
    AuditActionCreate       AuditAction = "create"
    AuditActionUpdate       AuditAction = "update"
    AuditActionSoftDelete   AuditAction = "soft_delete"
    AuditActionRestore      AuditAction = "restore"
    AuditActionLogin        AuditAction = "login"
    AuditActionLoginFailed  AuditAction = "login_failed"
    AuditActionOTPRequested AuditAction = "otp_requested"
    AuditActionOTPVerified  AuditAction = "otp_verified"
)

// AuditLog is an immutable record of a significant system event.
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

type ListAuditLogsParams struct {
    EntityType string
    EntityID   string
    ActorID    string
    Action     AuditAction
    StartDate  *time.Time
    EndDate    *time.Time
    Limit      int32
    Offset     int32
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
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/audit_logs.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.9 (admin handler).

### Error cases to handle

| Scenario              | Sentinel Error        | HTTP Status |
| --------------------- | --------------------- | ----------- |
| Audit log not found   | `domain.ErrNotFound`  | `404`       |
| Tenant mismatch       | `domain.ErrForbidden` | `403`       |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] `AuditRepository` interface is defined in `internal/domain/audit.go`.
- [x] No update or delete methods on `AuditRepository` (append-only).
- [x] `AuditLog` has no `DeletedAt` field (audit logs are never soft-deleted).
- [x] `AuditLog` struct uses `time.Time` (not `pgtype`).
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | ✅ done    |
| Task 1.2.2 — `domain/role.go`    | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A (pure types/interfaces — tested via service layer in 1.4.5)

### Integration tests (`//go:build integration`)

Covered by Task 1.3.7 repository integration tests.

---

## 8. Open Questions

| # | Question                                                                   | Owner | Resolution |
| - | -------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should audit entries be written synchronously or via a background channel? | —     | Synchronous for Phase 1; async queue in Phase 5. |
| 2 | Should sysadmin audit logs be written to a global table or per-tenant?     | —     | Per-tenant with `tenant_id = "SYSTEM"` for cross-tenant events. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | —      | Entity and Repository interface implemented |
