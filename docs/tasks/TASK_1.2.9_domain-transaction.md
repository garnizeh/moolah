# Task 1.2.9 — Domain Transaction Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `Transaction` domain entity and `TransactionRepository` interface in `internal/domain/transaction.go`. Transactions are the core financial records — debits, credits, and transfers — linking an account, a category, and an optional master purchase.

---

## 2. Context & Motivation

Transactions are the most frequently read and written entity in the system. The domain entity separates the sqlc model from the business layer, and the `TransactionRepository` interface makes the service layer fully unit-testable. All monetary amounts are stored in cents (`int64`) to avoid floating-point drift.

---

## 3. Scope

### In scope

- [ ] `Transaction` struct and `TransactionType` constants.
- [ ] `CreateTransactionInput` and `UpdateTransactionInput` value objects.
- [ ] `ListTransactionsParams` for filter/pagination support.
- [ ] `TransactionRepository` interface: CRUD + `ListByAccount` with filters.

### Out of scope

- Credit card installment transactions (Phase 2 — `MasterPurchaseID` field is present but populated only in Phase 2 flows).
- Concrete repository implementation (Task 1.3.6).
- Service layer (Task 1.4.5).
- HTTP handlers (Task 1.5.8).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                            | Purpose                              |
| ------ | ------------------------------- | ------------------------------------ |
| CREATE | `internal/domain/transaction.go` | Entity, input types, repo interface |

### Key interfaces / types

```go
// TransactionType mirrors the DB enum.
type TransactionType string

const (
    TransactionTypeIncome   TransactionType = "income"
    TransactionTypeExpense  TransactionType = "expense"
    TransactionTypeTransfer TransactionType = "transfer"
)

// Transaction is a single financial event on an account.
type Transaction struct {
    ID               string
    TenantID         string
    AccountID        string
    CategoryID       string
    UserID           string
    MasterPurchaseID string      // Empty for regular transactions; set in Phase 2
    Description      string
    AmountCents      int64       // Always in cents; positive value; type determines direction
    Type             TransactionType
    OccurredAt       time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        *time.Time
}

type CreateTransactionInput struct {
    AccountID        string          `validate:"required"`
    CategoryID       string          `validate:"required"`
    Description      string          `validate:"required,min=1,max=255"`
    AmountCents      int64           `validate:"required,gt=0"`
    Type             TransactionType `validate:"required,oneof=income expense transfer"`
    OccurredAt       time.Time       `validate:"required"`
    MasterPurchaseID string          `validate:"omitempty"`
}

type UpdateTransactionInput struct {
    CategoryID  *string          `validate:"omitempty"`
    Description *string          `validate:"omitempty,min=1,max=255"`
    AmountCents *int64           `validate:"omitempty,gt=0"`
    OccurredAt  *time.Time       `validate:"omitempty"`
}

type ListTransactionsParams struct {
    AccountID   string
    CategoryID  string
    Type        TransactionType
    StartDate   *time.Time
    EndDate     *time.Time
    Limit       int32
    Offset      int32
}

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
    Create(ctx context.Context, tenantID string, input CreateTransactionInput) (*Transaction, error)
    GetByID(ctx context.Context, tenantID, id string) (*Transaction, error)
    List(ctx context.Context, tenantID string, params ListTransactionsParams) ([]Transaction, error)
    Update(ctx context.Context, tenantID, id string, input UpdateTransactionInput) (*Transaction, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/transactions.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.8.

### Error cases to handle

| Scenario                  | Sentinel Error           | HTTP Status |
| ------------------------- | ------------------------ | ----------- |
| Transaction not found     | `domain.ErrNotFound`     | `404`       |
| Tenant mismatch           | `domain.ErrForbidden`    | `403`       |
| Account not found         | `domain.ErrNotFound`     | `404`       |
| Invalid amount (≤ 0)      | `domain.ErrInvalidInput` | `422`       |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] `TransactionRepository` interface is defined in `internal/domain/transaction.go`.
- [ ] `AmountCents` is `int64` — no `float` anywhere.
- [ ] `Transaction` struct uses `time.Time` / `*time.Time` (not `pgtype`).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                        | Type     | Status     |
| --------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`   | Upstream | 🔵 backlog |
| Task 1.2.7 — `domain/account.go`  | Upstream | 🔵 backlog |
| Task 1.2.8 — `domain/category.go` | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A (pure types/interfaces — tested via service layer in 1.4.5)

### Integration tests (`//go:build integration`)

Covered by Task 1.3.6 repository integration tests.

---

## 8. Open Questions

| # | Question                                                                  | Owner | Resolution |
| - | ------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should transfers link two transactions (debit + credit) or be a single row? | —   | Single row with `type=transfer`; balance recalc handles both accounts. |
| 2 | Should `AmountCents` always be positive with direction derived from type? | —     | Yes — always positive; type determines sign in balance calculation. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
