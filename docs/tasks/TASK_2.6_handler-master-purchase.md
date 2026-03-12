# Task 2.6 — Handler: `MasterPurchaseHandler` — CRUD Endpoints

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `internal/handler/master_purchase_handler.go` with full CRUD endpoints for master purchases. The `POST /v1/master-purchases` response must include the runtime-projected instalment schedule alongside the created record (no extra round-trip needed by the client).

---

## 2. Context & Motivation

The handler bridges HTTP and the `MasterPurchaseService` (Task 2.5). Following Phase 1 patterns: extract `tenantID`/`userID` from context (set by the auth middleware), decode and validate the request body using `go-playground/validator`, and map domain errors to HTTP status codes. Swaggo annotations are mandatory for every endpoint.

The `POST` response returning projected instalments is a deliberate UX decision: the client immediately sees the full schedule without a second request, even though the instalments are not yet stored in the DB.

---

## 3. Scope

### In scope

- [ ] `MasterPurchaseHandler` struct with constructor accepting `domain.MasterPurchaseService`.
- [ ] `Create` handler: `POST /v1/master-purchases`.
- [ ] `GetByID` handler: `GET /v1/master-purchases/{id}`.
- [ ] `ListByTenant` handler: `GET /v1/master-purchases`.
- [ ] `ListByAccount` handler: `GET /v1/accounts/{account_id}/master-purchases`.
- [ ] `Update` handler: `PATCH /v1/master-purchases/{id}`.
- [ ] `Delete` handler: `DELETE /v1/master-purchases/{id}`.
- [ ] Route registrations in `internal/server/routes.go`.
- [ ] Idempotency middleware wired on `POST /v1/master-purchases`.
- [ ] Swaggo annotations on all handlers.
- [ ] Unit tests with mocked `MasterPurchaseService` (≥ 80% coverage).

### Out of scope

- `POST /v1/accounts/{id}/close-invoice` (Task 2.8).
- InvoiceCloser internals (Task 2.7).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                             | Purpose                              |
| ------ | ------------------------------------------------ | ------------------------------------ |
| CREATE | `internal/handler/master_purchase_handler.go`    | HTTP handler implementation          |
| CREATE | `internal/handler/master_purchase_handler_test.go` | Unit tests with mocked service     |
| MODIFY | `internal/server/routes.go`                      | Register Phase 2 routes              |

### API endpoints

| Method   | Path                                          | Auth Required | Idempotency Key | Description                                       |
| -------- | --------------------------------------------- | ------------- | --------------- | ------------------------------------------------- |
| `POST`   | `/v1/master-purchases`                        | ✅ Bearer     | ✅ Required     | Create master purchase; returns projected schedule |
| `GET`    | `/v1/master-purchases`                        | ✅ Bearer     | —               | List all master purchases for the tenant           |
| `GET`    | `/v1/master-purchases/{id}`                   | ✅ Bearer     | —               | Get master purchase by ID                          |
| `GET`    | `/v1/accounts/{account_id}/master-purchases`  | ✅ Bearer     | —               | List master purchases for a specific account       |
| `PATCH`  | `/v1/master-purchases/{id}`                   | ✅ Bearer     | —               | Update description/category                        |
| `DELETE` | `/v1/master-purchases/{id}`                   | ✅ Bearer     | —               | Soft delete (forbidden if status=closed)           |

### Request / Response shapes

```go
// POST /v1/master-purchases — request
type CreateMasterPurchaseRequest struct {
    AccountID            string `json:"account_id"             validate:"required"`
    CategoryID           string `json:"category_id"            validate:"required"`
    Description          string `json:"description"            validate:"required,min=1,max=255"`
    TotalAmountCents     int64  `json:"total_amount_cents"     validate:"required,gt=0"`
    InstallmentCount     int32  `json:"installment_count"      validate:"required,min=2,max=48"`
    ClosingDay           int32  `json:"closing_day"            validate:"required,min=1,max=28"`
    FirstInstallmentDate string `json:"first_installment_date" validate:"required,datetime=2006-01-02"`
}

// POST /v1/master-purchases — response (201 Created)
type CreateMasterPurchaseResponse struct {
    MasterPurchase      domain.MasterPurchase        `json:"master_purchase"`
    ProjectedSchedule   []domain.ProjectedInstallment `json:"projected_schedule"`
}
```

### Error cases to handle

| Scenario                             | Sentinel Error           | HTTP Status |
| ------------------------------------ | ------------------------ | ----------- |
| Account not found / wrong tenant     | `domain.ErrNotFound`     | `404`       |
| Account is not credit_card type      | `domain.ErrInvalidInput` | `422`       |
| Master purchase not found            | `domain.ErrNotFound`     | `404`       |
| Delete on closed purchase            | `domain.ErrForbidden`    | `403`       |
| Invalid request body / validation    | validation error         | `422`       |

---

## 5. Acceptance Criteria

- [ ] All exported functions have Go doc comments.
- [ ] `POST` response includes both `master_purchase` and `projected_schedule` fields.
- [ ] `tenantID` and `userID` are always extracted from context (never from request body).
- [ ] Idempotency middleware is applied to `POST /v1/master-purchases`.
- [ ] All domain errors are mapped to correct HTTP status codes.
- [ ] Swaggo annotations present on all 6 handlers; `swag init` succeeds.
- [ ] Routes registered using Go 1.22 `METHOD /path/{param}` syntax.
- [ ] Unit tests cover happy path and all error paths for each handler.
- [ ] Test coverage for new code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.6 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status       |
| ---------------------------------------------- | -------- | ------------ |
| Task 2.5 — `MasterPurchaseService`             | Upstream | 🔵 backlog   |
| Phase 1 handler patterns & middleware          | Upstream | ✅ done      |
| `platform/middleware/auth.go`                  | Upstream | ✅ done      |
| `platform/middleware/idempotency.go`           | Upstream | ✅ done      |
| Task 2.8 — close-invoice endpoint (same file?) | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/master_purchase_handler_test.go`
- **Cases:**
  - `Create` 201: valid body → response contains `master_purchase` + `projected_schedule`.
  - `Create` 422: missing required fields.
  - `Create` 422: account type not credit_card.
  - `Create` 404: account not found.
  - `GetByID` 200 / 404.
  - `ListByTenant` 200 with empty and non-empty lists.
  - `ListByAccount` 200 filtered correctly.
  - `Update` 200 / 404 / 422.
  - `Delete` 204 / 404 / 403 (closed purchase).

### Integration tests (`//go:build integration`)

N/A — handler layer unit tests cover behaviour. End-to-end flow is covered in Task 2.12.

---

## 8. Open Questions

| # | Question                                                                           | Owner | Resolution |
| - | ---------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `GET /v1/master-purchases` support pagination/filtering?                    | —     | Yes — add `limit`/`offset` query params following the transaction list pattern. |
| 2 | Should the `projected_schedule` also appear in the `GET` responses?                | —     | Yes — include it in `GetByID`; omit from list responses for performance. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
