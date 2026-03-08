# Task 1.3.8 — Repository: Admin

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the three admin repository interfaces (`AdminTenantRepository`, `AdminUserRepository`, `AdminAuditRepository`) defined in `internal/domain/admin.go` (Task 1.2.11) as concrete types in `internal/platform/repository/admin_repo.go`. These implementations omit the standard `tenant_id` filter and are exclusively callable by the sysadmin role.

---

## 2. Context & Motivation

Admin operations require cross-tenant access that is structurally impossible with the tenant-scoped repositories. By placing the admin implementations in a separate file and constructor, the dependency injection wiring can ensure these repositories are only injected into the admin service, making it an architecture-level safety guarantee rather than a runtime check.

New or supplemental SQL query files will be needed in `internal/platform/db/queries/` for admin-specific queries (those that omit `WHERE tenant_id = $1`).

---

## 3. Scope

### In scope

- [x] Concrete `adminTenantRepo`, `adminUserRepo`, `adminAuditRepo` structs.
- [x] Constructor functions for each: `NewAdminTenantRepository`, `NewAdminUserRepository`, `NewAdminAuditRepository`.
- [x] New sqlc query files for admin queries (no `tenant_id` filter):
  - `internal/platform/db/queries/admin_tenants.sql`
  - `internal/platform/db/queries/admin_users.sql`
  - `internal/platform/db/queries/admin_audit_logs.sql`
- [x] Run `sqlc generate` after adding admin queries.
- [x] `HardDelete` deletes the tenant row permanently (relies on `ON DELETE CASCADE`).

### Out of scope

- Admin service orchestration (Task 1.4.6).
- Admin HTTP handlers (Task 1.5.9).
- Billing / plan enforcement (Phase 4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                             |
| ------ | ---------------------------------------------- | ----------------------------------- |
| CREATE | `internal/platform/repository/admin_repo.go`  | Three admin repository impls        |
| CREATE | `internal/platform/db/queries/admin_tenants.sql` | Cross-tenant tenant queries       |
| CREATE | `internal/platform/db/queries/admin_users.sql`   | Cross-tenant user queries         |
| CREATE | `internal/platform/db/queries/admin_audit_logs.sql` | Global audit log queries       |

### Key interfaces / types

```go
type adminTenantRepo struct { q *sqlc.Queries }
type adminUserRepo   struct { q *sqlc.Queries }
type adminAuditRepo  struct { q *sqlc.Queries }

func NewAdminTenantRepository(q *sqlc.Queries) domain.AdminTenantRepository
func NewAdminUserRepository(q *sqlc.Queries)   domain.AdminUserRepository
func NewAdminAuditRepository(q *sqlc.Queries)  domain.AdminAuditRepository
```

### SQL queries (sqlc) — new files needed

**admin_tenants.sql:**

```sql
-- name: AdminListAllTenants :many
-- name: AdminGetTenantByID :one
-- name: AdminUpdateTenantPlan :one
-- name: AdminSuspendTenant :exec
-- name: AdminRestoreTenant :exec
-- name: AdminHardDeleteTenant :exec
```

**admin_users.sql:**

```sql
-- name: AdminListAllUsers :many
-- name: AdminGetUserByID :one
-- name: AdminForceDeleteUser :exec
```

**admin_audit_logs.sql:**

```sql
-- name: AdminListAllAuditLogs :many
```

### Error cases to handle

| Scenario            | pgx Error       | Domain Error           |
| ------------------- | --------------- | ---------------------- |
| Row not found       | `pgx.ErrNoRows` | `domain.ErrNotFound`   |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] All three structs implement their respective admin interfaces (compiler-verified).
- [x] Admin queries do **not** include `WHERE tenant_id = $1`.
- [x] `HardDelete` performs a permanent row deletion (not soft-delete).
- [x] `sqlc generate` passes after adding new query files.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                        | Type       | Status     |
| --------------------------------- | ---------- | ---------- |
| Task 1.2.11 — `domain/admin.go`   | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files     | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate        | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests    | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

Implemented in `internal/platform/repository/admin_repo_test.go`.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — specifically:

- `ListAll` returns tenants across all tenants (no filter).
- `Suspend` + `Restore` toggles `deleted_at` correctly.
- `HardDelete` permanently removes the row and cascades.
- `AdminListAllAuditLogs` returns entries from multiple tenants.

---

## 8. Open Questions

| # | Question                                                                      | Owner | Resolution |
| - | ----------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `HardDelete` also clean up uploaded files or external resources?       | —     | Out of scope for Phase 1 — no file storage yet. |
| 2 | Should `AdminListAllTenants` accept `withDeleted bool` in SQL or Go?          | —     | In SQL via a nullable parameter; sqlc `nullable` argument. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | Copilot| Implemented admin repositories and tests |
