# Task 3.6 — Service: `InvestmentService` — Position CRUD, Allocation & Receivable Lifecycle

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Service Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-14
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `InvestmentService` — the orchestration layer that fulfils the `domain.InvestmentService` interface. Covers position CRUD with validation, real-time portfolio allocation calculation, snapshot assembly, and the receivable lifecycle (marking `position_income_events` as `received` or `cancelled`). Full unit test coverage with mocked repositories is required.

---

## 2. Context & Motivation

Services in this project orchestrate multiple repositories and encode business rules. Handlers call the service; repositories know nothing about business logic. The `InvestmentService` is the largest service in Phase 3 because positions span capital, income scheduling, and receivable tracking.

Key business rules from ADR-003:

- `CreatePosition`: account must be of type `investment`; a job position has `quantity=1`, `avg_cost_cents=0`.
- `GetPortfolioSummary`: aggregates `quantity × last_price_cents` per position; groups allocation by `AssetType`; uses `CurrencyConverter` to normalise to tenant base currency.
- `TakeSnapshot`: assembles the summary and writes a `portfolio_snapshots` row (idempotent via unique constraint).
- `MarkIncomeReceived`: transitions `pending → received`; optionally creates a credit `transactions` row in the linked cash account.
- `CancelIncome`: transitions `pending → cancelled`.

**Reference:** ADR-003 §2.2, §2.5, §9, §10.

---

## 3. Scope

### In scope

- [x] `internal/service/investment_service.go` — `InvestmentService` implementing `domain.InvestmentService`.
- [x] `internal/service/investment_service_test.go` — unit tests with mocked repositories.
- [x] Position CRUD methods: `CreatePosition`, `GetPosition`, `ListPositions`, `UpdatePosition`, `DeletePosition`.
- [x] Receivable lifecycle: `MarkIncomeReceived(ctx, tenantID, eventID string)`, `CancelIncome(ctx, tenantID, eventID string)`.
- [x] Portfolio methods: `GetPortfolioSummary`, `TakeSnapshot`.
- [x] Constructor receives: `positionRepo`, `positionSnapshotRepo`, `positionIncomeRepo`, `portfolioSnapshotRepo`, `assetRepo`, `tenantAssetConfigRepo`, `accountRepo`, `auditRepo`, `currencyConverter`.

### Out of scope

- Income scheduler goroutine (Task 3.13 — separate background job).
- HTTP handler (Task 3.14).
- `CurrencyConverter` implementation (Task 3.10 — injected via interface).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                       |
| ------ | ---------------------------------------------- | --------------------------------------------- |
| CREATE | `internal/service/investment_service.go`       | `InvestmentService` concrete implementation   |
| CREATE | `internal/service/investment_service_test.go`  | Unit tests with mocked repos                  |

### Constructor

```go
type InvestmentService struct {
    positionRepo          domain.PositionRepository
    positionSnapshotRepo  domain.PositionSnapshotRepository
    positionIncomeRepo    domain.PositionIncomeEventRepository
    portfolioSnapshotRepo domain.PortfolioSnapshotRepository
    assetRepo             domain.AssetRepository
    tenantConfigRepo      domain.TenantAssetConfigRepository
    accountRepo           domain.AccountRepository
    auditRepo             domain.AuditRepository
    currencyConverter     domain.CurrencyConverter
}

func NewInvestmentService(
    positionRepo          domain.PositionRepository,
    positionSnapshotRepo  domain.PositionSnapshotRepository,
    positionIncomeRepo    domain.PositionIncomeEventRepository,
    portfolioSnapshotRepo domain.PortfolioSnapshotRepository,
    assetRepo             domain.AssetRepository,
    tenantConfigRepo      domain.TenantAssetConfigRepository,
    accountRepo           domain.AccountRepository,
    auditRepo             domain.AuditRepository,
    currencyConverter     domain.CurrencyConverter,
) *InvestmentService { ... }
```

### Key method signatures

```go
// Position CRUD
func (s *InvestmentService) CreatePosition(ctx context.Context, tenantID string, in domain.CreatePositionInput) (*domain.Position, error)
func (s *InvestmentService) ListPositions(ctx context.Context, tenantID string) ([]domain.Position, error)
func (s *InvestmentService) UpdatePosition(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error)
func (s *InvestmentService) DeletePosition(ctx context.Context, tenantID, id string) error

// Receivable lifecycle
func (s *InvestmentService) MarkIncomeReceived(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error)
func (s *InvestmentService) CancelIncome(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error)

// Portfolio
func (s *InvestmentService) GetPortfolioSummary(ctx context.Context, tenantID string) (*domain.PortfolioSummary, error)
func (s *InvestmentService) TakeSnapshot(ctx context.Context, tenantID string) (*domain.PortfolioSnapshot, error)
```

### API endpoints (if applicable)

N/A — service layer; no HTTP endpoints.

### Error cases to handle

| Scenario                        | Sentinel Error          | HTTP Status (handler converts) |
| ------------------------------- | ----------------------- | ------------------------------ |
| Account is not `investment` type| `domain.ErrInvalidInput`| `422`                          |
| Position not found              | `domain.ErrNotFound`    | `404`                          |
| Income event already received   | `domain.ErrConflict`    | `409`                          |
| Income event already cancelled  | `domain.ErrConflict`    | `409`                          |

---

## 5. Acceptance Criteria

- [x] `InvestmentService` implements the `domain.InvestmentService` interface.
- [x] `CreatePosition` rejects accounts whose `type != "investment"`.
- [x] `MarkIncomeReceived` returns `domain.ErrConflict` if `status != "pending"`.
- [x] `CancelIncome` returns `domain.ErrConflict` if `status != "pending"`.
- [x] `GetPortfolioSummary` uses `CurrencyConverter` to normalise to tenant base currency before summing.
- [x] `TakeSnapshot` is idempotent (unique constraint on `(tenant_id, snapshot_date)` — log and return existing on conflict).
- [x] All service methods pass `tenant_id` to repositories; cross-tenant access is impossible.
- [x] Unit tests cover each business rule with mocked repos.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.6 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                              |
| ---------- | ------ | --------------------------------------------------- |
| 2026-03-14 | Copilot| Task completed; all business rules implemented and verified with tests. |
| 2026-03-13 | —      | Task created; updated for ADR v3 (receivable lifecycle, income scheduling, COALESCE assets, full constructor) |
