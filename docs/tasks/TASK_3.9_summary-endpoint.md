# Task 3.9 — Domain: Position Family Entities + Repository Interfaces

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Domain Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the domain entities and repository interfaces for the position family: `Position`, `PositionSnapshot`, `PositionIncomeEvent`, and `PortfolioSnapshot`. Also defines the `IncomeType` and `ReceivableStatus` enums and the `PortfolioSummary` view type used by the service and handler layers. No business logic lives here.

---

## 2. Context & Motivation

Task 3.3 covered the asset subsystem. This task covers the capital and income subsystem — the four tables that are tenant-scoped and form the core of portfolio tracking and receivables management. Splitting the domain into two tasks (3.3 and 3.9) keeps each file to a manageable size.

`PositionIncomeEvent` is the receivables ledger (ADR §2.5, §10). Its `status` field (`pending` → `received` | `cancelled`) is the primary lifecycle state tracked by the service layer (Task 3.6).

**Reference:** ADR-003 §2.2, §2.3, §2.5, §3.3–3.6.

---

## 3. Scope

### In scope

- [ ] `IncomeType` string enum with all six constants (`none`, `dividend`, `coupon`, `rent`, `interest`, `salary`).
- [ ] `ReceivableStatus` string enum with three constants (`pending`, `received`, `cancelled`).
- [ ] `Position` struct with all capital + income-schedule fields + JSON tags.
- [ ] `CreatePositionInput`, `UpdatePositionInput` input types.
- [ ] `PositionRepository` interface.
- [ ] `PositionSnapshot` struct + `PositionSnapshotRepository` interface.
- [ ] `PositionIncomeEvent` struct + `PositionIncomeEventRepository` interface.
- [ ] `PortfolioSnapshot` struct (with `DetailsJSON` for the JSONB field) + `PortfolioSnapshotRepository` interface.
- [ ] `PortfolioSummary` view type (computed, not persisted) used by `GetPortfolioSummary`.
- [ ] Sentinel errors: `ErrPositionNotFound`, `ErrIncomeEventNotFound`, `ErrPortfolioSnapshotExists`.
- [ ] Unit tests in `internal/domain/position_test.go`.

### Out of scope

- Asset / TenantAssetConfig entities (Task 3.3).
- CurrencyConverter interface (Task 3.10).
- Concrete implementations (Task 3.12).
- SQLC queries (Task 3.11).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                               | Purpose                                          |
| ------ | ---------------------------------- | ------------------------------------------------ |
| CREATE | `internal/domain/position.go`      | Position-family entities, enums, interfaces      |
| CREATE | `internal/domain/position_test.go` | Unit tests                                       |

### Key types

```go
type IncomeType       string
type ReceivableStatus string

const (
    IncomeTypeNone      IncomeType = "none"
    IncomeTypeDividend  IncomeType = "dividend"
    IncomeTypeCoupon    IncomeType = "coupon"
    IncomeTypeRent      IncomeType = "rent"
    IncomeTypeInterest  IncomeType = "interest"
    IncomeTypeSalary    IncomeType = "salary"
)

const (
    ReceivableStatusPending   ReceivableStatus = "pending"
    ReceivableStatusReceived  ReceivableStatus = "received"
    ReceivableStatusCancelled ReceivableStatus = "cancelled"
)

// Position represents a tenant's holding of a specific asset in an account.
type Position struct {
    ID                 string         `json:"id"`
    TenantID           string         `json:"-"`
    AssetID            string         `json:"asset_id"`
    AccountID          string         `json:"account_id"`
    Quantity           decimal.Decimal `json:"quantity"`    // use shopspring/decimal
    AvgCostCents       int64          `json:"avg_cost_cents"`
    LastPriceCents     int64          `json:"last_price_cents"`
    Currency           string         `json:"currency"`
    PurchasedAt        time.Time      `json:"purchased_at"`
    IncomeType         IncomeType     `json:"income_type"`
    IncomeIntervalDays *int           `json:"income_interval_days,omitempty"`
    IncomeAmountCents  *int64         `json:"income_amount_cents,omitempty"`
    IncomeRateBps      *int           `json:"income_rate_bps,omitempty"`
    NextIncomeAt       *time.Time     `json:"next_income_at,omitempty"`
    MaturityAt         *time.Time     `json:"maturity_at,omitempty"`
    CreatedAt          time.Time      `json:"created_at"`
    UpdatedAt          time.Time      `json:"updated_at"`
    DeletedAt          *time.Time     `json:"-"`
}

// PositionIncomeEvent is one entry in the receivables ledger.
type PositionIncomeEvent struct {
    ID           string           `json:"id"`
    TenantID     string           `json:"-"`
    PositionID   string           `json:"position_id"`
    AccountID    string           `json:"account_id"`
    IncomeType   IncomeType       `json:"income_type"`
    AmountCents  int64            `json:"amount_cents"`
    Currency     string           `json:"currency"`
    DueAt        time.Time        `json:"due_at"`
    ReceivedAt   *time.Time       `json:"received_at,omitempty"`
    Status       ReceivableStatus `json:"status"`
    Notes        *string          `json:"notes,omitempty"`
    CreatedAt    time.Time        `json:"created_at"`
}

// PortfolioSummary is a computed (not persisted) view for the summary endpoint.
type PortfolioSummary struct {
    TotalValueCents  int64                     `json:"total_value_cents"`
    TotalIncomeCents int64                     `json:"total_income_cents"`
    Currency         string                    `json:"currency"`
    AllocationByType map[AssetType]AllocationSlice `json:"allocation_by_type"`
    Positions        []PositionView            `json:"positions"`
}
```

