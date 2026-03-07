# Task 1.3.1 — Repository: Tenant

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `TenantRepository` in `internal/platform/repository/tenant_repo.go` using the sqlc-generated code. This is the concrete database layer for tenant management, translating domain calls into SQL via the generated `Queries` struct.

---

## 2. Context & Motivation

The `TenantRepository` interface is defined in `internal/domain/tenant.go` (Task 1.2.3). This task provides the concrete implementation backed by PostgreSQL. Tenants are the root of the multi-tenancy hierarchy; all other entities depend on a valid tenant existing. The implementation must enforce soft deletes and map between sqlc models and domain entities.

---

## 3. Scope

### In scope

- [ ] Concrete `tenantRepo` struct implementing `domain.TenantRepository`.
- [ ] Constructor `NewTenantRepository(q *sqlc.Queries) domain.TenantRepository`.
- [ ] Mapping functions between `sqlc.Tenant` and `domain.Tenant`.
- [ ] Error translation: `pgx` not-found → `domain.ErrNotFound`, unique violation → `domain.ErrConflict`.

### Out of scope

- Service layer (Task 1.4.2).
- HTTP handlers (Task 1.5.5).
- Admin cross-tenant queries (Task 1.3.8).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                          |
| ------ | ---------------------------------------------- | -------------------------------- |
| CREATE | `internal/platform/repository/tenant_repo.go`  | Concrete TenantRepository impl   |

### Key interfaces / types

```go
type tenantRepo struct {
    q *sqlc.Queries
}

func NewTenantRepository(q *sqlc.Queries) domain.TenantRepository {
    return &tenantRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/tenants.sql` (Task 1.1.7/1.1.8):

| Query name        | sqlc mode  | Used by         |
| ----------------- | ---------- | --------------- |
| `CreateTenant`    | `:one`     | `Create`        |
| `GetTenantByID`   | `:one`     | `GetByID`       |
| `ListTenants`     | `:many`    | `List`          |
| `UpdateTenant`    | `:one`     | `Update`        |
| `SoftDeleteTenant`| `:exec`    | `Delete`        |

### Error cases to handle

| Scenario                 | pgx Error                      | Domain Error        |
| ------------------------ | ------------------------------ | ------------------- |
| Row not found            | `pgx.ErrNoRows`                | `domain.ErrNotFound` |
| Duplicate tenant name    | `23505` unique_violation       | `domain.ErrConflict` |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] Struct implements `domain.TenantRepository` (verified by compiler).
- [ ] All pgx errors are translated to domain sentinel errors.
- [ ] sqlc queries include only tenant-scoped operations.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                        | Type     | Status  |
| --------------------------------- | -------- | ------- |
| Task 1.2.3 — `domain/tenant.go`   | Upstream | ✅ done |
| Task 1.1.7 — sqlc query files     | Upstream | ✅ done |
| Task 1.1.8 — sqlc generate        | Upstream | ✅ done |
| Task 1.3.9 — integration tests    | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — repository implementations are tested via integration tests.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — all repository integration tests run together in a single Testcontainers suite.

---

## 8. Open Questions

| # | Question                                                    | Owner | Resolution |
| - | ----------------------------------------------------------- | ----- | ---------- |
| 1 | Should `List` include soft-deleted tenants via a parameter? | —     | No — admins use `AdminTenantRepository.ListAll`; `List` is active only. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
