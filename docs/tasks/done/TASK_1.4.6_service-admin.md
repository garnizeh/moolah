# Task 1.4.6 — `service/admin_service.go` + unit tests

> **Roadmap Ref:** Phase 1 — MVP › 1.4 Service Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `AdminService`, which aggregates cross-tenant sysadmin operations by wrapping `AdminTenantRepository`, `AdminUserRepository`, and `AdminAuditRepository`. It provides system-wide management capabilities (list all tenants, suspend/restore, hard delete, cross-tenant user lookup, and global audit log queries) restricted to the `sysadmin` role.

---

## 2. Context & Motivation

Sysadmin operations deliberately bypass the `tenant_id` isolation imposed on all standard repositories. They are needed for support, compliance, and billing operations. The `AdminService` acts as a clear boundary, ensuring these elevated-privilege operations are explicit and audited. See `docs/ARCHITECTURE.md` and roadmap item 1.4.6.

---

## 3. Scope

### In scope

- [ ] `internal/service/admin_service.go` — concrete `AdminService` struct.
- [ ] `internal/domain/admin.go` — add `AdminService` interface definition.
- [ ] **Tenant operations:**
  - `ListAllTenants(ctx, withDeleted bool) ([]Tenant, error)`
  - `GetTenantByID(ctx, id string) (*Tenant, error)` — no tenant guard.
  - `UpdateTenantPlan(ctx, id string, plan TenantPlan) (*Tenant, error)` — audit log `update`.
  - `SuspendTenant(ctx, id string) error` — soft-delete + audit log `soft_delete`.
  - `RestoreTenant(ctx, id string) error` — restore + audit log `restore`.
  - `HardDeleteTenant(ctx, id string) error` — permanent; audit log `soft_delete` (records intent); requires explicit caller confirmation via non-empty `confirmationToken` param.
- [ ] **User operations:**
  - `ListAllUsers(ctx) ([]User, error)`
  - `GetUserByID(ctx, id string) (*User, error)` — no tenant guard.
  - `ForceDeleteUser(ctx, id string) error` — hard delete; audit log `soft_delete`.
- [ ] **Audit log operations:**
  - `ListAuditLogs(ctx, params ListAuditLogsParams) ([]AuditLog, error)` — global, all tenants.
- [ ] Full unit tests in `internal/service/admin_service_test.go`.

### Out of scope

- HTTP handler (`handler/admin_handler.go`) — Task 1.5.9.
- Billing / plan enforcement middleware — Task 4.2.
- MRR / churn analytics endpoint — Task 4.6.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                       | Purpose                                          |
| ------ | ------------------------------------------ | ------------------------------------------------ |
| MODIFY | `internal/domain/admin.go`                 | Add `AdminService` interface                     |
| CREATE | `internal/service/admin_service.go`        | Concrete service implementation                  |
| CREATE | `internal/service/admin_service_test.go`   | Unit tests with mocked deps                      |

### Key interfaces / types

```go
// AdminService defines the business-logic contract for cross-tenant sysadmin operations.
type AdminService interface {
    // Tenant management
    ListAllTenants(ctx context.Context, withDeleted bool) ([]Tenant, error)
    GetTenantByID(ctx context.Context, id string) (*Tenant, error)
    UpdateTenantPlan(ctx context.Context, id string, plan TenantPlan) (*Tenant, error)
    SuspendTenant(ctx context.Context, id string) error
    RestoreTenant(ctx context.Context, id string) error
    HardDeleteTenant(ctx context.Context, id, confirmationToken string) error

    // User management
    ListAllUsers(ctx context.Context) ([]User, error)
    GetUserByID(ctx context.Context, id string) (*User, error)
    ForceDeleteUser(ctx context.Context, id string) error

    // Audit
    ListAuditLogs(ctx context.Context, params ListAuditLogsParams) ([]AuditLog, error)
}
```

### Business rules

1. `HardDeleteTenant`: `confirmationToken` must equal the tenant `ID` (a simple idempotent guard against accidental calls). If it does not match, return `domain.ErrInvalidInput`. Write audit log before executing the hard delete.
2. `SuspendTenant`: soft-delete the tenant; all subsequent auth attempts for users of that tenant should fail (handled at middleware level via token validation, not service level).
3. `RestoreTenant`: reverse soft-delete; write audit log `restore`.
4. All mutating operations write an audit log using `actorID = "SYSTEM"` when called outside a user context, or the sysadmin's user ID when called via HTTP.

### Error cases

| Scenario                              | Sentinel Error           | Notes                               |
| ------------------------------------- | ------------------------ | ----------------------------------- |
| Tenant not found                      | `domain.ErrNotFound`     | all tenant operations               |
| User not found                        | `domain.ErrNotFound`     | user operations                     |
| Confirmation token mismatch           | `domain.ErrInvalidInput` | `HardDeleteTenant`                  |

---

## 5. Acceptance Criteria

- [ ] `AdminService` interface defined in `internal/domain/admin.go`.
- [ ] `NewAdminService` constructor accepts `AdminTenantRepository`, `AdminUserRepository`, `AdminAuditRepository`, and `AuditRepository` (for write-side audit logs).
- [ ] `HardDeleteTenant` returns `ErrInvalidInput` when `confirmationToken != id`.
- [ ] `SuspendTenant` and `RestoreTenant` write the correct audit log actions.
- [ ] `ForceDeleteUser` writes audit log before executing.
- [ ] Unit tests cover all happy paths and all error branches (≥ 80% coverage).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                           | Type     | Status  |
| ------------------------------------ | -------- | ------- |
| Task 1.3.7 — `audit_repo.go`         | Upstream | ✅ done |
| Task 1.3.8 — `admin_repo.go`         | Upstream | ✅ done |
| Task 1.1.17 — `testutil/mocks`       | Upstream | ✅ done |
| `domain.ErrInvalidInput` sentinel    | Upstream | verify in `domain/errors.go` |

---

## 7. Testing Plan

Unit tests only.

Key scenarios:

- `ListAllTenants` → returns all tenants, with and without deleted.
- `UpdateTenantPlan` → plan changed → audit log `update`.
- `SuspendTenant` → soft-deleted → audit log `soft_delete`.
- `RestoreTenant` → restored → audit log `restore`.
- `HardDeleteTenant` → wrong confirmation token → `ErrInvalidInput`, no delete.
- `HardDeleteTenant` → correct token → deleted → audit log written.
- `ForceDeleteUser` → deleted → audit log written.
- `ListAuditLogs` → returns global logs.

---

## 8. Open Questions

| # | Question                                                                            | Owner | Resolution |
| - | ----------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `HardDeleteTenant` also hard-delete all users, accounts, transactions?       | —     | Yes — rely on DB `ON DELETE CASCADE`; no application-level cascade needed. |
| 2 | Should `ListAllUsers` support pagination in Phase 1?                                | —     | No; bounded by reasonable tenant count in Phase 1. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
