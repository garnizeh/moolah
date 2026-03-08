# Task 1.4.5 — `service/transaction_service.go` + unit tests

> **Roadmap Ref:** Phase 1 — MVP › 1.4 Service Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Implement `TransactionService`, the most critical service in Phase 1. It orchestrates `TransactionRepository` and `AccountRepository` to guarantee that every financial event atomically creates a transaction record AND updates the account balance. All balance mutation logic, validation, and audit trail writes live here.

---

## 2. Context & Motivation

Transactions are the core of the cash flow feature. A transaction is meaningless without a corresponding balance change on its account, and a balance change without a transaction is a financial integrity violation. The service must enforce this atomicity at the application level (Phase 1 uses sequential calls; Phase 3+ may introduce DB transactions). See `docs/ARCHITECTURE.md` and roadmap item 1.4.5.

---

## 3. Scope

### In scope

- [ ] `internal/service/transaction_service.go` — concrete `TransactionService` struct.
- [ ] `internal/domain/transaction.go` — add `TransactionService` interface definition.
- [ ] `Create(ctx, tenantID string, input CreateTransactionInput) (*Transaction, error)`:
  - Verify account belongs to tenant.
  - Verify category belongs to tenant and type matches transaction type.
  - Persist transaction via `TransactionRepository.Create`.
  - Update account balance via `AccountRepository.UpdateBalance` (apply delta based on type: income = +, expense = -, transfer = − source).
  - Write audit log `create`.
- [ ] `GetByID(ctx, tenantID, id string) (*Transaction, error)` — tenant-scoped fetch.
- [ ] `List(ctx, tenantID string, params ListTransactionsParams) ([]Transaction, error)` — with filters.
- [ ] `Update(ctx, tenantID, id string, input UpdateTransactionInput) (*Transaction, error)`:
  - If `AmountCents` changes: revert old delta, apply new delta on account balance.
  - Write audit log `update` with old/new values.
- [ ] `Delete(ctx, tenantID, id string) error`:
  - Revert the transaction's delta from the account balance before soft-deleting.
  - Write audit log `soft_delete`.
- [ ] Full unit tests in `internal/service/transaction_service_test.go`.

### Out of scope

- HTTP handler (`handler/transaction_handler.go`) — Task 1.5.8.
- Transfer between two accounts (debit source + credit destination) — Phase 2; Phase 1 records transfer as single-account expense.
- DB-level transactions (ACID) wrapping the two writes — deferred to Phase 2; Phase 1 uses best-effort sequential writes.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                            | Purpose                                         |
| ------ | ----------------------------------------------- | ----------------------------------------------- |
| MODIFY | `internal/domain/transaction.go`                | Add `TransactionService` interface              |
| CREATE | `internal/service/transaction_service.go`       | Concrete service implementation                 |
| CREATE | `internal/service/transaction_service_test.go`  | Unit tests with mocked deps                     |

### Key interfaces / types

```go
// TransactionService defines the business-logic contract for transaction management.
type TransactionService interface {
    Create(ctx context.Context, tenantID string, input CreateTransactionInput) (*Transaction, error)
    GetByID(ctx context.Context, tenantID, id string) (*Transaction, error)
    List(ctx context.Context, tenantID string, params ListTransactionsParams) ([]Transaction, error)
    Update(ctx context.Context, tenantID, id string, input UpdateTransactionInput) (*Transaction, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### Business rules — balance delta logic

```
income  → balance += AmountCents
expense → balance -= AmountCents
transfer → balance -= AmountCents  (Phase 1: single account debit only)
```

For `Update` when `AmountCents` changes:

```
revert old delta: apply inverse of old type for old amount
apply new delta:  apply new type for new amount
```

For `Delete`:

```
revert delta: apply inverse of transaction type for amount
```

### Error cases

| Scenario                            | Sentinel Error          | Notes                            |
| ----------------------------------- | ----------------------- | -------------------------------- |
| Transaction not found               | `domain.ErrNotFound`    | `GetByID`, `Update`, `Delete`    |
| Account not in tenant               | `domain.ErrNotFound`    | `Create`                         |
| Category not in tenant              | `domain.ErrNotFound`    | `Create`                         |
| Category type mismatch              | `domain.ErrInvalidInput`| `Create` e.g. income tx on expense category |
| Amount ≤ 0                          | validation error        | `Create`, `Update`               |

---

## 5. Acceptance Criteria

- [ ] `TransactionService` interface defined in `internal/domain/transaction.go`.
- [ ] `NewTransactionService` constructor accepts `TransactionRepository`, `AccountRepository`, `CategoryRepository`, and `AuditRepository`.
- [ ] `Create` verifies account and category belong to tenant before persisting.
- [ ] `Create` updates account balance after persisting transaction.
- [ ] `Update` with new `AmountCents` correctly reverts old delta and applies new delta.
- [ ] `Delete` reverts the balance delta before soft-deleting.
- [ ] Audit log written for `create`, `update`, `soft_delete` events.
- [ ] `AmountCents` is always stored as a positive `int64`; direction encoded by `Type`.
- [ ] Unit tests cover all happy paths and all error branches (≥ 80% coverage).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                           | Type     | Status  |
| ------------------------------------ | -------- | ------- |
| Task 1.3.4 — `account_repo.go`       | Upstream | ✅ done |
| Task 1.3.5 — `category_repo.go`      | Upstream | ✅ done |
| Task 1.3.6 — `transaction_repo.go`   | Upstream | ✅ done |
| Task 1.3.7 — `audit_repo.go`         | Upstream | ✅ done |
| Task 1.1.17 — `testutil/mocks`       | Upstream | ✅ done |
| `domain.ErrInvalidInput` sentinel    | Upstream | verify in `domain/errors.go` |

---

## 7. Testing Plan

Unit tests only.

Key scenarios:

- `Create` income → account balance increased → audit log written.
- `Create` expense → account balance decreased → audit log written.
- `Create` with unknown account → `ErrNotFound`.
- `Create` with unknown category → `ErrNotFound`.
- `Create` category type mismatch → `ErrInvalidInput`.
- `Update` with new amount → old delta reverted, new delta applied.
- `Update` without amount change → no balance update.
- `Delete` → balance reverted → audit log `soft_delete`.

---

## 8. Open Questions

| # | Question                                                                             | Owner | Resolution |
| - | ------------------------------------------------------------------------------------ | ----- | ---------- |
| 1 | Should `Create` fail if the balance would go negative (overdraft protection)?        | —     | No for Phase 1; allow negative balances. |
| 2 | Should `Update` allow changing the `AccountID` (moving to a different account)?      | —     | No for Phase 1; `AccountID` is immutable after creation. |
| 3 | Should `Update` allow changing the `Type` of a transaction?                          | —     | No for Phase 1; type is immutable. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
