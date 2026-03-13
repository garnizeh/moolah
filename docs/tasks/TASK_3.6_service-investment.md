# Task 3.6 — `service/investment_service.go` + Unit Tests

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Service Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `InvestmentService` — the orchestration layer that fulfils the `domain.InvestmentService` interface defined in Task 3.3. The service encodes all Phase 3 business rules: position upsert semantics, allocation calculation, and portfolio summary assembly. Full unit test coverage is required using mocked repositories.

---

## 2. Context & Motivation

Services in this project sit between handlers and repositories. They hold business logic that spans multiple repositories or requires computed results — such as summing position values for a net-worth summary, or calculating allocation percentages per asset type. Repositories do no computation; handlers do no business logic.

The `InvestmentService` must be fully mockable (it implements `domain.InvestmentService`), so handlers can be unit-tested without hitting the database.

---

## 3. Scope

### In scope

- [ ] `internal/service/investment_service.go` — implements `domain.InvestmentService`.
- [ ] Business rules:
  - `CreatePosition`: validates account type is `investment`; rejects other account types.
  - `GetPortfolioSummary`: sums `quantity × last_price_cents` per position; groups allocation by `AssetType`.
  - `TakeSnapshot`: assembles summary and persists via `PortfolioSnapshotRepository`.
- [ ] `internal/service/investment_service_test.go` — unit tests with mocked repositories.

### Out of scope

- HTTP handler (Task 3.7).
- Scheduler/cron integration (Task 3.8).
- Currency conversion (Task 3.10 provides the hook the service will call).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                          |
| ------ | ---------------------------------------------- | ------------------------------------------------ |
| CREATE | `internal/service/investment_service.go`       | `InvestmentService` concrete implementation      |
| CREATE | `internal/service/investment_service_test.go`  | Unit tests with mocked repos                     |

### Constructor & struct

```go
type InvestmentService struct {
    assetRepo    domain.AssetRepository
    positionRepo domain.PositionRepository
    snapshotRepo domain.PortfolioSnapshotRepository
    accountRepo  domain.AccountRepository // to validate account type
    auditRepo    domain.AuditRepository
}

func NewInvestmentService(
    assetRepo    domain.AssetRepository,
    positionRepo domain.PositionRepository,
    snapshotRepo domain.PortfolioSnapshotRepository,
    accountRepo  domain.AccountRepository,
    auditRepo    domain.AuditRepository,
) *InvestmentService {
    return &InvestmentService{
        assetRepo:    assetRepo,
        positionRepo: positionRepo,
        snapshotRepo: snapshotRepo,
        accountRepo:  accountRepo,
        auditRepo:    auditRepo,
    }
}
```

### Key business rule — portfolio summary

```go
func (s *InvestmentService) GetPortfolioSummary(ctx context.Context, tenantID string) (*domain.PortfolioSummary, error) {
    positions, err := s.positionRepo.ListByTenant(ctx, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to list positions: %w", err)
    }

    var totalValueCents int64
    allocationByType := make(map[string]int64)

    for _, p := range positions {
        qty := parseQuantity(p.Quantity) // convert string to rational; no float
        posValue := int64(qty * float64(p.LastPriceCents)) // use integer math
        totalValueCents += posValue

        asset, err := s.assetRepo.GetByID(ctx, p.AssetID)
        if err == nil {
            allocationByType[string(asset.AssetType)] += posValue
        }
    }

    return &domain.PortfolioSummary{
        TotalValueCents:  totalValueCents,
        Currency:         "BRL", // TODO: resolved via Task 3.10 currency hook
        AllocationByType: allocationByType,
        Positions:        positions,
    }, nil
}
```

> **Note:** Quantity multiplication must never use `float64` for the final `int64` result. Task 3.10 provides the currency conversion hook.

### Error cases

| Scenario                                    | Returns                                       |
| ------------------------------------------- | --------------------------------------------- |
| Account not found                           | `domain.ErrNotFound`                          |
| Account type is not `investment`            | `domain.ErrInvalidInput` ("account type must be investment") |
| Position not found                          | `domain.ErrNotFound`                          |
| Duplicate snapshot for same month           | Wrapped DB constraint error (unique violation)|

---

## 5. Acceptance Criteria

- [ ] `InvestmentService` compiles and satisfies `domain.InvestmentService` interface.
- [ ] `CreatePosition` returns error if `accountType != "investment"`.
- [ ] `GetPortfolioSummary` returns `total_value_cents` as the integer sum of all position values.
- [ ] `TakeSnapshot` persists a `PortfolioSnapshot` and returns it.
- [ ] Unit tests cover all happy paths and the key error paths above.
- [ ] Unit tests use gomock/moq mocks — no real DB calls.
- [ ] `t.Parallel()` is set in all test functions and subtests.
- [ ] Test coverage for new service code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.6 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                         | Type     | Status     |
| -------------------------------------------------- | -------- | ---------- |
| Task 3.3 — Domain interfaces + `InvestmentService` | Upstream | 🔵 backlog |
| Task 3.5 — Repository implementations (for DI)    | Upstream | 🔵 backlog |
| `internal/testutil/mocks` — repo mocks available   | Upstream | ✅ done    |
| `domain.ErrInvalidInput` defined (or add in this task) | Upstream | TBD   |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/service/investment_service_test.go`
- **Cases:**
  - `CreatePosition`: happy path — valid investment account.
  - `CreatePosition`: error — account not found → `ErrNotFound`.
  - `CreatePosition`: error — wrong account type → `ErrInvalidInput`.
  - `GetPortfolioSummary`: returns correct total and allocation map.
  - `GetPortfolioSummary`: empty positions → zero value summary (no error).
  - `TakeSnapshot`: delegates to snapshot repo; returns persisted snapshot.
  - `DeletePosition`: delegates to repo; returns no error.

### Integration tests

N/A — service unit tests use mocks. End-to-end coverage deferred to smoke test (Task 3.x or equivalent).

---

## 8. Open Questions

| # | Question                                                                                             | Owner | Resolution |
| - | ---------------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `GetPortfolioSummary` resolve the tenant's base currency from the `Tenant` record?           | —     | TBD — design in Task 3.10 |
| 2 | Should `TakeSnapshot` be idempotent (upsert) or fail if a snapshot already exists for the month?    | —     | Fail with clear error for MVP; upsert can be added later |
| 3 | Should `CreatePosition` also create an audit log entry?                                              | —     | Yes — follow Phase 1/2 pattern using `auditRepo` |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
