# Task 3.14 — HTTP Handlers: Positions, Income Events & Portfolio Summary

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › HTTP Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-14
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Implement the `PositionHandler` struct and register all position-related HTTP endpoints: position CRUD, per-account position listing, income event listing and receivable lifecycle (mark received / cancel), manual snapshot trigger, and the net-worth summary. Full unit test coverage and Swaggo annotations required.

---

## 2. Context & Motivation

Task 3.7 covered asset catalogue endpoints. This task covers the tenant's day-to-day investment operations: opening and closing positions, tracking income receivables, triggering snapshots, and reading the portfolio summary. These are the most frequently called investment endpoints.

**Reference:** ADR-003 §2.2, §2.5, §10; handler pattern in `internal/handler/account_handler.go`.

---

## 3. Scope

### In scope

- [ ] `internal/handler/position_handler.go` — `PositionHandler` struct + all handlers.
- [ ] `internal/handler/position_handler_test.go` — unit tests with mocked `InvestmentService`.
- [ ] Route registration in `internal/server/routes.go`.
- [ ] DI wiring in `internal/server/server.go` (add `positionHandler`).
- [ ] Idempotency middleware on `POST` and `PATCH` routes.
- [ ] Swaggo annotations on every handler function.

### Out of scope

- Asset catalogue endpoints (Task 3.7).
- Tenant asset config endpoints (Task 3.7).
- Background job wiring (Tasks 3.8 / 3.13).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                          |
| ------ | ---------------------------------------------- | ------------------------------------------------ |
| CREATE | `internal/handler/position_handler.go`         | HTTP handler implementation                      |
| CREATE | `internal/handler/position_handler_test.go`    | Unit tests with mocked service                   |
| MODIFY | `internal/server/routes.go`                    | Register all position + income + summary routes  |
| MODIFY | `internal/server/server.go`                    | Add `positionHandler` field                      |

### API Endpoints

| Method | Path                                          | Auth      | Idempotency | Description                                  |
| ------ | --------------------------------------------- | --------- | ----------- | -------------------------------------------- |
| GET    | `/v1/positions`                               | ✅ Bearer  | ❌          | List all tenant positions                    |
| POST   | `/v1/positions`                               | ✅ Bearer  | ✅          | Create a new position                        |
| GET    | `/v1/positions/{id}`                          | ✅ Bearer  | ❌          | Get single position                          |
| PATCH  | `/v1/positions/{id}`                          | ✅ Bearer  | ✅          | Update position (qty, price, income schedule)|
| DELETE | `/v1/positions/{id}`                          | ✅ Bearer  | ❌          | Soft-delete position (close/resign)          |
| GET    | `/v1/accounts/{id}/positions`                 | ✅ Bearer  | ❌          | List positions for a specific account        |
| GET    | `/v1/income-events`                           | ✅ Bearer  | ❌          | List income events (all statuses)            |
| GET    | `/v1/income-events/pending`                   | ✅ Bearer  | ❌          | List pending receivables only                |
| PATCH  | `/v1/income-events/{id}/receive`              | ✅ Bearer  | ✅          | Mark income event as received                |
| PATCH  | `/v1/income-events/{id}/cancel`               | ✅ Bearer  | ❌          | Cancel (write off) an income event           |
| POST   | `/v1/portfolio/snapshot`                      | ✅ Bearer  | ✅          | Manually trigger portfolio snapshot          |
| GET    | `/v1/investments/summary`                     | ✅ Bearer  | ❌          | Net worth + allocation breakdown             |

### Request / response types (excerpt)

```go
type CreatePositionRequest struct {
    AssetID             string     `json:"asset_id"              validate:"required"`
    AccountID           string     `json:"account_id"            validate:"required"`
    Quantity            string     `json:"quantity"              validate:"required"`
    AvgCostCents        int64      `json:"avg_cost_cents"        validate:"gte=0"`
    LastPriceCents      int64      `json:"last_price_cents"      validate:"gte=0"`
    Currency            string     `json:"currency"              validate:"required,len=3"`
    PurchasedAt         time.Time  `json:"purchased_at"          validate:"required"`
    IncomeType          string     `json:"income_type"           validate:"required,oneof=none dividend coupon rent interest salary"`
    IncomeIntervalDays  *int       `json:"income_interval_days"  validate:"omitempty,gt=0"`
    IncomeAmountCents   *int64     `json:"income_amount_cents"   validate:"omitempty,gte=0"`
    IncomeRateBps       *int       `json:"income_rate_bps"       validate:"omitempty,gte=0"`
    NextIncomeAt        *time.Time `json:"next_income_at"`
    MaturityAt          *time.Time `json:"maturity_at"`
}
```

### Error cases to handle

| Scenario                         | Sentinel Error              | HTTP Status |
| -------------------------------- | --------------------------- | ----------- |
| Position not found               | `domain.ErrNotFound`        | `404`       |
| Account is not investment type   | `domain.ErrInvalidInput`    | `422`       |
| Income event already received    | `domain.ErrConflict`        | `409`       |
| Income event already cancelled   | `domain.ErrConflict`        | `409`       |

---

## 5. Acceptance Criteria

- [x] All 12 endpoints registered and reachable.
- [x] `POST /v1/positions` validates `income_type` and related scheduling fields.
- [x] `PATCH /v1/income-events/{id}/receive` returns `409` if status is not `pending`.
- [x] `GET /v1/investments/summary` returns `total_value_cents` as `int64` (no floats in JSON).
- [x] Idempotency middleware applied to all `POST` and `PATCH` write endpoints.
- [x] Swaggo annotations present on all handlers.
- [x] Unit tests achieve ≥ 80% coverage for the handler package additions.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.14 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-13 | —      | Task created (new) |
| 2026-03-14 | GitHub Copilot | Task completed: Handlers, Routes, and Tests implemented. |
