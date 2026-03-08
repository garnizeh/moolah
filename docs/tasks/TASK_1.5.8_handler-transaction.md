# Task 1.5.8 — `handler/transaction_handler.go` — full CRUD + list with filters

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Implement the transaction HTTP handler in `internal/handler/transaction_handler.go`. It exposes full CRUD endpoints plus a filtered list endpoint for financial transactions. This is the core financial data entry point and the most complex handler in Phase 1.

---

## 2. Context & Motivation

The `TransactionService` is fully implemented (Task 1.4.5) but has no HTTP entry point. Transactions are the heartbeat of the cash flow system — creating, updating, and deleting them each mutate the parent account's balance. The handler must faithfully delegate to the service and surface all relevant error codes. See `docs/ARCHITECTURE.md` and roadmap item 1.5.8.

---

## 3. Scope

### In scope

- [ ] `internal/handler/transaction_handler.go` — `TransactionHandler` struct + 5 HTTP handler methods.
- [ ] `GET /v1/transactions` — list transactions with optional filters.
- [ ] `POST /v1/transactions` — create a new transaction.
- [ ] `GET /v1/transactions/{id}` — get a single transaction by ID.
- [ ] `PATCH /v1/transactions/{id}` — partial update (amount, description, category, date).
- [ ] `DELETE /v1/transactions/{id}` — soft delete + balance revert.
- [ ] Query parameter filters on `GET /v1/transactions`: `account_id`, `category_id`, `type`, `from` (date), `to` (date), `limit`, `offset`.
- [ ] Unit tests in `internal/handler/transaction_handler_test.go`.

### Out of scope

- Master Purchase (installment) transactions — Phase 2 (Task 2.6).
- Batch import / CSV upload — not in Phase 1.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                   |
| ------ | ---------------------------------------------- | ----------------------------------------- |
| CREATE | `internal/handler/transaction_handler.go`      | HTTP handler for transaction endpoints    |
| CREATE | `internal/handler/transaction_handler_test.go` | Unit tests with mocked TransactionService |

### Request / Response types

```go
type CreateTransactionRequest struct {
    AccountID   string                  `json:"account_id"   validate:"required"`
    CategoryID  string                  `json:"category_id"  validate:"required"`
    Type        domain.TransactionType  `json:"type"         validate:"required"`
    AmountCents int64                   `json:"amount_cents" validate:"required,min=1"`
    Description string                  `json:"description"  validate:"required,min=1,max=255"`
    OccurredAt  time.Time               `json:"occurred_at"  validate:"required"`
}

type UpdateTransactionRequest struct {
    AmountCents *int64   `json:"amount_cents" validate:"omitempty,min=1"`
    Description *string  `json:"description"  validate:"omitempty,min=1,max=255"`
    CategoryID  *string  `json:"category_id"  validate:"omitempty"`
}

type ListTransactionsQuery struct {
    AccountID  string `schema:"account_id"`
    CategoryID string `schema:"category_id"`
    Type       string `schema:"type"`
    From       string `schema:"from"`
    To         string `schema:"to"`
    Limit      int    `schema:"limit"`
    Offset     int    `schema:"offset"`
}
```

### API endpoints

| Method | Path                        | Auth Required | Description                           |
| ------ | --------------------------- | ------------- | ------------------------------------- |
| GET    | `/v1/transactions`          | ✅ Bearer     | List with filters                     |
| POST   | `/v1/transactions`          | ✅ Bearer     | Create a transaction                  |
| GET    | `/v1/transactions/{id}`     | ✅ Bearer     | Get transaction by ID                 |
| PATCH  | `/v1/transactions/{id}`     | ✅ Bearer     | Update transaction                    |
| DELETE | `/v1/transactions/{id}`     | ✅ Bearer     | Soft-delete + revert balance          |

### Error cases to handle

| Scenario                         | Sentinel Error           | HTTP Status |
| -------------------------------- | ------------------------ | ----------- |
| Not found                        | `domain.ErrNotFound`     | `404`       |
| Validation failure               | —                        | `422`       |
| Category type mismatch           | `domain.ErrInvalidInput` | `422`       |
| Account not found                | `domain.ErrNotFound`     | `404`       |
| Forbidden                        | `domain.ErrForbidden`    | `403`       |

---

## 5. Acceptance Criteria

- [ ] All 5 endpoints decode, validate, and call the service correctly.
- [ ] `tenant_id` is always sourced from context.
- [ ] `GET /v1/transactions` supports all query parameters listed above.
- [ ] All domain error sentinels map to the correct HTTP status codes.
- [ ] `DELETE` returns `204 No Content` on success.
- [ ] Unit tests cover all happy paths and all error cases.
- [ ] Test coverage for handler ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                 | Type     | Status  |
| ------------------------------------------ | -------- | ------- |
| Task 1.4.5 — `service/transaction_service` | Upstream | ✅ done |
| Task 1.5.7 — `handler/category_handler`    | Related  | 🔵 backlog |
| Task 1.5.6 — `handler/account_handler`     | Related  | 🔵 backlog |
| Task 1.1.9 — Auth middleware               | Upstream | ✅ done |
| `domain.TransactionService` interface      | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/transaction_handler_test.go`
- **Cases:**
  - `List`: no filters → returns all → `200 OK`.
  - `List`: with `account_id` filter → calls service with correct params.
  - `Create`: valid → `201 Created` with resource JSON.
  - `Create`: invalid body → `422`.
  - `Create`: category type mismatch → `422`.
  - `GetByID`: found → `200 OK`.
  - `GetByID`: not found → `404`.
  - `Update`: amount changed → `200 OK`.
  - `Delete`: success → `204 No Content`.
  - `Delete`: not found → `404`.

---

## 8. Open Questions

| # | Question                                                    | Owner | Resolution |
| - | ----------------------------------------------------------- | ----- | ---------- |
| 1 | Should the list response include total count for pagination? | —    | Yes — include `{"data": [...], "total": N}` envelope. |
| 2 | Default page size for list?                                 | —     | `limit=50`, `offset=0`. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
