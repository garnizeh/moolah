# Task 3.9 — `GET /v1/investments/summary`: Net Worth + Allocation Breakdown

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › HTTP Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement the `GET /v1/investments/summary` endpoint, which returns a tenant's current net worth across all investment positions, broken down by asset type (stocks, bonds, funds, crypto, real estate). This is a read-only, computed endpoint that aggregates data from the `positions` table in real time — no snapshot required.

This task is scoped narrowly: the endpoint wire-up, handler function, and handler unit test. The business logic lives entirely in `InvestmentService.GetPortfolioSummary` (Task 3.6).

---

## 2. Context & Motivation

The portfolio summary is the primary read-facing feature of Phase 3. Users want a single API call that answers: "What is my total investment value right now, and how is it distributed across asset classes?" This is the most frequently called endpoint in the investment domain — it must be fast (in-memory aggregation over DB list query), and it must never return monetary values as floats.

This task may be largely complete as a by-product of Task 3.7 (which registers `GET /v1/investments/summary` as part of the handler batch). If that route is already wired in Task 3.7, this task's deliverable is the explicit unit test coverage and Swaggo annotation verification for that specific endpoint.

---

## 3. Scope

### In scope

- [ ] `GET /v1/investments/summary` is registered in `internal/server/routes.go`.
- [ ] `InvestmentHandler.GetSummary` function exists with full Swaggo annotation.
- [ ] Response shape: `PortfolioSummary` (total_value_cents, currency, allocation_by_type map, positions array).
- [ ] Unit test: `TestInvestmentHandler_GetSummary` in `internal/handler/investment_handler_test.go`.
- [ ] Integration test assertions in the Phase 3 smoke test (future task).

### Out of scope

- Historical trend endpoint (requires portfolio_snapshots; separate task if added).
- Currency conversion (Task 3.10 provides the hook).
- Caching / Redis-backed response cache (Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                          | Purpose                                               |
| ------ | --------------------------------------------- | ----------------------------------------------------- |
| MODIFY | `internal/handler/investment_handler.go`      | Add `GetSummary` handler if not already present       |
| MODIFY | `internal/handler/investment_handler_test.go` | Add `TestInvestmentHandler_GetSummary` unit test      |
| MODIFY | `internal/server/routes.go`                   | Verify route is registered (may already be from 3.7)  |

### Handler function

```go
// GetSummary handles GET /v1/investments/summary
//
// @Summary     Investment portfolio summary
// @Description Returns the current net worth and asset-type allocation for the authenticated tenant.
// @Tags        investments
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} domain.PortfolioSummary
// @Failure     401 {object} map[string]string "Unauthorized"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /v1/investments/summary [get]
func (h *InvestmentHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
    tenantID, ok := middleware.TenantIDFromCtx(r.Context())
    if !ok {
        respondError(w, r, "unauthorized", http.StatusUnauthorized)
        return
    }

    summary, err := h.service.GetPortfolioSummary(r.Context(), tenantID)
    if err != nil {
        handleError(w, r, err, "failed to get portfolio summary")
        return
    }

    respondJSON(w, r, summary, http.StatusOK)
}
```

### Response shape

```json
{
  "total_value_cents": 1500000,
  "currency": "BRL",
  "allocation_by_type": {
    "stock":  900000,
    "fund":   450000,
    "crypto": 150000
  },
  "positions": [
    {
      "id": "01HZ...",
      "asset_id": "01HY...",
      "account_id": "01HX...",
      "quantity": "10.5",
      "avg_cost_cents": 130000,
      "last_price_cents": 143000,
      "currency": "BRL",
      "purchased_at": "2025-06-01T00:00:00Z"
    }
  ]
}
```

### Error cases

| Scenario                            | HTTP Status |
| ----------------------------------- | ----------- |
| Missing or invalid auth token       | `401`       |
| Service internal error              | `500`       |
| Tenant has no positions             | `200` with zero values (not an error) |

---

## 5. Acceptance Criteria

- [ ] `GET /v1/investments/summary` is registered and returns `200` for an authenticated tenant.
- [ ] Response includes `total_value_cents` (int64), `currency` (string), `allocation_by_type` (map), and `positions` (array).
- [ ] `total_value_cents` is never a float — always an integer.
- [ ] Empty positions returns `200` with `total_value_cents: 0` and empty maps/arrays.
- [ ] Swaggo annotation is complete and `api/swagger.json` is regenerated.
- [ ] `TestInvestmentHandler_GetSummary` passes with mocked service.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.9 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                                  | Type     | Status     |
| ----------------------------------------------------------- | -------- | ---------- |
| Task 3.6 — `InvestmentService.GetPortfolioSummary` exists   | Upstream | 🔵 backlog |
| Task 3.7 — `InvestmentHandler` struct exists                | Upstream | 🔵 backlog |
| `domain.PortfolioSummary` type defined (Task 3.3)           | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/investment_handler_test.go`
- **Cases:**
  - Happy path: `GetSummary` returns `200` with correct `PortfolioSummary` body.
  - No auth: expects `401`.
  - Service returns error: expects `500`.
  - Empty portfolio: expects `200` with `total_value_cents: 0`.

### Integration tests

Covered by the Phase 3 smoke test (future task that mirrors `TestSmoke_Phase2HappyPath`).

---

## 8. Open Questions

| # | Question                                                                     | Owner | Resolution |
| - | ---------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `allocation_by_type` include percentage values alongside absolute cents? | —  | Cents only for MVP; percentages can be computed client-side |
| 2 | Should this endpoint be paginated (if a tenant has hundreds of positions)?  | —     | Not for MVP; add pagination in a future task if load requires it |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
