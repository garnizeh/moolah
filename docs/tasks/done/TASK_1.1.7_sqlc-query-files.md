# Task 1.1.7 — sqlc Query Files for All Phase 1 Entities

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Write all raw SQL query files consumed by `sqlc` to generate the typed Go data-access layer for every Phase 1 entity: tenants, users, OTP requests, accounts, categories, transactions, and audit logs. Every query must be tenant-scoped and include soft-delete guards.

---

## 2. Context & Motivation

`sqlc` generates type-safe Go code from annotated SQL. Writing the `.sql` files is a prerequisite for task 1.1.8 (`sqlc generate`) and all repository tasks (1.3.x). Every query is the canonical definition of what data is fetched; correctness here prevents entire classes of runtime bugs. The `sqlc.yaml` is already configured with `emit_interface: true`, so `sqlc` will also generate repository interfaces.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.7
- Schema reference: `docs/schema.sql`
- Config reference: `sqlc.yaml`

---

## 3. Scope

### In scope

- [ ] `internal/platform/db/queries/tenants.sql`
- [ ] `internal/platform/db/queries/users.sql`
- [ ] `internal/platform/db/queries/auth.sql` (otp_requests)
- [ ] `internal/platform/db/queries/accounts.sql`
- [ ] `internal/platform/db/queries/categories.sql`
- [ ] `internal/platform/db/queries/transactions.sql`
- [ ] `internal/platform/db/queries/audit_logs.sql`

### Out of scope

- `master_purchases` queries (Phase 2 — task 2.3)
- Admin cross-tenant queries (task 1.2.11 / 1.3.8 — separate query files)
- Full-text search queries (deferred to Phase 5)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                               | Purpose                              |
| ------ | -------------------------------------------------- | ------------------------------------ |
| CREATE | `internal/platform/db/queries/tenants.sql`         | Tenant CRUD + plan management        |
| CREATE | `internal/platform/db/queries/users.sql`           | User CRUD + email lookup             |
| CREATE | `internal/platform/db/queries/auth.sql`            | OTP request create, lookup, mark used |
| CREATE | `internal/platform/db/queries/accounts.sql`        | Account CRUD + balance update        |
| CREATE | `internal/platform/db/queries/categories.sql`      | Category CRUD + hierarchy            |
| CREATE | `internal/platform/db/queries/transactions.sql`    | Transaction CRUD + list with filters |
| CREATE | `internal/platform/db/queries/audit_logs.sql`      | Audit log insert + list              |

### Query catalogue

**tenants.sql**

```sql
-- name: CreateTenant :one
-- name: GetTenantByID :one
-- name: UpdateTenant :one
-- name: SoftDeleteTenant :exec
```

**users.sql**

```sql
-- name: CreateUser :one
-- name: GetUserByID :one
-- name: GetUserByEmail :one          -- WHERE tenant_id = $1 AND email = $2
-- name: ListUsersByTenant :many
-- name: UpdateUser :one
-- name: UpdateUserLastLogin :exec
-- name: SoftDeleteUser :exec
```

**auth.sql**

```sql
-- name: CreateOTPRequest :one
-- name: GetActiveOTPByEmail :one     -- WHERE email = $1 AND used = FALSE AND expires_at > NOW()
-- name: MarkOTPUsed :exec
-- name: DeleteExpiredOTPs :exec      -- housekeeping, called by scheduled job or on-demand
```

**accounts.sql**

```sql
-- name: CreateAccount :one
-- name: GetAccountByID :one          -- WHERE tenant_id = $1 AND id = $2
-- name: ListAccountsByTenant :many
-- name: UpdateAccount :one
-- name: UpdateAccountBalance :exec
-- name: SoftDeleteAccount :exec
```

**categories.sql**

```sql
-- name: CreateCategory :one
-- name: GetCategoryByID :one
-- name: ListCategoriesByTenant :many
-- name: ListRootCategoriesByTenant :many  -- WHERE parent_id IS NULL
-- name: ListChildCategories :many         -- WHERE parent_id = $2
-- name: UpdateCategory :one
-- name: SoftDeleteCategory :exec
```

**transactions.sql**

```sql
-- name: CreateTransaction :one
-- name: GetTransactionByID :one
-- name: ListTransactionsByTenant :many    -- paginated, filtered by date range / account / category
-- name: ListTransactionsByAccount :many
-- name: UpdateTransaction :one
-- name: SoftDeleteTransaction :exec
```

**audit_logs.sql**

```sql
-- name: CreateAuditLog :one
-- name: ListAuditLogsByTenant :many       -- paginated
-- name: ListAuditLogsByEntity :many       -- WHERE entity_type = $2 AND entity_id = $3
```

### Mandatory constraints for every tenant-scoped query

1. First parameter is always `tenant_id` (`$1`).
2. All SELECT/UPDATE/DELETE include `AND deleted_at IS NULL` (except restore operations).
3. Pagination uses `LIMIT $n OFFSET $m` with named params `@limit` / `@offset`.
4. ORDER BY is always explicit (never rely on undefined ordering).

### SQL queries (sqlc)

These files _are_ the sqlc query definitions.

### API endpoints (if applicable)

N/A

### Error cases to handle

N/A at this layer — `sqlc` codegen; no runtime logic.

---

## 5. Acceptance Criteria

- [ ] `sqlc generate` runs without errors after all query files are written.
- [ ] Every tenant-scoped query has `tenant_id = $1` (or `@tenant_id`) as the first filter.
- [ ] Every SELECT/UPDATE query includes `AND deleted_at IS NULL`.
- [ ] No query uses `SELECT *` — columns are always explicit.
- [ ] `sqlc vet` passes (if available in the installed version).
- [ ] Generated code compiles with `go build ./...`.
- [ ] `docs/ROADMAP.md` row 1.1.7 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                       | Type     | Status     |
| ------------------------------------------------ | -------- | ---------- |
| Task 1.1.6 migrations complete (schema defined)  | Upstream | 🔵 backlog |
| `sqlc.yaml` configured (`emit_interface: true`)  | Upstream | ✅ done   |
| `github.com/sqlc-dev/sqlc` installed             | Tool     | ✅ done   |

---

## 7. Testing Plan

### Unit tests

N/A — `.sql` files have no Go logic.

### Integration tests (`//go:build integration`)

- Covered by repository-layer integration tests (tasks 1.3.1–1.3.8), which execute the generated queries against a live PostgreSQL instance via `testcontainers-go`.

### CI verification (task 1.1.8)

- `sqlc generate` must produce zero diff against committed generated code.

---

## 8. Open Questions

| # | Question                                                              | Owner | Resolution |
| - | --------------------------------------------------------------------- | ----- | ---------- |
| 1 | Use `LIMIT`/`OFFSET` or keyset pagination for `ListTransactions`?     | —     | `LIMIT`/`OFFSET` for Phase 1 simplicity; keyset deferred to Phase 5 performance work. |
| 2 | Should `ListTransactionsByTenant` support multiple filter combinations? | — | One query with nullable filter params (use `COALESCE` / `IS NULL OR` pattern). |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.7 |
