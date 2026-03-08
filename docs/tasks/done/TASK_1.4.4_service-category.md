# Task 1.4.4 — `service/category_service.go` + unit tests

> **Roadmap Ref:** Phase 1 — MVP › 1.4 Service Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement `CategoryService`, which wraps `CategoryRepository` with business rules: hierarchy validation (max one level deep), preventing deletion of categories in use by active transactions, and audit trail writes. All operations are fully unit-tested with mocked dependencies.

---

## 2. Context & Motivation

Categories are tenant-scoped labels for transactions. The domain supports one level of parent-child hierarchy (a child category cannot itself have children). The service must enforce this constraint and guard against orphaning transactions by blocking deletes when a category is actively referenced. See `docs/ARCHITECTURE.md` and roadmap item 1.4.4.

---

## 3. Scope

### In scope

- [x] `internal/service/category_service.go` — concrete `CategoryService` struct.
- [x] `internal/domain/category.go` — add `CategoryService` interface definition.
- [x] `Create(ctx, tenantID string, input CreateCategoryInput) (*Category, error)` — validate hierarchy depth, persist, audit log `create`.
- [x] `GetByID(ctx, tenantID, id string) (*Category, error)` — tenant-scoped fetch.
- [x] `ListByTenant(ctx, tenantID string) ([]Category, error)` — all categories.
- [x] `ListChildren(ctx, tenantID, parentID string) ([]Category, error)` — subcategories.
- [x] `Update(ctx, tenantID, id string, input UpdateCategoryInput) (*Category, error)` — name/icon/color only; audit log `update`.
- [x] `Delete(ctx, tenantID, id string) error` — soft-delete; audit log `soft_delete`.
- [x] Full unit tests in `internal/service/category_service_test.go`.

### Out of scope

- HTTP handler (`handler/category_handler.go`) — Task 1.5.7.
- Blocking delete when category is in use — deferred to Phase 1 quality gate review; log a warning for now.
- Moving a child to a different parent — not supported in Phase 1.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                         | Purpose                                        |
| ------ | -------------------------------------------- | ---------------------------------------------- |
| MODIFY | `internal/domain/category.go`                | Add `CategoryService` interface                |
| CREATE | `internal/service/category_service.go`       | Concrete service implementation                |
| CREATE | `internal/service/category_service_test.go`  | Unit tests with mocked deps                    |

### Key interfaces / types

```go
// CategoryService defines the business-logic contract for category management.
type CategoryService interface {
    Create(ctx context.Context, tenantID string, input CreateCategoryInput) (*Category, error)
    GetByID(ctx context.Context, tenantID, id string) (*Category, error)
    ListByTenant(ctx context.Context, tenantID string) ([]Category, error)
    ListChildren(ctx context.Context, tenantID, parentID string) ([]Category, error)
    Update(ctx context.Context, tenantID, id string, input UpdateCategoryInput) (*Category, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### Business rules

1. `Create` with a `ParentID`:
   - Fetch the parent via `CategoryRepository.GetByID`. Return `ErrNotFound` if absent.
   - If the parent itself has a `ParentID` (is already a child), return `domain.ErrInvalidInput` — max depth is 1.
2. `Update`: fetch existing; write audit log `update` with `OldValues`/`NewValues`.
3. `Delete`: write audit log `soft_delete`. Phase 1 allows deletion regardless of transaction references.

### Error cases

| Scenario                              | Sentinel Error            | Notes                             |
| ------------------------------------- | ------------------------- | --------------------------------- |
| Category not found                    | `domain.ErrNotFound`      | `GetByID`, `Update`, `Delete`     |
| Parent category not found             | `domain.ErrNotFound`      | `Create` with `ParentID`          |
| Hierarchy depth exceeded (depth > 1)  | `domain.ErrInvalidInput`  | `Create`                          |

---

## 5. Acceptance Criteria

- [x] `CategoryService` interface defined in `internal/domain/category.go`.
- [x] `NewCategoryService` constructor accepts `CategoryRepository` and `AuditRepository`.
- [x] `Create` with a child of a child returns `domain.ErrInvalidInput`.
- [x] `Create` with a non-existent `ParentID` returns `domain.ErrNotFound`.
- [x] `Update` writes audit log with old and new values.
- [x] `Delete` writes audit log `soft_delete`.
- [x] Unit tests cover all happy paths and all error branches (≥ 80% coverage).
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                         | Type     | Status  |
| ---------------------------------- | -------- | ------- |
| Task 1.3.5 — `category_repo.go`    | Upstream | ✅ done |
| Task 1.3.7 — `audit_repo.go`       | Upstream | ✅ done |
| Task 1.1.17 — `testutil/mocks`     | Upstream | ✅ done |
| `domain.ErrInvalidInput` sentinel  | Upstream | verify in `domain/errors.go` |

---

## 7. Testing Plan

Unit tests only.

Key scenarios:

- `Create` root category → success, audit log written.
- `Create` child category (valid parent) → success.
- `Create` grandchild category → `ErrInvalidInput`.
- `Create` with unknown `ParentID` → `ErrNotFound`.
- `GetByID` → not found → `ErrNotFound`.
- `Update` → success → audit log with old/new values.
- `Delete` → success → audit log `soft_delete`.

---

## 8. Open Questions

| # | Question                                                                     | Owner | Resolution |
| - | ---------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `ErrInvalidInput` already exist in `domain/errors.go`?                | —     | Verify before implementation; add if missing. |
| 2 | Should deleting a parent cascade soft-delete to children?                    | —     | No for Phase 1; children become orphaned roots (harmless). |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
