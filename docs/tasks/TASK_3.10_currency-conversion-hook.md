# Task 3.10 — Currency Conversion Hook (Extensible; No External API in MVP)

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Infrastructure
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define a `CurrencyConverter` interface in the domain layer and provide two implementations: a no-op passthrough (for MVP, when all positions share the same currency) and a static rate table (loaded from config or a DB table). The `InvestmentService` receives the converter via dependency injection, so a real external-feed implementation can be swapped in later without touching service or handler code.

---

## 2. Context & Motivation

Phase 3 positions can be denominated in different currencies (e.g., USD stocks in a BRL-denominated portfolio). Without a conversion hook, `GetPortfolioSummary` would either:

1. Sum incompatible currency values (incorrect), or
2. Crash/error when positions use different currencies.

The project's financial rules forbid `float64` for monetary values, so the converter must work entirely in `int64` cents using integer exchange rates (e.g., stored as cents-per-unit or as a fixed-point multiplier).

For Phase 3 MVP, only a static rate table is required. The interface design ensures that future phases can inject a live-feed implementation (e.g., from Open Exchange Rates or a Redis-cached feed) without modifying calling code.

---

## 3. Scope

### In scope

- [ ] `domain.CurrencyConverter` interface defined in `internal/domain/currency.go`.
- [ ] `pkg/currency/noop_converter.go` — passthrough that returns the input unchanged (all positions same currency).
- [ ] `pkg/currency/static_converter.go` — reads a hard-coded or config-provided `map[string]map[string]int64` rate table (e.g., `"USD" → "BRL" → 500` meaning 1 USD = R$5.00 = 500 cents).
- [ ] `internal/domain/currency_test.go` — unit tests for the interface and both implementations.
- [ ] `InvestmentService` injected with `CurrencyConverter` (update Task 3.6 constructor).
- [ ] `GetPortfolioSummary` uses the converter to normalise all position values into the tenant's base currency before summing.

### Out of scope

- External API integration (e.g., Open Exchange Rates, ECB feed) — deferred.
- Real-time rate caching in Redis — Phase 5.
- UI for managing exchange rates — out of scope for all MVP phases.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                    | Purpose                                              |
| ------ | --------------------------------------- | ---------------------------------------------------- |
| CREATE | `internal/domain/currency.go`           | `CurrencyConverter` interface + `ConvertParams` type |
| CREATE | `internal/domain/currency_test.go`      | Unit tests                                           |
| CREATE | `pkg/currency/noop_converter.go`        | No-op implementation (same-currency portfolios)      |
| CREATE | `pkg/currency/static_converter.go`      | Static rate table implementation                     |
| CREATE | `pkg/currency/static_converter_test.go` | Unit tests for static converter                      |
| MODIFY | `internal/service/investment_service.go`| Inject `CurrencyConverter`; use in `GetPortfolioSummary` |

### Interface definition

```go
// CurrencyConverter converts a monetary amount from one currency to another.
// All amounts are expressed in cents (int64) to avoid float imprecision.
type CurrencyConverter interface {
    // Convert returns amountCents in the target currency.
    // Returns an error if the conversion rate is not available.
    Convert(ctx context.Context, amountCents int64, from, to string) (int64, error)
}
```

### Static converter

```go
// StaticConverter holds a hard-coded rate table.
// Rates are expressed as integer multipliers: rate["USD"]["BRL"] = 500
// means 1 USD = 500 cents BRL (i.e., R$5.00).
type StaticConverter struct {
    rates map[string]map[string]int64
}

func NewStaticConverter(rates map[string]map[string]int64) *StaticConverter {
    return &StaticConverter{rates: rates}
}

func (c *StaticConverter) Convert(_ context.Context, amountCents int64, from, to string) (int64, error) {
    if from == to {
        return amountCents, nil
    }
    inner, ok := c.rates[from]
    if !ok {
        return 0, fmt.Errorf("no exchange rate from %s: %w", from, domain.ErrNotFound)
    }
    rate, ok := inner[to]
    if !ok {
        return 0, fmt.Errorf("no exchange rate from %s to %s: %w", from, to, domain.ErrNotFound)
    }
    return amountCents * rate / 100, nil // rate is in cents-per-100-cents of source
}
```

### Noop converter

```go
// NoopConverter returns the input amount unchanged.
// Used when all positions are already in the tenant's base currency.
type NoopConverter struct{}

func (NoopConverter) Convert(_ context.Context, amountCents int64, _, _ string) (int64, error) {
    return amountCents, nil
}
```

### Updated `GetPortfolioSummary` (delta)

```go
for _, p := range positions {
    converted, err := s.converter.Convert(ctx, p.LastPriceCents, p.Currency, baseCurrency)
    if err != nil {
        // log and skip position, do not abort entire summary
        s.logger.WarnContext(ctx, "investment: currency conversion failed",
            "position_id", p.ID, "from", p.Currency, "to", baseCurrency, "error", err)
        continue
    }
    // ... sum using `converted`
}
```

---

## 5. Acceptance Criteria

- [ ] `domain.CurrencyConverter` interface is in `internal/domain/currency.go`.
- [ ] `NoopConverter.Convert` returns the input unchanged for any currency pair.
- [ ] `StaticConverter.Convert` returns the correct converted amount for a known rate.
- [ ] `StaticConverter.Convert` returns a wrapped `domain.ErrNotFound` for an unknown pair.
- [ ] Same-currency conversion (`from == to`) always returns the input unchanged (even in `StaticConverter`).
- [ ] `InvestmentService.GetPortfolioSummary` uses the converter and logs + skips positions where conversion fails.
- [ ] Unit tests cover all three converters and the updated `GetPortfolioSummary` branch.
- [ ] No `float64` or `float32` used in any conversion arithmetic.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.10 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                                        | Type     | Status     |
| ----------------------------------------------------------------- | -------- | ---------- |
| Task 3.3 — `domain.ErrNotFound` and `InvestmentService` interface | Upstream | 🔵 backlog |
| Task 3.6 — `InvestmentService` constructor (to inject converter)  | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/domain/currency_test.go`
  - Same-currency conversion is a no-op.
- **File:** `pkg/currency/static_converter_test.go`
  - Known rate → correct integer result.
  - Unknown source currency → `ErrNotFound`.
  - Unknown target currency → `ErrNotFound`.
  - `from == to` → passthrough.
- **File:** `internal/service/investment_service_test.go` (updated)
  - `GetPortfolioSummary` with mixed-currency positions uses the converter mock.
  - Converter returns error for one position → that position is skipped, others counted.

### Integration tests

N/A — conversion is pure in-memory logic. Integration is covered by the service integration test when combined with real positions.

---

## 8. Open Questions

| # | Question                                                                                         | Owner | Resolution |
| - | ------------------------------------------------------------------------------------------------ | ----- | ---------- |
| 1 | How should the rate be represented to avoid integer overflow for large amounts?                  | —     | Use `int64` with cents-per-100-source-cents; document maximum safe amount in comments |
| 2 | Should rates be loaded from a DB table (`exchange_rates`) or only from config?                   | —     | Config/env for MVP (no new DB table); add DB table when live-feed is needed |
| 3 | Should the base currency be configurable per tenant (from the `tenants` table) or global?        | —     | Per-tenant; read from `tenant.Currency` field (may require adding that column in a separate task) |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
