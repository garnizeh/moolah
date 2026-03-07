# Task 1.3.2 — Repository: User

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `UserRepository` in `internal/platform/repository/user_repo.go` using the sqlc-generated code. The user repository manages household member records and supports both tenant-scoped lookups and the cross-tenant `GetByEmail` used by the auth flow.

---

## 2. Context & Motivation

The `UserRepository` interface is defined in `internal/domain/user.go` (Task 1.2.4). Every operation except `GetByEmail` must be filtered by `tenant_id`. The repository is consumed by both the auth service (for login) and the tenant service (for member management). Error translation from pgx to domain sentinels keeps the service layer clean of database concerns.

---

## 3. Scope

### In scope

- [ ] Concrete `userRepo` struct implementing `domain.UserRepository`.
- [ ] Constructor `NewUserRepository(q *sqlc.Queries) domain.UserRepository`.
- [ ] Mapping functions between `sqlc.User` and `domain.User`.
- [ ] Error translation: `pgx.ErrNoRows` → `domain.ErrNotFound`, unique violation → `domain.ErrConflict`.
- [ ] `GetByEmail` intentionally omits `tenant_id` filter — document this as an auth-flow exception.

### Out of scope

- Service layer (Task 1.4.2).
- HTTP handlers (Task 1.5.5).
- Admin cross-tenant user queries (Task 1.3.8).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                         | Purpose                        |
| ------ | -------------------------------------------- | ------------------------------ |
| CREATE | `internal/platform/repository/user_repo.go` | Concrete UserRepository impl   |

### Key interfaces / types

```go
type userRepo struct {
    q *sqlc.Queries
}

func NewUserRepository(q *sqlc.Queries) domain.UserRepository {
    return &userRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/users.sql` (Task 1.1.7/1.1.8):

| Query name            | sqlc mode | Used by             |
| --------------------- | --------- | ------------------- |
| `CreateUser`          | `:one`    | `Create`            |
| `GetUserByID`         | `:one`    | `GetByID`           |
| `GetUserByEmail`      | `:one`    | `GetByEmail`        |
| `ListUsersByTenant`   | `:many`   | `ListByTenant`      |
| `UpdateUser`          | `:one`    | `Update`            |
| `UpdateUserLastLogin` | `:exec`   | `UpdateLastLogin`   |
| `SoftDeleteUser`      | `:exec`   | `Delete`            |

### Error cases to handle

| Scenario              | pgx Error                 | Domain Error          |
| --------------------- | ------------------------- | --------------------- |
| Row not found         | `pgx.ErrNoRows`           | `domain.ErrNotFound`  |
| Duplicate email       | `23505` unique_violation  | `domain.ErrConflict`  |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] Struct implements `domain.UserRepository` (verified by compiler).
- [ ] All methods except `GetByEmail` include `tenant_id` in the SQL query.
- [ ] All pgx errors are translated to domain sentinel errors.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                      | Type       | Status     |
| ------------------------------- | ---------- | ---------- |
| Task 1.2.4 — `domain/user.go`   | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files   | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate      | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests  | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — repository implementations are tested via integration tests.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — all repository integration tests run together in a single Testcontainers suite.

---

## 8. Open Questions

| # | Question                                               | Owner | Resolution |
| - | ------------------------------------------------------ | ----- | ---------- |
| 1 | Should `GetByEmail` return `ErrNotFound` when no user exists for the given email? | — | Yes — the auth service interprets this to determine if a new user should be provisioned. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
