# Task 3.7 — `handler/investment_handler.go`: Positions, Allocation, History

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › HTTP Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the `InvestmentHandler` struct and register all Phase 3 HTTP endpoints in `internal/server/routes.go`. The handler translates HTTP requests into `domain.InvestmentService` calls and formats responses as JSON. Full unit test coverage using a mocked `InvestmentService` is required. Swaggo annotations must be added to every endpoint.

---

## 2. Context & Motivation

Handlers in this project are thin: they decode and validate requests, call the service, then encod responses. All business logic lives in the service. Phase 3 endpoints expose investment portfolio data — positions, asset catalogue, and summary — to authenticated tenant users. The `POST /v1/accounts/{id}/snapshot` endpoint triggers a manual portfolio snapshot (auto-scheduling is Phase 5).

---

## 3. Scope

### In scope

- [ ] `internal/handler/investment_handler.go` — `InvestmentHandler` struct + all handlers.
- [ ] `internal/handler/investment_handler_test.go` — unit tests with mocked service.
- [ ] Route registration for all Phase 3 endpoints in `internal/server/routes.go`.
- [ ] DI wiring in `internal/server/server.go` (add `investmentHandler` field).
- [ ] Swaggo annotations on every handler function.
- [ ] Idempotency middleware applied to `POST` and `PATCH` routes.

### Out of scope

- Asset admin endpoints (creating/managing the global asset catalogue) — considered a sysadmin concern; can be added in a future task.
- Price feed WebSocket push (deferred).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                          |
| ------ | ---------------------------------------------- | ------------------------------------------------ |
| CREATE | `internal/handler/investment_handler.go`       | HTTP handler implementation                      |
| CREATE | `internal/handler/investment_handler_test.go`  | Unit tests with mocked service                   |
| MODIFY | `internal/server/routes.go`                    | Register Phase 3 routes                          |
| MODIFY | `internal/server/server.go`                    | Add `investmentHandler` field and DI constructor |

### API Endpoints

| Method | Path                                       | Auth     | Idempotency | Description                              |
| ------ | ------------------------------------------ | -------- | ----------- | ---------------------------------------- |
| GET    | `/v1/assets`                               | ✅ Bearer | ❌         | List global asset catalogue              |
| GET    | `/v1/assets/{id}`                          | ✅ Bearer | ❌         | Get asset by ID                          |
| GET    | `/v1/positions`                            | ✅ Bearer | ❌         | List tenant positions                    |
| POST   | `/v1/positions`                            | ✅ Bearer | ✅         | Create a new position                    |
| GET    | `/v1/positions/{id}`                       | ✅ Bearer | ❌         | Get position by ID                       |
| PATCH  | `/v1/positions/{id}`                       | ✅ Bearer | ✅         | Update position (qty, price)             |
| DELETE | `/v1/positions/{id}`                       | ✅ Bearer | ❌         | Soft-delete position                     |
| GET    | `/v1/accounts/{id}/positions`              | ✅ Bearer | ❌         | List positions scoped to account         |
| POST   | `/v1/accounts/{id}/snapshot`               | ✅ Bearer | ✅         | Manually trigger portfolio snapshot      |
| GET    | `/v1/investments/summary`                  | ✅ Bearer | ❌         | Net worth + allocation breakdown         |

### Request / response types (excerpt)

```go
type CreatePositionRequest struct {
    AssetID        string    `json:"asset_id"         validate:"required"`
    AccountID      string    `json:"account_id"       validate:"required"`
    Quantity       string    `json:"quantity"         validate:"required"`
    AvgCostCents   int64     `json:"avg_cost_cents"   validate:"required,gt=0"`
    LastPriceCents int64     `json:"last_price_cents" validate:"omitempty,gte=0"`
    Currency       string    `json:"currency"         validate:"required,len=3"`
    PurchasedAt    time.Time `json:"purchased_at"     validate:"required"`
}

type UpdatePositionRequest struct {
    Quantity       *string `json:"quantity"         validate:"omitempty"`
    AvgCostCents   *int64  `json:"avg_cost_cents"   validate:"omitempty,gt=0"`
    LastPriceCents *int64  `json:"last_price_cents" validate:"omitempty,gte=0"`
}
```

### Error cases to handle

| Scenario                           | Sentinel Error           | HTTP Status |
| ---------------------------------- | ------------------------ | ----------- |
| Position not found                 | `domain.ErrNotFound`     | `404`       |
| Asset not found                    | `domain.ErrNotFound`     | `404`       |
| Account type not `investment`      | `domain.ErrInvalidInput` | `422`       |
| Missing idempotency key            | middleware               | `400`       |
| Validation failure                 | validator error          | `422`       |

---

## 5. Acceptance Criteria

- [ ] All 10 endpoints are registered in `routes.go` with correct HTTP methods.
- [ ] `POST` and `PATCH` routes are wrapped with the idempotency middleware.
- [ ] Every handler decodes, validates, and delegates to `InvestmentService`.
- [ ] Swaggo annotations present on every handler (summary, tags, params, responses).
- [ ] `internal/handler/investment_handler_test.go` covers all happy paths and documented error cases.
- [ ] Unit tests use the `domain.InvestmentService` mock — zero real DB calls.
- [ ] `t.Parallel()` set in all test functions and subtests.
- [ ] Test coverage for new handler code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.7 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                             | Type     | Status     |
| ------------------------------------------------------ | -------- | ---------- |
| Task 3.3 — `domain.InvestmentService` interface        | Upstream | 🔵 backlog |
| Task 3.6 — Service implementation for DI wiring        | Upstream | 🔵 backlog |
| `internal/server/routes.go` Phase 2 routes registered  | Upstream | ✅ done    |
| Idempotency middleware                                  | Upstream | ✅ done    |
| `go-playground/validator` already in vendor            | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/investment_handler_test.go`
- **Cases:**
  - `POST /v1/positions` — 201 on valid input.
  - `POST /v1/positions` — 422 on invalid account type (service returns `ErrInvalidInput`).
  - `GET /v1/positions/{id}` — 200 on found; 404 on missing.
  - `PATCH /v1/positions/{id}` — 200 on valid patch.
  - `DELETE /v1/positions/{id}` — 204 on success.
  - `GET /v1/investments/summary` — 200 with correct structure.
  - `POST /v1/accounts/{id}/snapshot` — 200 on success.

### Integration tests

Covered by future Phase 3 smoke test (pattern mirrors `TestSmoke_Phase1HappyPath`).

---

## 8. Open Questions

| # | Question                                                                         | Owner | Resolution |
| - | -------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `GET /v1/assets` be public (no auth) to allow pre-auth asset lookup?     | —     | Require auth for MVP; revisit if public catalogue is needed |
| 2 | Should asset creation be sysadmin-only or open to any authenticated user?       | —     | Sysadmin-only for the global shared catalogue |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
