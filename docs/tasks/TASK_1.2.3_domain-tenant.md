# Task 1.2.3 — Domain Tenant Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `Tenant` domain entity and `TenantRepository` interface in `internal/domain/tenant.go`. This establishes the contract that the concrete repository implementation (Task 1.3.1) must satisfy and provides the canonical Go type used across all layers of the application.

---

## 2. Context & Motivation

The `Tenant` represents a household — the root multi-tenancy entity. Every other entity (users, accounts, transactions) hangs off a tenant via `tenant_id`. The domain entity translates the raw sqlc model into a clean business type, decoupling the service and handler layers from the DB schema. The `TenantRepository` interface enables full mock-based unit testing of the service layer without touching the database.

---

## 3. Scope

### In scope

- [x] `Tenant` struct with all business-relevant fields.
- [x] `CreateTenantInput` and `UpdateTenantInput` value objects.
- [x] `TenantRepository` interface covering CRUD operations.
- [x] Unit tests for any helper methods on `Tenant`.

### Out of scope

- Concrete repository implementation (Task 1.3.1).
- Service layer (Task 1.4.2).
- HTTP handlers (Task 1.5.5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                        | Purpose                              |
| ------ | --------------------------- | ------------------------------------ |
| CREATE | `internal/domain/tenant.go` | Entity, input types, repo interface  |

### Key interfaces / types

```go
// TenantPlan mirrors the DB enum.
type TenantPlan string

const (
    TenantPlanFree    TenantPlan = "free"
    TenantPlanBasic   TenantPlan = "basic"
    TenantPlanPremium TenantPlan = "premium"
)

// Tenant is the root multi-tenancy entity (a household).
type Tenant struct {
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
    ID        string
    Name      string
    Plan      TenantPlan
}

type CreateTenantInput struct {
    Name string `validate:"required,min=2,max=100"`
}

type UpdateTenantInput struct {
    Name *string `validate:"omitempty,min=2,max=100"`
    Plan *TenantPlan
}

// TenantRepository defines persistence operations for tenants.
// All write operations are performed by sysadmin only; tenants cannot self-provision.
type TenantRepository interface {
    Create(ctx context.Context, input CreateTenantInput) (*Tenant, error)
    GetByID(ctx context.Context, id string) (*Tenant, error)
    List(ctx context.Context) ([]Tenant, error)
    Update(ctx context.Context, id string, input UpdateTenantInput) (*Tenant, error)
    Delete(ctx context.Context, id string) error
}
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/tenants.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.5.

### Error cases to handle

| Scenario                | Sentinel Error        | HTTP Status |
| ----------------------- | --------------------- | ----------- |
| Tenant not found        | `domain.ErrNotFound`  | `404`       |
| Name already taken      | `domain.ErrConflict`  | `409`       |
| Invalid input           | `domain.ErrInvalidInput` | `422`    |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] `TenantRepository` interface is defined in `internal/domain/tenant.go`.
- [x] `Tenant` struct uses `time.Time` (not `pgtype.Timestamptz`) for clean domain isolation.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** N/A (pure types/interfaces — tested via service layer in 1.4.2)

### Integration tests (`//go:build integration`)

- Covered by Task 1.3.1 repository integration tests.

---

## 8. Open Questions

| # | Question                                                             | Owner | Resolution |
| - | -------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `TenantPlan` type be re-used from sqlc or redefined in domain? | —    | Redefine in domain to avoid sqlc coupling. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | Agent  | Created entity and repo interface. Applied fieldalignment fix. |

| 2026-03-07 | —      | Task created from roadmap |
