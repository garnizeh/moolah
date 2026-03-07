# Task 1.2.8 — Domain Category Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `Category` domain entity and `CategoryRepository` interface in `internal/domain/category.go`. Categories are tenant-scoped labels (income, expense, transfer) attached to transactions, and support one level of parent-child hierarchy.

---

## 2. Context & Motivation

Categories are necessary for all financial reporting. The parent–child hierarchy enables grouping (e.g., "Food" → "Restaurants", "Groceries"). Defining the interface in the domain layer ensures the category service (Task 1.4.4) can be unit-tested independently of the database.

---

## 3. Scope

### In scope

- [x] `Category` struct and `CategoryType` constants.
- [x] `CreateCategoryInput` and `UpdateCategoryInput` value objects.
- [x] `CategoryRepository` interface: CRUD + list with optional parent filter.
- [x] Helper method `IsRoot() bool` on `Category` (parent_id is empty).

### Out of scope

- Concrete repository implementation (Task 1.3.5).
- Service layer (Task 1.4.4).
- HTTP handlers (Task 1.5.7).
- Deep hierarchy / multi-level nesting (not in scope for Phase 1).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                          | Purpose                              |
| ------ | ----------------------------- | ------------------------------------ |
| CREATE | `internal/domain/category.go` | Entity, input types, repo interface  |
| CREATE | `internal/domain/category_test.go` | Unit tests for Category helpers |

### Key interfaces / types

```go
// CategoryType mirrors the database enum for transaction categories.
type CategoryType string

const (
    CategoryTypeIncome   CategoryType = "income"
    CategoryTypeExpense  CategoryType = "expense"
    CategoryTypeTransfer CategoryType = "transfer"
)

// Category is a tenant-scoped label for classifying transactions.
type Category struct {
    CreatedAt time.Time    `json:"created_at"`
    UpdatedAt time.Time    `json:"updated_at"`
    DeletedAt *time.Time   `json:"deleted_at,omitempty"`
    ID        string       `json:"id"`
    TenantID  string       `json:"tenant_id"`
    ParentID  string       `json:"parent_id"` // Empty string means root category
    Name      string       `json:"name"`
    Icon      string       `json:"icon"`  // Optional emoji or icon identifier
    Color     string       `json:"color"` // Optional hex color, e.g. "#FF5733"
    Type      CategoryType `json:"type"`
}

// IsRoot returns true when the category has no parent.
func (c *Category) IsRoot() bool { return c.ParentID == "" }

type CreateCategoryInput struct {
    ParentID string       `validate:"omitempty"`
    Name     string       `validate:"required,min=1,max=100"`
    Icon     string       `validate:"omitempty,max=10"`
    Color    string       `validate:"omitempty,hexcolor"`
    Type     CategoryType `validate:"required,oneof=income expense transfer"`
}

type UpdateCategoryInput struct {
    Name  *string `validate:"omitempty,min=1,max=100"`
    Icon  *string `validate:"omitempty,max=10"`
    Color *string `validate:"omitempty,hexcolor"`
}

// CategoryRepository defines persistence operations for categories.
type CategoryRepository interface {
    Create(ctx context.Context, tenantID string, input CreateCategoryInput) (*Category, error)
    GetByID(ctx context.Context, tenantID, id string) (*Category, error)
    ListByTenant(ctx context.Context, tenantID string) ([]Category, error)
    ListChildren(ctx context.Context, tenantID, parentID string) ([]Category, error)
    Update(ctx context.Context, tenantID, id string, input UpdateCategoryInput) (*Category, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/categories.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.7.

### Error cases to handle

| Scenario                      | Sentinel Error           | HTTP Status |
| ----------------------------- | ------------------------ | ----------- |
| Category not found            | `domain.ErrNotFound`     | `404`       |
| Parent category not found     | `domain.ErrNotFound`     | `404`       |
| Tenant mismatch               | `domain.ErrForbidden`    | `403`       |
| Duplicate name within tenant  | `domain.ErrConflict`     | `409`       |
| Invalid input                 | `domain.ErrInvalidInput` | `422`       |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] `CategoryRepository` interface is defined in `internal/domain/category.go`.
- [x] `IsRoot()` helper works correctly.
- [x] `Category` struct uses `time.Time` / `*time.Time` (not `pgtype`).
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/domain/category_test.go` (if helpers are added)
- **Cases:**
  - [x] `IsRoot()` returns `true` when `ParentID` is empty.
  - [x] `IsRoot()` returns `false` when `ParentID` is set.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.5 repository integration tests.

---

## 8. Open Questions

| # | Question                                                            | Owner | Resolution |
| - | ------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should we prevent deleting a category that has child categories?    | —     | Service-layer guard — return `ErrConflict` if children exist. |
| 2 | Should categories be shared across tenants (system defaults)?        | —     | No — all categories are tenant-scoped in Phase 1. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | —      | Entity and Repository interface implemented with unit tests |
