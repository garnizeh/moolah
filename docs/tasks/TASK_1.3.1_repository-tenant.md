# Task 1.3.1 — Repository: Tenant

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Implement `TenantRepository` in `internal/platform/repository/tenant_repo.go` using the sqlc-generated code. This is the concrete database layer for tenant management, translating domain calls into SQL via the generated `Querier` interface.

---

## 2. Context & Motivation

The `TenantRepository` interface is defined in `internal/domain/tenant.go` (Task 1.2.3). This task provides the concrete implementation backed by PostgreSQL. Tenants are the root of the multi-tenancy hierarchy; all other entities depend on a valid tenant existing. The implementation must enforce soft deletes and map between sqlc models and domain entities.

---

## 3. Scope

### In scope

- [x] Concrete `tenantRepo` struct implementing `domain.TenantRepository`.
- [x] Constructor `NewTenantRepository(q sqlc.Querier) domain.TenantRepository`.
- [x] Mapping functions between `sqlc.Tenant` and `domain.Tenant`.
- [x] Error translation: `pgx` not-found → `domain.ErrNotFound`, unique violation → `domain.ErrConflict`.
- [x] Unit tests with 100% coverage using mocked `sqlc.Querier`.

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
| CREATE | `internal/platform/repository/tenant_repo_test.go` | Unit tests with 100% coverage |
| CREATE | `internal/platform/db/sqlc/mock_querier.go`    | Shared mock for all repositories |

### Key interfaces / types

```go
type tenantRepo struct {
    q sqlc.Querier
}

func NewTenantRepository(q sqlc.Querier) domain.TenantRepository {
    return &tenantRepo{q: q}
}
```

### SQL queries (sqlc)

All queries exist in `internal/platform/db/queries/tenants.sql`:

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

- [x] All exported types and functions have Go doc comments.
- [x] Struct implements `domain.TenantRepository` (verified by compiler).
- [x] All pgx errors are translated to domain sentinel errors.
- [x] sqlc queries include only tenant-scoped operations.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

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

Implemented in `internal/platform/repository/tenant_repo_test.go`. Achieved 100% statement coverage using `testify/mock`.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — all repository integration tests run together in a single Testcontainers suite.

---

## 8. Open Questions

| # | Question                                                    | Owner | Resolution |
| - | ----------------------------------------------------------- | ----- | ---------- |
| 1 | Should `List` include soft-deleted tenants via a parameter? | —     | No — admins use `AdminTenantRepository.ListAll`; `List` is active only. |

---

## 9. Change Log

| Date       | Author          | Change                                |
| ---------- | --------------- | ------------------------------------- |
| 2026-03-07 | GitHub Copilot  | Initial implementation and unit tests |
| 2026-03-07 | GitHub Copilot  | Added shared MockQuerier and attained 100% coverage |
