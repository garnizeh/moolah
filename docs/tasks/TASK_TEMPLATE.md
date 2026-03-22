# Task X.Y.Z — [Short Title]

> **Roadmap Ref:** Phase X — [Phase Name] › [Section Name]
> **Status:** 🔵 `backlog`
> **Last Updated:** YYYY-MM-DD
> **Assignee:** —
> **Estimated Effort:** S / M / L / XL

---

## 1. Summary

> One paragraph describing what this task delivers and why it matters to the system.

---

## 2. Context & Motivation

> Explain the problem being solved or the capability being added. Reference the relevant
> architecture decision, domain rule, or upstream dependency that makes this task necessary.
>
> - Link to relevant section in `docs/ARCHITECTURE.md` if applicable.
> - Link to relevant roadmap row: `docs/ROADMAP.md#X.Y.Z`.

---

## 3. Scope

### In scope

- [ ] Item 1
- [ ] Item 2

### Out of scope

- Item A (deferred to task X.Y.Z+1 or phase N)

---

## 4. Technical Design

### Files to create / modify

| Action   | Path                                      | Purpose                       |
| -------- | ----------------------------------------- | ----------------------------- |
| CREATE   | `internal/domain/example.go`              | Entity definition + interface |
| CREATE   | `internal/platform/repository/example.go` | Concrete repository impl      |
| CREATE   | `internal/service/example_service.go`     | Business-logic orchestration  |
| MODIFY   | `cmd/api/routes.go`                       | Register new routes           |

### Key interfaces / types

```go
// Domain entity — store in internal/domain/
type Example struct {
    ID       string    `json:"id"`
    TenantID string    `json:"-"`
    // ...
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    DeletedAt *time.Time `json:"-"`
}

// Repository interface — always in internal/domain/
type ExampleRepository interface {
    Create(ctx context.Context, tenantID string, input CreateExampleInput) (*Example, error)
    GetByID(ctx context.Context, tenantID, id string) (*Example, error)
    List(ctx context.Context, tenantID string, params ListParams) ([]Example, error)
    Update(ctx context.Context, tenantID, id string, input UpdateExampleInput) (*Example, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### SQL queries (sqlc)

List the named queries that need to be created in `internal/platform/db/queries/`:

```sql
-- name: CreateExample :one
-- name: GetExampleByID :one
-- name: ListExamples :many
-- name: UpdateExample :one
-- name: SoftDeleteExample :exec
```

### API endpoints (if applicable)

| Method | Path                     | Auth Required | Description           |
| ------ | ------------------------ | ------------- | --------------------- |
| POST   | `/v1/examples`           | ✅ Bearer     | Create a new example  |
| GET    | `/v1/examples/:id`       | ✅ Bearer     | Get by ID             |
| GET    | `/v1/examples`           | ✅ Bearer     | List with filters     |
| PATCH  | `/v1/examples/:id`       | ✅ Bearer     | Partial update        |
| DELETE | `/v1/examples/:id`       | ✅ Bearer     | Soft delete           |

### Error cases to handle

| Scenario                        | Sentinel Error       | HTTP Status |
| ------------------------------- | -------------------- | ----------- |
| Record not found                | `domain.ErrNotFound` | `404`       |
| Tenant mismatch / unauthorized  | `domain.ErrForbidden`| `403`       |
| Invalid input                   | validation error     | `422`       |

---

## 5. Acceptance Criteria

- [ ] All new exported types and functions have Go doc comments.
- [ ] Repository interface is defined in `internal/domain/`.
- [ ] Every SQL query includes `WHERE tenant_id = $1` and `AND deleted_at IS NULL`.
- [ ] Unit tests cover the service layer with mocked repository (gomock/moq).
- [ ] Integration tests cover the repository layer using `testcontainers-go` (build tag `integration`).
- [ ] Test coverage for new code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] Swaggo annotations added to handler (if HTTP endpoints were introduced).
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                      | Type     | Status     |
| ------------------------------- | -------- | ---------- |
| Task X.Y.Z-1 must be completed  | Upstream | 🔵 backlog |
| `domain.ErrNotFound` defined    | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/service/example_service_test.go`
- **Cases:**
  - Happy path: create, get, list, update, delete.
  - Error propagation: repository returns `ErrNotFound` → service returns it unwrapped.
  - Tenant isolation: calls always pass `tenantID` from context.

### Integration tests (`//go:build integration`)

- **File:** `internal/platform/repository/example_repo_test.go`
- **Cases:**
  - Insert and retrieve by ID.
  - Soft delete does not appear in list queries.
  - Cross-tenant queries return zero rows.

---

## 8. Open Questions

| # | Question                                            | Owner | Resolution |
| - | --------------------------------------------------- | ----- | ---------- |
| 1 | Should soft-deleted records be purged after N days? | —     | —          |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| YYYY-MM-DD | —      | Task created from roadmap |