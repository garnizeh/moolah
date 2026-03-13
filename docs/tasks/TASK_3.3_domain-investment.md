# Task 3.3 — `domain/investment.go`: Entities + Repository Interfaces

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Domain Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define all Phase 3 domain entities (`Asset`, `Position`, `PortfolioSnapshot`) and their repository interfaces in `internal/domain/investment.go`. This file is the authoritative contract for the investment subdomain — every other Phase 3 task (repository, service, handler) depends on it. No business logic lives here; only structs, input types, and interface definitions.

---

## 2. Context & Motivation

The project follows interface-driven development: every repository and service is defined as a Go interface in the `internal/domain/` package before any concrete implementation is written. This enables robust mocking in unit tests and makes the architecture explicitly documented in code.

Phase 1 established `AccountRepository`, `CategoryRepository`, etc. Phase 2 added `MasterPurchaseRepository`. Phase 3 must follow the same pattern with `AssetRepository`, `PositionRepository`, and `PortfolioSnapshotRepository`.

---

## 3. Scope

### In scope

- [ ] `Asset` struct with JSON tags.
- [ ] `Position` struct with JSON tags.
- [ ] `PortfolioSnapshot` struct with JSON tags.
- [ ] `AssetType` string enum type with constants.
- [ ] Input types: `CreateAssetInput`, `CreatePositionInput`, `UpdatePositionInput`.
- [ ] `AssetRepository` interface.
- [ ] `PositionRepository` interface.
- [ ] `PortfolioSnapshotRepository` interface.
- [ ] `InvestmentService` interface (used by handler layer).
- [ ] Unit tests for domain validation helpers (if any) in `internal/domain/investment_test.go`.

### Out of scope

- Concrete repository implementations (Task 3.5).
- Service implementation (Task 3.6).
- SQL queries (Task 3.4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                    | Purpose                                  |
| ------ | --------------------------------------- | ---------------------------------------- |
| CREATE | `internal/domain/investment.go`         | Entities, input types, interfaces        |
| CREATE | `internal/domain/investment_test.go`    | Unit tests for domain types/helpers      |

### Key interfaces / types

```go
// AssetType represents the investment instrument category.
type AssetType string

const (
    AssetTypeStock      AssetType = "stock"
    AssetTypeBond       AssetType = "bond"
    AssetTypeFund       AssetType = "fund"
    AssetTypeCrypto     AssetType = "crypto"
    AssetTypeRealEstate AssetType = "real_estate"
)

// Asset is a global reference record for a tradeable instrument.
type Asset struct {
    ID        string    `json:"id"`
    Ticker    string    `json:"ticker"`
    Name      string    `json:"name"`
    AssetType AssetType `json:"asset_type"`
    Currency  string    `json:"currency"`
    CreatedAt time.Time `json:"created_at"`
}

// Position represents a tenant's holding in a specific asset.
type Position struct {
    ID              string    `json:"id"`
    TenantID        string    `json:"-"`
    AssetID         string    `json:"asset_id"`
    AccountID       string    `json:"account_id"`
    Quantity        string    `json:"quantity"` // NUMERIC returned as string to avoid float imprecision
    AvgCostCents    int64     `json:"avg_cost_cents"`
    LastPriceCents  int64     `json:"last_price_cents"`
    Currency        string    `json:"currency"`
    PurchasedAt     time.Time `json:"purchased_at"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    DeletedAt       *time.Time `json:"-"`
}

// PortfolioSnapshot is a monthly point-in-time valuation.
type PortfolioSnapshot struct {
    ID               string          `json:"id"`
    TenantID         string          `json:"-"`
    SnapshotDate     time.Time       `json:"snapshot_date"`
    TotalValueCents  int64           `json:"total_value_cents"`
    Currency         string          `json:"currency"`
    Details          json.RawMessage `json:"details,omitempty"`
    CreatedAt        time.Time       `json:"created_at"`
}

// --- Input types ---

type CreateAssetInput struct {
    Ticker    string    `json:"ticker"     validate:"required,min=1,max=20"`
    Name      string    `json:"name"       validate:"required,min=1,max=200"`
    AssetType AssetType `json:"asset_type" validate:"required,oneof=stock bond fund crypto real_estate"`
    Currency  string    `json:"currency"   validate:"required,len=3"`
}

type CreatePositionInput struct {
    AssetID        string    `json:"asset_id"         validate:"required"`
    AccountID      string    `json:"account_id"       validate:"required"`
    Quantity       string    `json:"quantity"         validate:"required"`
    AvgCostCents   int64     `json:"avg_cost_cents"   validate:"required,gt=0"`
    LastPriceCents int64     `json:"last_price_cents" validate:"omitempty,gte=0"`
    Currency       string    `json:"currency"         validate:"required,len=3"`
    PurchasedAt    time.Time `json:"purchased_at"     validate:"required"`
}

