# Task 1.2.10 — Domain AuditLog Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** 🔵 `backlog`
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

- [ ] `AuditLog` struct and `AuditAction` constants.
- [ ] `CreateAuditLogInput` value object.
- [ ] `AuditRepository` interface: `Create` + `ListByTenant` + `ListByEntity`.
- [ ] Audit logs are **append-only** — no update or delete methods on the repository.

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
// AuditAction mirrors the DB enum.
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
    ID         string
    TenantID   string
    ActorID    string      // User ULID or "SYSTEM" for automated actions
    ActorRole  Role
    Action     AuditAction
    EntityType string      // e.g. "transaction", "account", "user"
    EntityID   string      // ULID of the affected entity
    OldValues  []byte      // JSON snapshot before change (nil for creates)
    NewValues  []byte      // JSON snapshot after change (nil for deletes)
    IPAddress  string      // IPv4 or IPv6, empty if not applicable
    UserAgent  string      // HTTP User-Agent header, empty if not applicable
    CreatedAt  time.Time
}

type CreateAuditLogInput struct {
    TenantID   string      `validate:"required"`
    ActorID    string      `validate:"required"`
    ActorRole  Role        `validate:"required"`
    Action     AuditAction `validate:"required"`
    EntityType string      `validate:"required"`
    EntityID   string      `validate:"omitempty"`
    OldValues  []byte
    NewValues  []byte
    IPAddress  string
    UserAgent  string
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

- [ ] All exported types and functions have Go doc comments.
- [ ] `AuditRepository` interface is defined in `internal/domain/audit.go`.
- [ ] No update or delete methods on `AuditRepository` (append-only).
- [ ] `AuditLog` has no `DeletedAt` field (audit logs are never soft-deleted).
- [ ] `AuditLog` struct uses `time.Time` (not `pgtype`).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | 🔵 backlog |
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
