# Task 1.2.7 — Domain Account Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `Account` domain entity and `AccountRepository` interface in `internal/domain/account.go`. An `Account` represents a financial account (checking, savings, credit card, or investment) owned by a `User` within a `Tenant`.

---

## 2. Context & Motivation

Accounts are the foundational finance entities — every Transaction is associated with an Account. The domain entity wraps sqlc fields into clean Go types and the `AccountRepository` interface enables full mock-based unit testing of the account service (Task 1.4.3) without running a database.

---

## 3. Scope

### In scope

- [ ] `Account` struct and `AccountType` constants.
- [ ] `CreateAccountInput` and `UpdateAccountInput` value objects.
- [ ] `AccountRepository` interface: CRUD + balance recalculation trigger.
- [ ] Balance must use `int64` (cents) — never `float64`.

### Out of scope

- Concrete repository implementation (Task 1.3.4).
- Service layer (Task 1.4.3).
- HTTP handlers (Task 1.5.6).
- Investment account specifics (Phase 3).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                         | Purpose                              |
| ------ | ---------------------------- | ------------------------------------ |
| CREATE | `internal/domain/account.go` | Entity, input types, repo interface  |

### Key interfaces / types

```go
// AccountType mirrors the DB enum.
type AccountType string

const (
    AccountTypeChecking   AccountType = "checking"
    AccountTypeSavings    AccountType = "savings"
    AccountTypeCreditCard AccountType = "credit_card"
    AccountTypeInvestment AccountType = "investment"
)

// Account represents a financial account owned by a user within a household.
type Account struct {
    ID           string
    TenantID     string
    UserID       string
    Name         string
    Type         AccountType
    Currency     string      // ISO 4217, e.g. "BRL", "USD"
    BalanceCents int64       // Always in cents; never float
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time
}

type CreateAccountInput struct {
    UserID   string      `validate:"required"`
    Name     string      `validate:"required,min=1,max=100"`
    Type     AccountType `validate:"required,oneof=checking savings credit_card investment"`
    Currency string      `validate:"required,len=3"`
}

type UpdateAccountInput struct {
    Name     *string      `validate:"omitempty,min=1,max=100"`
    Currency *string      `validate:"omitempty,len=3"`
}

// AccountRepository defines persistence operations for accounts.
type AccountRepository interface {
    Create(ctx context.Context, tenantID string, input CreateAccountInput) (*Account, error)
    GetByID(ctx context.Context, tenantID, id string) (*Account, error)
    ListByTenant(ctx context.Context, tenantID string) ([]Account, error)
    ListByUser(ctx context.Context, tenantID, userID string) ([]Account, error)
    Update(ctx context.Context, tenantID, id string, input UpdateAccountInput) (*Account, error)
    UpdateBalance(ctx context.Context, tenantID, id string, newBalanceCents int64) error
    Delete(ctx context.Context, tenantID, id string) error
}
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/accounts.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.6.

### Error cases to handle

| Scenario                 | Sentinel Error           | HTTP Status |
| ------------------------ | ------------------------ | ----------- |
| Account not found        | `domain.ErrNotFound`     | `404`       |
| Tenant mismatch          | `domain.ErrForbidden`    | `403`       |
| Duplicate account name   | `domain.ErrConflict`     | `409`       |
| Invalid input            | `domain.ErrInvalidInput` | `422`       |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] `AccountRepository` interface is defined in `internal/domain/account.go`.
- [ ] `BalanceCents` is `int64` — no `float` anywhere.
- [ ] `Account` struct uses `time.Time` / `*time.Time` (not `pgtype`).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A (pure types/interfaces — tested via service layer in 1.4.3)

### Integration tests (`//go:build integration`)

Covered by Task 1.3.4 repository integration tests.

---

## 8. Open Questions

| # | Question                                                         | Owner | Resolution |
| - | ---------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `UpdateBalance` be exposed on the repo or only via service? | —   | Expose on repo; service ensures transactional consistency. |
| 2 | Should initial balance be 0 always or configurable at creation?  | —     | Configurable — some accounts may be imported with existing balance. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
