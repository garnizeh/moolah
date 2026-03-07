# Task 1.2.11 — Domain Admin Repository Interfaces

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define admin-only repository interfaces in `internal/domain/admin.go`. These interfaces expose cross-tenant and privileged operations required exclusively by the sysadmin role — listing all tenants, suspending users, reading global audit logs, and performing system health queries. They are kept separate from tenant-scoped repository interfaces to enforce the principle of least privilege.

---

## 2. Context & Motivation

The sysadmin role needs to manage tenants and users across the entire system, operations that must never be exposed through tenant-scoped repositories (which always filter by `tenant_id`). Defining separate admin interfaces in the domain layer makes it impossible for regular service code to accidentally call cross-tenant queries, and enables the admin service (Task 1.4.6) to be unit-tested with dedicated mocks.

---

## 3. Scope

### In scope

- [ ] `AdminTenantRepository` — cross-tenant tenant management (list all, suspend, restore).
- [ ] `AdminUserRepository` — cross-tenant user management (list all users, force-delete).
- [ ] `AdminAuditRepository` — global audit log query (no tenant filter).
- [ ] All interfaces enforce no `tenant_id` filter — callers supply it explicitly when needed.

### Out of scope

- Concrete repository implementation (Task 1.3.8).
- Admin service orchestration (Task 1.4.6).
- Admin HTTP handlers (Task 1.5.9).
- Billing/plan management (Phase 4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                       | Purpose                                |
| ------ | -------------------------- | -------------------------------------- |
| CREATE | `internal/domain/admin.go` | Admin-only repository interfaces       |

### Key interfaces / types

```go
// AdminTenantRepository defines cross-tenant operations for sysadmin use only.
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
    // Must only be called after explicit confirmation. Irreversible.
    HardDelete(ctx context.Context, id string) error
}

// AdminUserRepository defines cross-tenant user operations for sysadmin use only.
type AdminUserRepository interface {
    // ListAll returns every user in the system regardless of tenant.
    ListAll(ctx context.Context) ([]User, error)
    // GetByID retrieves a user without a tenant_id guard.
    GetByID(ctx context.Context, id string) (*User, error)
    // ForceDelete hard-deletes a user record. Use with caution.
    ForceDelete(ctx context.Context, id string) error
}

// AdminAuditRepository defines global audit log queries without tenant isolation.
type AdminAuditRepository interface {
    // ListAll returns audit logs across all tenants with optional filters.
    ListAll(ctx context.Context, params ListAuditLogsParams) ([]AuditLog, error)
}
```

### SQL queries (sqlc)

Admin queries will require new or supplemental query files added during Task 1.3.8. Current queries in `internal/platform/db/queries/` filter by `tenant_id`; admin queries omit this filter.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.9.

### Error cases to handle

| Scenario                  | Sentinel Error        | HTTP Status |
| ------------------------- | --------------------- | ----------- |
| Tenant not found          | `domain.ErrNotFound`  | `404`       |
| User not found            | `domain.ErrNotFound`  | `404`       |
| Caller is not sysadmin    | `domain.ErrForbidden` | `403`       |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] `AdminTenantRepository`, `AdminUserRepository`, and `AdminAuditRepository` are defined in `internal/domain/admin.go`.
- [ ] No regular (tenant-scoped) repository methods are duplicated — only admin-specific overrides.
- [ ] `HardDelete` carries a doc comment warning that it is irreversible.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                        | Type     | Status     |
| --------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`   | Upstream | 🔵 backlog |
| Task 1.2.3 — `domain/tenant.go`   | Upstream | 🔵 backlog |
| Task 1.2.4 — `domain/user.go`     | Upstream | 🔵 backlog |
| Task 1.2.10 — `domain/audit.go`   | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A (pure interfaces — tested via admin service in 1.4.6)

### Integration tests (`//go:build integration`)

Covered by Task 1.3.8 admin repository integration tests.

---

## 8. Open Questions

| # | Question                                                                         | Owner | Resolution |
| - | -------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `HardDelete` cascade to all child entities at the DB level or application? | —    | DB-level cascade via `ON DELETE CASCADE` in migrations. |
| 2 | Should sysadmin be able to impersonate users for debugging?                       | —    | Defer to Phase 4 — too risky without audit trail in place. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
