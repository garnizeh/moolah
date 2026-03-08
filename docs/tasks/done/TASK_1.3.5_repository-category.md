# Task 1.3.5 — Repository: Category

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `CategoryRepository` in `internal/platform/repository/category_repo.go` using the sqlc-generated code. Categories are tenant-scoped labels that support one level of parent-child hierarchy for organizing transactions.

---

## 2. Context & Motivation

The `CategoryRepository` interface is defined in `internal/domain/category.go` (Task 1.2.8). The `ListChildren` method supports the hierarchy view: given a parent category ID, return all direct subcategories. This allows the service layer to enforce the hierarchy constraint (no grandchildren in Phase 1) and to guard against orphaned categories on delete.

---

## 3. Scope

### In scope

- [x] Concrete `categoryRepo` struct implementing `domain.CategoryRepository`.
- [x] Constructor `NewCategoryRepository(q *sqlc.Queries) domain.CategoryRepository`.
- [x] Mapping functions between `sqlc.Category` and `domain.Category`.
- [x] Error translation: `pgx.ErrNoRows` → `domain.ErrNotFound`, unique violation → `domain.ErrConflict`.
- [x] All queries enforce `WHERE tenant_id = $1 AND deleted_at IS NULL`.

### Out of scope

- Hierarchy depth enforcement (service layer, Task 1.4.4).
- HTTP handlers (Task 1.5.7).
- Deep multi-level nesting (Phase 3+).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                             | Purpose                           |
| ------ | ------------------------------------------------ | --------------------------------- |
| CREATE | `internal/platform/repository/category_repo.go` | Concrete CategoryRepository impl  |
| CREATE | `internal/platform/repository/category_repo_test.go` | Unit tests for CategoryRepository |

### Key interfaces / types

```go
type categoryRepo struct {
    q sqlc.Querier
}

func NewCategoryRepository(q sqlc.Querier) domain.CategoryRepository {
    return &categoryRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/categories.sql` (Task 1.1.7/1.1.8):

| Query name               | sqlc mode | Used by          |
| ------------------------ | --------- | ---------------- |
| `CreateCategory`         | `:one`    | `Create`         |
| `GetCategoryByID`        | `:one`    | `GetByID`        |
| `ListCategoriesByTenant` | `:many`   | `ListByTenant`   |
| `ListCategoryChildren`   | `:many`   | `ListChildren`   |
| `UpdateCategory`         | `:one`    | `Update`         |
| `SoftDeleteCategory`     | `:exec`   | `Delete`         |

### Error cases to handle

| Scenario                      | pgx Error                | Domain Error           |
| ----------------------------- | ------------------------ | ---------------------- |
| Row not found                 | `pgx.ErrNoRows`          | `domain.ErrNotFound`   |
| Duplicate category name       | `23505` unique_violation | `domain.ErrConflict`   |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] Struct implements `domain.CategoryRepository` (verified by compiler).
- [x] Every SQL query enforces `tenant_id` isolation and `deleted_at IS NULL`.
- [x] `ListChildren` correctly filters by `parent_id` and `tenant_id`.
- [x] All pgx errors are translated to domain sentinel errors.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                         | Type       | Status     |
| ---------------------------------- | ---------- | ---------- |
| Task 1.2.8 — `domain/category.go`  | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files      | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate         | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests     | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

Implemented in `internal/platform/repository/category_repo_test.go` with 100% code coverage.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — specifically:

- Create root category and child category.
- `ListChildren` returns only direct children.
- Cross-tenant category lookup returns `ErrNotFound`.
- Soft delete removes category from list queries.

---

## 8. Open Questions

| # | Question                                                               | Owner | Resolution |
| - | ---------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `Delete` block when child categories exist?                     | —     | Service-layer guard — repository only performs the soft delete. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | —      | Task completed with 100% unit test coverage |