type UpdatePositionInput struct {
    Quantity       *string `json:"quantity"         validate:"omitempty"`
    AvgCostCents   *int64  `json:"avg_cost_cents"   validate:"omitempty,gt=0"`
    LastPriceCents *int64  `json:"last_price_cents" validate:"omitempty,gte=0"`
}

// --- Repository interfaces ---

type AssetRepository interface {
    Create(ctx context.Context, input CreateAssetInput) (*Asset, error)
    GetByID(ctx context.Context, id string) (*Asset, error)
    GetByTicker(ctx context.Context, ticker string) (*Asset, error)
    List(ctx context.Context) ([]Asset, error)
}

type PositionRepository interface {
    Create(ctx context.Context, tenantID string, input CreatePositionInput) (*Position, error)
    GetByID(ctx context.Context, tenantID, id string) (*Position, error)
    ListByTenant(ctx context.Context, tenantID string) ([]Position, error)
    ListByAccount(ctx context.Context, tenantID, accountID string) ([]Position, error)
    Update(ctx context.Context, tenantID, id string, input UpdatePositionInput) (*Position, error)
    Delete(ctx context.Context, tenantID, id string) error
}

type PortfolioSnapshotRepository interface {
    Create(ctx context.Context, tenantID string, snapshot PortfolioSnapshot) (*PortfolioSnapshot, error)
    GetByDate(ctx context.Context, tenantID string, date time.Time) (*PortfolioSnapshot, error)
    ListByTenant(ctx context.Context, tenantID string) ([]PortfolioSnapshot, error)
}

// InvestmentService is the interface consumed by the HTTP handler layer.
type InvestmentService interface {
    CreateAsset(ctx context.Context, input CreateAssetInput) (*Asset, error)
    GetAsset(ctx context.Context, id string) (*Asset, error)
    ListAssets(ctx context.Context) ([]Asset, error)

    CreatePosition(ctx context.Context, tenantID string, input CreatePositionInput) (*Position, error)
    GetPosition(ctx context.Context, tenantID, id string) (*Position, error)
    ListPositions(ctx context.Context, tenantID string) ([]Position, error)
    ListPositionsByAccount(ctx context.Context, tenantID, accountID string) ([]Position, error)
    UpdatePosition(ctx context.Context, tenantID, id string, input UpdatePositionInput) (*Position, error)
    DeletePosition(ctx context.Context, tenantID, id string) error

    GetPortfolioSummary(ctx context.Context, tenantID string) (*PortfolioSummary, error)
    TakeSnapshot(ctx context.Context, tenantID string) (*PortfolioSnapshot, error)
}

// PortfolioSummary is the computed read model returned by GET /v1/investments/summary.
type PortfolioSummary struct {
    TotalValueCents    int64               `json:"total_value_cents"`
    Currency           string              `json:"currency"`
    AllocationByType   map[string]int64    `json:"allocation_by_type"`
    Positions          []Position          `json:"positions"`
}
```

---

## 5. Acceptance Criteria

- [ ] `internal/domain/investment.go` compiles with zero errors.
- [ ] All exported types have Go doc comments.
- [ ] All repository interfaces include `context.Context` as first parameter.
- [ ] Every mutating repository method passes `tenantID` for tenant isolation.
- [ ] `AssetRepository.List` has no `tenantID` (global catalogue — justified by ADR 3.1).
- [ ] `internal/domain/investment_test.go` covers any domain validation helpers.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.3 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                 | Type     | Status     |
| ------------------------------------------ | -------- | ---------- |
| Task 3.1 — ADR finalises entity shapes     | Upstream | 🔵 backlog |
| `domain.ErrNotFound` already defined       | Upstream | ✅ done    |
| `domain.ErrForbidden` already defined      | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/domain/investment_test.go`
- **Cases:**
  - `AssetType` constants have correct string values.
  - Confirm `PortfolioSummary` fields are correctly zero-valued upon init (if helper constructors are added).

### Integration tests

N/A — this task produces only domain types and interfaces.

---

## 8. Open Questions

| # | Question                                                                             | Owner | Resolution |
| - | ------------------------------------------------------------------------------------ | ----- | ---------- |
| 1 | Should `Quantity` be `string` (to avoid float) or a dedicated `decimal.Decimal` type? | —   | Decide during Task 3.1 ADR |
| 2 | Should `PortfolioSummary.Currency` be the tenant's base currency or configurable?    | —     | —          |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
