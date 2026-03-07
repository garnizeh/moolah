# Task 1.3.6 — Repository: Transaction

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `TransactionRepository` in `internal/platform/repository/transaction_repo.go` using the sqlc-generated code. Transactions are the most frequently accessed entity in the system. The `List` method must support efficient filtered queries (by account, category, type, date range) with pagination.

---

## 2. Context & Motivation

The `TransactionRepository` interface is defined in `internal/domain/transaction.go` (Task 1.2.9). Transactions drive all financial reporting: cash flow summaries, account balance history, and category breakdowns. The repository must translate the `ListTransactionsParams` struct into the correct SQL parameters while maintaining strict tenant isolation. All amounts are `int64` cents — no floating-point mapping allowed.

---

## 3. Scope

### In scope

- [ ] Concrete `transactionRepo` struct implementing `domain.TransactionRepository`.
- [ ] Constructor `NewTransactionRepository(q *sqlc.Queries) domain.TransactionRepository`.
- [ ] Mapping functions between `sqlc.Transaction` and `domain.Transaction`.
- [ ] `List` method correctly applies all `ListTransactionsParams` filters.
- [ ] Error translation: `pgx.ErrNoRows` → `domain.ErrNotFound`, FK violation → `domain.ErrNotFound`.

### Out of scope

- Balance recalculation (service layer, Task 1.4.5).
- Audit event emission (service layer, Task 1.4.5).
- HTTP handlers (Task 1.5.8).
- Installment / master purchase logic (Phase 2).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                               | Purpose                             |
| ------ | -------------------------------------------------- | ----------------------------------- |
| CREATE | `internal/platform/repository/transaction_repo.go` | Concrete TransactionRepository impl |

### Key interfaces / types

```go
type transactionRepo struct {
    q *sqlc.Queries
}

func NewTransactionRepository(q *sqlc.Queries) domain.TransactionRepository {
    return &transactionRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/transactions.sql` (Task 1.1.7/1.1.8):

| Query name                  | sqlc mode | Used by          |
| --------------------------- | --------- | ---------------- |
| `CreateTransaction`         | `:one`    | `Create`         |
| `GetTransactionByID`        | `:one`    | `GetByID`        |
| `ListTransactions`          | `:many`   | `List`           |
| `UpdateTransaction`         | `:one`    | `Update`         |
| `SoftDeleteTransaction`     | `:exec`   | `Delete`         |

### Error cases to handle

| Scenario                    | pgx Error                | Domain Error           |
| --------------------------- | ------------------------ | ---------------------- |
| Row not found               | `pgx.ErrNoRows`          | `domain.ErrNotFound`   |
| FK violation (account/cat)  | `23503` fk_violation     | `domain.ErrNotFound`   |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] Struct implements `domain.TransactionRepository` (verified by compiler).
- [ ] Every SQL query enforces `tenant_id` isolation and `deleted_at IS NULL`.
- [ ] `AmountCents` maps to/from `int64` with no float conversion.
- [ ] `List` correctly applies all `ListTransactionsParams` filters (nil date pointers are ignored).
- [ ] All pgx errors are translated to domain sentinel errors.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                            | Type       | Status     |
| ------------------------------------- | ---------- | ---------- |
| Task 1.2.9 — `domain/transaction.go`  | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files         | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate            | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests        | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — repository implementations are tested via integration tests.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — specifically:

- Create transaction and retrieve by ID.
- `List` with date range filter returns only matching rows.
- `List` with account filter returns only that account's transactions.
- Cross-tenant transaction lookup returns `ErrNotFound`.
- Soft delete removes transaction from list queries.

---

## 8. Open Questions

| # | Question                                                                     | Owner | Resolution |
| - | ---------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `List` use cursor-based or offset pagination?                         | —     | Offset pagination for Phase 1; cursor-based deferred to Phase 5. |
| 2 | Should filters be AND-combined or OR-combined?                               | —     | Always AND — narrowing filters only. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
