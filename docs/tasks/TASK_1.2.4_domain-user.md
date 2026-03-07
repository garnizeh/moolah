# Task 1.2.4 — Domain User Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `User` domain entity and `UserRepository` interface in `internal/domain/user.go`. A `User` belongs to exactly one `Tenant` (household) and has a `Role` that governs what they can do within that tenant.

---

## 2. Context & Motivation

Users are the actors of the system. Every authenticated request carries a `user_id` and `tenant_id` extracted from the PASETO token. The domain entity wraps the sqlc model into a clean business type and the `UserRepository` interface allows the service and auth layers to be fully tested without the database.

---

## 3. Scope

### In scope

- [ ] `User` struct with all business-relevant fields.
- [ ] `CreateUserInput` and `UpdateUserInput` value objects.
- [ ] `UserRepository` interface covering CRUD + lookup-by-email.
- [ ] `UpdateLastLogin` method on the repository (called after successful OTP verification).

### Out of scope

- Concrete repository implementation (Task 1.3.2).
- Service layer (Task 1.4.2).
- HTTP handlers (Task 1.5.5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                      | Purpose                             |
| ------ | ------------------------- | ----------------------------------- |
| CREATE | `internal/domain/user.go` | Entity, input types, repo interface |

### Key interfaces / types

```go
// User is a person who belongs to a Tenant (household).
type User struct {
    ID          string
    TenantID    string
    Email       string
    Name        string
    Role        Role
    LastLoginAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

type CreateUserInput struct {
    TenantID string `validate:"required"`
    Email    string `validate:"required,email"`
    Name     string `validate:"required,min=2,max=100"`
    Role     Role   `validate:"required,oneof=admin member"`
}

type UpdateUserInput struct {
    Name *string `validate:"omitempty,min=2,max=100"`
    Role *Role   `validate:"omitempty,oneof=admin member"`
}

// UserRepository defines persistence operations for users.
type UserRepository interface {
    Create(ctx context.Context, tenantID string, input CreateUserInput) (*User, error)
    GetByID(ctx context.Context, tenantID, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    ListByTenant(ctx context.Context, tenantID string) ([]User, error)
    Update(ctx context.Context, tenantID, id string, input UpdateUserInput) (*User, error)
    UpdateLastLogin(ctx context.Context, id string) error
    Delete(ctx context.Context, tenantID, id string) error
}
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/users.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.5.

### Error cases to handle

| Scenario              | Sentinel Error           | HTTP Status |
| --------------------- | ------------------------ | ----------- |
| User not found        | `domain.ErrNotFound`     | `404`       |
| Email already exists  | `domain.ErrConflict`     | `409`       |
| Tenant mismatch       | `domain.ErrForbidden`    | `403`       |
| Invalid input         | `domain.ErrInvalidInput` | `422`       |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] `UserRepository` interface is defined in `internal/domain/user.go`.
- [ ] `User` struct uses `time.Time` / `*time.Time` (not `pgtype`).
- [ ] `GetByEmail` does **not** require a `tenantID` (used during OTP flow before auth).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | 🔵 backlog |
| Task 1.2.2 — `domain/role.go`    | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A (pure types/interfaces — tested via service layer in 1.4.2)

### Integration tests (`//go:build integration`)

Covered by Task 1.3.2 repository integration tests.

---

## 8. Open Questions

| # | Question                                                    | Owner | Resolution |
| - | ----------------------------------------------------------- | ----- | ---------- |
| 1 | Should a sysadmin user be scoped to a tenant or global?     | —     | Global — `tenant_id` is empty string for sysadmin. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
