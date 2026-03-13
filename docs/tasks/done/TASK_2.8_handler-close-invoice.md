# Task 2.8 — Handler: `POST /v1/accounts/{id}/close-invoice` (Manual Trigger)

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › HTTP Handler Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-12

---

## 1. Summary

Expose the `InvoiceCloser` service (Task 2.7) via a single HTTP endpoint: `POST /v1/accounts/{id}/close-invoice`. This allows users to manually trigger invoice closing for a specific credit card account, returning a summary of how many instalments were materialised.

---

## 2. Context & Motivation

While a scheduled ticker will eventually automate invoice closing (Phase 5), the manual trigger is the MVP control surface. It follows the same handler pattern as Phase 1: auth middleware provides `tenantID`, the path parameter provides `accountID`. The endpoint validates that the account is `credit_card` type (delegating to `InvoiceCloser.CloseInvoice`), runs the closing algorithm, and returns a `204 No Content` or a `200 OK` with a summary body.

---

## 3. Scope

### In scope

- [x] `CloseInvoice` handler method on `AccountHandler` (extends existing `account_handler.go`), or a new `InvoiceCloserHandler`.
- [x] Route `POST /v1/accounts/{id}/close-invoice` registered in `internal/server/routes.go`.
- [x] Optional request body with `closing_date` (defaults to `time.Now()` if omitted).
- [x] Response body with `processed_count` and any partial errors.
- [x] Idempotency middleware applied (idempotency key scoped to `userID + accountID + date`).
- [x] Swaggo annotations.
- [x] Unit tests.

### Out of scope

- Automated scheduler/ticker (Phase 5).
- `InvoiceCloser` business logic (Task 2.7).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                          | Purpose                                     |
| ------ | --------------------------------------------- | ------------------------------------------- |
| MODIFY | `internal/handler/account_handler.go`         | Add `CloseInvoice` handler method           |
| MODIFY | `internal/handler/account_handler_test.go`    | Add unit tests for `CloseInvoice`           |
| MODIFY | `internal/server/routes.go`                   | Register `POST .../{id}/close-invoice`      |

### API endpoint

| Method | Path                                   | Auth Required | Idempotency Key | Description                             |
| ------ | -------------------------------------- | ------------- | --------------- | --------------------------------------- |
| `POST` | `/v1/accounts/{id}/close-invoice`      | ✅ Bearer     | ✅ Required     | Manually trigger invoice closing        |

### Request / Response shapes

```go
// POST /v1/accounts/{id}/close-invoice — optional body
type CloseInvoiceRequest struct {
    // ClosingDate defaults to today if omitted.
    ClosingDate *string `json:"closing_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// POST /v1/accounts/{id}/close-invoice — 200 OK response
type CloseInvoiceResponse struct {
    AccountID      string   `json:"account_id"`
    ProcessedCount int      `json:"processed_count"`
    Errors         []string `json:"errors,omitempty"` // partial failures (best-effort)
}
```

### Error cases to handle

| Scenario                              | Sentinel Error           | HTTP Status |
| ------------------------------------- | ------------------------ | ----------- |
| Account not found / wrong tenant      | `domain.ErrNotFound`     | `404`       |
| Account is not `credit_card` type     | `domain.ErrInvalidInput` | `422`       |
| No pending master purchases           | —                        | `200 OK` with `processed_count=0` |
| Partial DB failure                    | —                        | `200 OK` with non-empty `errors` array |

---

## 5. Acceptance Criteria

- [ ] Endpoint registered with Go 1.22 routing syntax: `POST /v1/accounts/{id}/close-invoice`.
- [ ] `tenantID` extracted from context; `accountID` from path parameter.
- [ ] `closing_date` defaults to `time.Now().UTC()` when omitted.
- [ ] Idempotency middleware wired on this route.
- [ ] Response returns `200 OK` with `CloseInvoiceResponse` even when `processed_count=0`.
- [ ] Swaggo annotations present and `swag init` succeeds.
- [ ] Unit tests cover: happy path, missing account, non-credit-card account, partial failures.
- [ ] Test coverage for new code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.8 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                    | Type     | Status       |
| --------------------------------------------- | -------- | ------------ |
| Task 2.7 — `InvoiceCloser` service            | Upstream | 🔵 backlog   |
| `handler/account_handler.go` (Phase 1)        | Upstream | ✅ done      |
| `platform/middleware/auth.go`                 | Upstream | ✅ done      |
| `platform/middleware/idempotency.go`          | Upstream | ✅ done      |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/account_handler_test.go` (extended)
- **Cases:**
  - `POST /v1/accounts/{id}/close-invoice` with no body → `closing_date` defaults to today, 200 OK.
  - `POST` with explicit `closing_date` → passed to `InvoiceCloser` unchanged.
  - Account not found → 404.
  - Account type `checking` → 422.
  - `ProcessedCount=3` returned correctly.
  - Partial errors from `InvoiceCloser` appear in response `errors` array.

### Integration tests (`//go:build integration`)

Covered by Task 2.12 — full invoice closing flow.

---

## 8. Open Questions

| # | Question                                                                           | Owner | Resolution |
| - | ---------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should the endpoint return `204 No Content` or `200 OK`?                          | —     | `200 OK` — the `processed_count` field gives meaningful feedback. |
| 2 | Should partial errors cause a `207 Multi-Status` response?                        | —     | No — `200 OK` with `errors` array keeps the response simple. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
