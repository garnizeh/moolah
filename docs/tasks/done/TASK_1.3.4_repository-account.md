# Task 1.3.4 — Repository: Account

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Implement `AccountRepository` in `internal/platform/repository/account_repo.go` using the sqlc-generated code. Accounts are the foundational financial entities — every transaction references an account. The repository enforces tenant isolation on all operations and provides a dedicated `UpdateBalance` method for the service layer to call after recording transactions.

---

## 2. Context & Motivation

The `AccountRepository` interface is defined in `internal/domain/account.go` (Task 1.2.7). Keeping balance updates as a discrete repository method (rather than recalculating in SQL triggers) ensures the service layer can maintain transactional consistency: create transaction → update balance within the same `pgx.Tx`. The mapping layer converts between the sqlc integer balance and the domain `BalanceCents int64` field.

---

## 3. Scope

### In scope

- [x] Concrete `accountRepo` struct implementing `domain.AccountRepository`.
- [x] Constructor `NewAccountRepository(q *sqlc.Queries) domain.AccountRepository`.
- [x] Mapping functions between `sqlc.Account` and `domain.Account`.
- [x] Error translation: `pgx.ErrNoRows` → `domain.ErrNotFound`, unique violation → `domain.ErrConflict`.
- [x] All queries include `WHERE tenant_id = $1 AND deleted_at IS NULL`.

### Out of scope

- Balance recalculation logic (service layer, Task 1.4.3).
- HTTP handlers (Task 1.5.6).
- Admin cross-tenant queries (Task 1.3.8).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                            | Purpose                          |
| ------ | ----------------------------------------------- | -------------------------------- |
| CREATE | `internal/platform/repository/account_repo.go` | Concrete AccountRepository impl  |
| CREATE | `internal/platform/repository/account_repo_test.go` | Unit tests for AccountRepository |
| MODIFY | `internal/platform/db/queries/accounts.sql`    | Added ListAccountsByUser query   |
| MODIFY | `internal/testutil/mocks/querier.go`           | Added mock ListAccountsByUser    |

### Key interfaces / types

```go
type accountRepo struct {
    q sqlc.Querier
}

func NewAccountRepository(q sqlc.Querier) domain.AccountRepository {
    return &accountRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/accounts.sql` (Task 1.1.7/1.1.8):

| Query name              | sqlc mode | Used by           |
| ----------------------- | --------- | ----------------- |
| `CreateAccount`         | `:one`    | `Create`          |
| `GetAccountByID`        | `:one`    | `GetByID`         |
| `ListAccountsByTenant`  | `:many`   | `ListByTenant`    |
| `ListAccountsByUser`    | `:many`   | `ListByUser`      |
| `UpdateAccount`         | `:one`    | `Update`          |
| `UpdateAccountBalance`  | `:exec`   | `UpdateBalance`   |
| `SoftDeleteAccount`     | `:exec`   | `Delete`          |

### Error cases to handle

| Scenario               | pgx Error                | Domain Error           |
| ---------------------- | ------------------------ | ---------------------- |
| Row not found          | `pgx.ErrNoRows`          | `domain.ErrNotFound`   |
| Duplicate account name | `23505` unique_violation | `domain.ErrConflict`   |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] Struct implements `domain.AccountRepository` (verified by compiler).
- [x] Every SQL query enforces `tenant_id` isolation and `deleted_at IS NULL`.
- [x] `BalanceCents` is correctly mapped to/from the database column (no float conversion).
- [x] All pgx errors are translated to domain sentinel errors.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./......` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                        | Type       | Status     |
| --------------------------------- | ---------- | ---------- |
| Task 1.2.7 — `domain/account.go`  | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files     | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate        | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests    | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

Implemented in `internal/platform/repository/account_repo_test.go` using testify/mock.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — specifically:

- Create account and retrieve by ID.
- Verify tenant isolation (cross-tenant lookup returns `ErrNotFound`).
- Update balance and verify new value.
- Soft delete and verify it no longer appears in list queries.

---

## 8. Open Questions

| # | Question                                                                    | Owner | Resolution |
| - | --------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `UpdateBalance` be transactional with `Create` transaction in service? | —   | Yes — the service layer wraps both calls in a `pgx.Tx`. The repo method accepts a `*sqlc.Queries` that can be backed by a transaction. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | GitHub Copilot | Implemented AccountRepository and unit tests |
