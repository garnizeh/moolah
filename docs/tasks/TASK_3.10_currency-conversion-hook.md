# Task 3.10 — `CurrencyConverter` Interface + Static Rate Implementation

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Infrastructure
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the `CurrencyConverter` interface in the domain layer and implement two concrete versions: a no-op passthrough (single-currency portfolios) and a static rate table (loaded from config). The `InvestmentService` (Task 3.6) receives the converter via dependency injection — a live-feed implementation can be swapped in a future phase without modifying any calling code.

---

## 2. Context & Motivation

Phase 3 positions can be denominated in different currencies (e.g., USD stock in a BRL-measured portfolio). All monetary arithmetic must remain in `int64` cents — `float64` is forbidden per project rules. The converter uses integer basis-point logic: `rate["USD"]["BRL"] = 50000` means 1 USD = R$5.00 = 50000 cents.

For Phase 3 MVP, only the static table is required. The interface design ensures Phase 4+ can inject a live-feed implementation without touching service or handler code.

**Reference:** ADR-003 §7.

---

## 3. Scope

### In scope

- [ ] `internal/domain/currency.go` — `CurrencyConverter` interface + `ConvertAmountInput` type.
- [ ] `pkg/currency/noop_converter.go` — passthrough that returns input unchanged (assumes all positions share the same currency).
- [ ] `pkg/currency/static_converter.go` — static rate table from `map[string]map[string]int64`; loaded from config ENV VARs or hard-coded defaults.
- [ ] `pkg/currency/static_converter_test.go` — unit tests for static converter.
- [ ] `internal/domain/currency_test.go` — unit tests for interface and noop implementation.

### Out of scope

- External API integration (Open Exchange Rates, ECB feed) — deferred to Phase 4.
- Real-time rate caching in Redis — Phase 5.
- UI for managing exchange rates.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                    | Purpose                                          |
| ------ | --------------------------------------- | ------------------------------------------------ |
| CREATE | `internal/domain/currency.go`           | `CurrencyConverter` interface + types            |
| CREATE | `internal/domain/currency_test.go`      | Unit tests                                       |
| CREATE | `pkg/currency/noop_converter.go`        | No-op implementation                             |
| CREATE | `pkg/currency/static_converter.go`      | Static rate table implementation                 |
| CREATE | `pkg/currency/static_converter_test.go` | Unit tests for static converter                  |

### Interface

```go
// CurrencyConverter converts a monetary amount from one currency to another.
// All amounts are expressed in cents (int64) to avoid float imprecision.
// Implementations must be safe for concurrent use.
type CurrencyConverter interface {
    // Convert returns amountCents in the target currency.
    // Returns an error if the conversion rate is not available.
    Convert(ctx context.Context, amountCents int64, fromCurrency, toCurrency string) (int64, error)
}

var ErrRateNotFound = errors.New("currency conversion rate not found")
```

### Static converter

```go
// StaticConverter holds a hard-coded or config-provided rate table.
// Rates are expressed in cents of the target currency per cent of the source.
// Example: rates["USD"]["BRL"] = 50000 → 1 USD = R$5.00 (= 500 cents × 100).
// Same-currency conversions always return amountCents unchanged.
type StaticConverter struct {
    rates map[string]map[string]int64
}

func NewStaticConverter(rates map[string]map[string]int64) *StaticConverter { ... }
```

### Error cases to handle

| Scenario                        | Error                         |
| ------------------------------- | ----------------------------- |
| Rate for pair not in table      | `domain.ErrRateNotFound`      |
| `fromCurrency == toCurrency`    | Return `amountCents` unchanged |

---

## 5. Acceptance Criteria

- [ ] `CurrencyConverter` interface defined in `internal/domain/`.
- [ ] `NoopConverter` returns `amountCents` unchanged for any currency pair.
- [ ] `StaticConverter` converts correctly using integer arithmetic (no `float64`).
- [ ] `StaticConverter` returns `ErrRateNotFound` for unknown currency pairs.
- [ ] Same-currency conversion (USD → USD) returns exact input.
- [ ] Unit tests cover: same-currency, known-pair, unknown-pair, zero amount.
- [ ] `make task-check` passes.
- [ ] `docs/ROADMAP.md` row 3.10 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                          |
| ---------- | ------ | ----------------------------------------------- |
| 2026-03-13 | —      | Task updated for ADR v3 (integer-cents logic; rate unit clarified) |
