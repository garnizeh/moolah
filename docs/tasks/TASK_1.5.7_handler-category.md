# Task 1.5.7 — `handler/category_handler.go` — CRUD

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement the category HTTP handler in `internal/handler/category_handler.go`. It exposes CRUD endpoints for transaction categories (income and expense labels with optional one-level parent-child hierarchy), scoped to the authenticated user's tenant.

---

## 2. Context & Motivation

The `CategoryService` is fully implemented (Task 1.4.4) but has no HTTP entry point. Categories are required before transactions can be created, so this handler must be in place before the transaction handler (Task 1.5.8). See `docs/ARCHITECTURE.md` and roadmap item 1.5.7.

---

## 3. Scope

### In scope

- [ ] `internal/handler/category_handler.go` — `CategoryHandler` struct + 5 HTTP handler methods.
- [ ] `GET /v1/categories` — list all categories for the tenant.
- [ ] `POST /v1/categories` — create a category (root or child with `parent_id`).
- [ ] `GET /v1/categories/{id}` — get a single category by ID.
- [ ] `PATCH /v1/categories/{id}` — partial update (name, icon, color).
- [ ] `DELETE /v1/categories/{id}` — soft delete.
- [ ] Unit tests in `internal/handler/category_handler_test.go`.

### Out of scope

- Blocking delete when category is referenced by active transactions — log warning for now.
- Sub-category listing endpoint — `ListChildren` is called internally; expose via query param `?parent_id=` on the list endpoint.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                        | Purpose                               |
| ------ | ------------------------------------------- | ------------------------------------- |
| CREATE | `internal/handler/category_handler.go`      | HTTP handler for category endpoints   |
| CREATE | `internal/handler/category_handler_test.go` | Unit tests with mocked CategoryService|

### Request / Response types

```go
type CreateCategoryRequest struct {
    Name     string            `json:"name"      validate:"required,min=1,max=100"`
    Type     domain.CategoryType `json:"type"    validate:"required"`
    ParentID *string           `json:"parent_id" validate:"omitempty"`
    Icon     *string           `json:"icon"      validate:"omitempty"`
    Color    *string           `json:"color"     validate:"omitempty"`
}

type UpdateCategoryRequest struct {
    Name  *string `json:"name"  validate:"omitempty,min=1,max=100"`
    Icon  *string `json:"icon"  validate:"omitempty"`
    Color *string `json:"color" validate:"omitempty"`
}
```

### API endpoints

| Method | Path                    | Auth Required | Description              |
| ------ | ----------------------- | ------------- | ------------------------ |
| GET    | `/v1/categories`        | ✅ Bearer     | List all categories      |
| POST   | `/v1/categories`        | ✅ Bearer     | Create a category        |
| GET    | `/v1/categories/{id}`   | ✅ Bearer     | Get category by ID       |
| PATCH  | `/v1/categories/{id}`   | ✅ Bearer     | Update category          |
| DELETE | `/v1/categories/{id}`   | ✅ Bearer     | Soft-delete category     |

### Error cases to handle

| Scenario                     | Sentinel Error           | HTTP Status |
| ---------------------------- | ------------------------ | ----------- |
| Not found                    | `domain.ErrNotFound`     | `404`       |
| Validation failure           | —                        | `422`       |
| Hierarchy violation          | `domain.ErrInvalidInput` | `422`       |
| Conflict (duplicate)         | `domain.ErrConflict`     | `409`       |

---

## 5. Acceptance Criteria

- [ ] All 5 endpoints decode, validate, and call the service correctly.
- [ ] `tenant_id` is always sourced from context.
- [ ] All domain error sentinels map to the correct HTTP status codes.
- [ ] Unit tests cover all happy paths and error cases.
- [ ] Test coverage for handler ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                               | Type     | Status  |
| ---------------------------------------- | -------- | ------- |
| Task 1.4.4 — `service/category_service`  | Upstream | ✅ done |
| Task 1.1.9 — Auth middleware             | Upstream | ✅ done |
| `domain.CategoryService` interface       | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/category_handler_test.go`
- **Cases:**
  - `List`: returns categories array → `200 OK`.
  - `Create`: root category → `201 Created`.
  - `Create`: child category with valid `parent_id` → `201 Created`.
  - `Create`: nested child (third level) → `422` (hierarchy violation).
  - `Create`: duplicate name → `409`.
  - `GetByID`: found → `200 OK`.
  - `GetByID`: not found → `404`.
  - `Update`: valid partial update → `200 OK`.
  - `Delete`: success → `204 No Content`.

---

## 8. Open Questions

| # | Question                                      | Owner | Resolution |
| - | --------------------------------------------- | ----- | ---------- |
| 1 | Should the list endpoint filter by `type`?    | —     | Yes — support `?type=income` and `?type=expense` query params. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