### Repository interfaces

```go
type PositionRepository interface {
    Create(ctx context.Context, tenantID string, in CreatePositionInput) (*Position, error)
    GetByID(ctx context.Context, tenantID, id string) (*Position, error)
    ListByTenant(ctx context.Context, tenantID string) ([]Position, error)
    ListByAccount(ctx context.Context, tenantID, accountID string) ([]Position, error)
    ListDueIncome(ctx context.Context, before time.Time) ([]Position, error) // for income scheduler
    Update(ctx context.Context, tenantID, id string, in UpdatePositionInput) (*Position, error)
    Delete(ctx context.Context, tenantID, id string) error
}

type PositionSnapshotRepository interface {
    Create(ctx context.Context, tenantID string, in CreatePositionSnapshotInput) (*PositionSnapshot, error)
    ListByPosition(ctx context.Context, tenantID, positionID string) ([]PositionSnapshot, error)
    ListByTenantSince(ctx context.Context, tenantID string, since time.Time) ([]PositionSnapshot, error)
}

type PositionIncomeEventRepository interface {
    Create(ctx context.Context, tenantID string, in CreatePositionIncomeEventInput) (*PositionIncomeEvent, error)
    GetByID(ctx context.Context, tenantID, id string) (*PositionIncomeEvent, error)
    ListByTenant(ctx context.Context, tenantID string) ([]PositionIncomeEvent, error)
    ListPending(ctx context.Context, tenantID string) ([]PositionIncomeEvent, error)
    UpdateStatus(ctx context.Context, tenantID, id string, status ReceivableStatus, receivedAt *time.Time) (*PositionIncomeEvent, error)
}

type PortfolioSnapshotRepository interface {
    Create(ctx context.Context, tenantID string, in CreatePortfolioSnapshotInput) (*PortfolioSnapshot, error)
    GetByDate(ctx context.Context, tenantID string, date time.Time) (*PortfolioSnapshot, error)
    ListByTenant(ctx context.Context, tenantID string) ([]PortfolioSnapshot, error)
}
```

---

## 5. Acceptance Criteria

- [ ] All six `IncomeType` constants defined and match ADR enum values exactly.
- [ ] All three `ReceivableStatus` constants defined and match ADR enum values exactly.
- [ ] `Position` struct includes all income-schedule nullable fields as pointers.
- [ ] `PositionRepository.ListDueIncome` exists (required by income scheduler Task 3.13).
- [ ] `PortfolioSummary` view type defined with `AllocationByType` map.
- [ ] All four repository interfaces defined in `internal/domain/`.
- [ ] Sentinel errors exported.
- [ ] Unit tests cover all enum constants.
- [ ] `make task-check` passes.
- [ ] `docs/ROADMAP.md` row 3.9 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                                |
| ---------- | ------ | ----------------------------------------------------- |
| 2026-03-13 | —      | Task repurposed from "summary endpoint" to "position family domain layer" per ADR v3 restructuring |
