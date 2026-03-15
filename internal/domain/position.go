package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// IncomeType categorization for investment returns.
type IncomeType string

const (
	IncomeTypeNone     IncomeType = "none"
	IncomeTypeDividend IncomeType = "dividend"
	IncomeTypeCoupon   IncomeType = "coupon"
	IncomeTypeRent     IncomeType = "rent"
	IncomeTypeInterest IncomeType = "interest"
	IncomeTypeSalary   IncomeType = "salary"
)

// ReceivableStatus lifecycle states for income events.
type ReceivableStatus string

const (
	ReceivableStatusPending   ReceivableStatus = "pending"
	ReceivableStatusReceived  ReceivableStatus = "received"
	ReceivableStatusCancelled ReceivableStatus = "cancelled"
)

// Sentinel errors for the position family.
var (
	ErrPositionNotFound         = errors.New("position not found")
	ErrIncomeEventNotFound      = errors.New("income event not found")
	ErrPortfolioSnapshotExists  = errors.New("portfolio snapshot already exists for this date")
	ErrPositionSnapshotNotFound = errors.New("position snapshot not found")
)

// Position represents a tenant's holding of a specific asset in an account.
// It tracks cost basis, current price, and income scheduling.
type Position struct {
	PurchasedAt        time.Time       `json:"purchased_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	CreatedAt          time.Time       `json:"created_at"`
	IncomeRateBps      *int            `json:"income_rate_bps,omitempty"`
	NextIncomeAt       *time.Time      `json:"next_income_at,omitempty"`
	DeletedAt          *time.Time      `json:"-"`
	MaturityAt         *time.Time      `json:"maturity_at,omitempty"`
	IncomeAmountCents  *int64          `json:"income_amount_cents,omitempty"`
	IncomeIntervalDays *int            `json:"income_interval_days,omitempty"`
	IncomeType         IncomeType      `json:"income_type"`
	AccountID          string          `json:"account_id"`
	Currency           string          `json:"currency"`
	ID                 string          `json:"id"`
	Quantity           decimal.Decimal `json:"quantity"`
	AssetID            string          `json:"asset_id"`
	TenantID           string          `json:"-"`
	LastPriceCents     int64           `json:"last_price_cents"`
	AvgCostCents       int64           `json:"avg_cost_cents"`
}

// CreatePositionInput defines the data required to open a new position.
type CreatePositionInput struct {
	PurchasedAt        time.Time       `json:"purchased_at" validate:"required"`
	IncomeAmountCents  *int64          `json:"income_amount_cents,omitempty" validate:"omitempty,min=0"`
	IncomeIntervalDays *int            `json:"income_interval_days,omitempty" validate:"omitempty,min=1"`
	IncomeRateBps      *int            `json:"income_rate_bps,omitempty" validate:"omitempty,min=0"`
	NextIncomeAt       *time.Time      `json:"next_income_at,omitempty"`
	MaturityAt         *time.Time      `json:"maturity_at,omitempty"`
	Quantity           decimal.Decimal `json:"quantity" validate:"required"`
	Currency           string          `json:"currency" validate:"required,len=3"`
	AccountID          string          `json:"account_id" validate:"required"`
	IncomeType         IncomeType      `json:"income_type" validate:"required"`
	AssetID            string          `json:"asset_id" validate:"required"`
	AvgCostCents       int64           `json:"avg_cost_cents" validate:"required,min=0"`
	LastPriceCents     int64           `json:"last_price_cents" validate:"required,min=0"`
}

// UpdatePositionInput defines the fields that can be modified on an existing position.
type UpdatePositionInput struct {
	Quantity           *decimal.Decimal `json:"quantity,omitempty"`
	AvgCostCents       *int64           `json:"avg_cost_cents,omitempty" validate:"omitempty,min=0"`
	LastPriceCents     *int64           `json:"last_price_cents,omitempty" validate:"omitempty,min=0"`
	IncomeType         *IncomeType      `json:"income_type,omitempty"`
	IncomeIntervalDays *int             `json:"income_interval_days,omitempty" validate:"omitempty,min=1"`
	IncomeAmountCents  *int64           `json:"income_amount_cents,omitempty" validate:"omitempty,min=0"`
	IncomeRateBps      *int             `json:"income_rate_bps,omitempty" validate:"omitempty,min=0"`
	NextIncomeAt       *time.Time       `json:"next_income_at,omitempty"`
	MaturityAt         *time.Time       `json:"maturity_at,omitempty"`
}

// PositionSnapshot tracks the state of a single position at a point in time.
type PositionSnapshot struct {
	SnapshotDate   time.Time       `json:"snapshot_date"`
	CreatedAt      time.Time       `json:"created_at"`
	ID             string          `json:"id"`
	TenantID       string          `json:"-"`
	PositionID     string          `json:"position_id"`
	Quantity       decimal.Decimal `json:"quantity"`
	Currency       string          `json:"currency"`
	LastPriceCents int64           `json:"last_price_cents"`
}

// CreatePositionSnapshotInput defines data for recording a periodic position state.
type CreatePositionSnapshotInput struct {
	SnapshotDate   time.Time       `json:"snapshot_date" validate:"required"`
	PositionID     string          `json:"position_id" validate:"required"`
	Quantity       decimal.Decimal `json:"quantity" validate:"required"`
	Currency       string          `json:"currency" validate:"required,len=3"`
	LastPriceCents int64           `json:"last_price_cents" validate:"required,min=0"`
}

// PositionIncomeEvent tracks individual income payments (receivables).
type PositionIncomeEvent struct {
	DueAt       time.Time        `json:"due_at"`
	CreatedAt   time.Time        `json:"created_at"`
	ReceivedAt  *time.Time       `json:"received_at,omitempty"`
	Notes       *string          `json:"notes,omitempty"`
	ID          string           `json:"id"`
	TenantID    string           `json:"-"`
	PositionID  string           `json:"position_id"`
	AccountID   string           `json:"account_id"`
	IncomeType  IncomeType       `json:"income_type"`
	Currency    string           `json:"currency"`
	Status      ReceivableStatus `json:"status"`
	AmountCents int64            `json:"amount_cents"`
}

// CreatePositionIncomeEventInput defines fields for generating a new receivable.
type CreatePositionIncomeEventInput struct {
	DueAt       time.Time  `json:"due_at" validate:"required"`
	Notes       *string    `json:"notes,omitempty"`
	PositionID  string     `json:"position_id" validate:"required"`
	AccountID   string     `json:"account_id" validate:"required"`
	IncomeType  IncomeType `json:"income_type" validate:"required"`
	Currency    string     `json:"currency" validate:"required,len=3"`
	AmountCents int64      `json:"amount_cents" validate:"required,min=0"`
}

// PortfolioSnapshot is a high-level aggregate of all positions for a tenant.
type PortfolioSnapshot struct {
	SnapshotDate time.Time `json:"snapshot_date"`
	CreatedAt    time.Time `json:"created_at"`
	ID           string    `json:"id"`
	TenantID     string    `json:"-"`
	DetailsJSON  []byte    `json:"details_json"`
}

// CreatePortfolioSnapshotInput defines data for a tenant's portfolio-wide state.
type CreatePortfolioSnapshotInput struct {
	SnapshotDate time.Time `json:"snapshot_date" validate:"required"`
	DetailsJSON  []byte    `json:"details_json" validate:"required"`
}

// PositionView carries denormalized asset info for the summary view.
type PositionView struct {
	AssetName   string    `json:"asset_name"`
	AssetTicker string    `json:"asset_ticker"`
	AssetType   AssetType `json:"asset_type"`
	Position
}

// AllocationSlice represents a percentage of the portfolio.
type AllocationSlice struct {
	Label      string  `json:"label"`
	Percentage float64 `json:"percentage"`
	ValueCents int64   `json:"value_cents"`
}

// PortfolioSummary is a computed view for dashboard and reporting.
type PortfolioSummary struct {
	AllocationByType map[AssetType][]AllocationSlice `json:"allocation_by_type"`
	Currency         string                          `json:"currency"`
	Positions        []PositionView                  `json:"positions"`
	TotalValueCents  int64                           `json:"total_value_cents"`
	TotalIncomeCents int64                           `json:"total_income_cents"`
}

// PositionRepository defines persistence for holdings.
type PositionRepository interface {
	// Create persists a new position for the specified tenant.
	Create(ctx context.Context, tenantID string, in CreatePositionInput) (*Position, error)

	// GetByID retrieves a specific position by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*Position, error)

	// ListByTenant returns all positions for a given tenant, optionally filtered by active/inactive status.
	ListByTenant(ctx context.Context, tenantID string) ([]Position, error)

	// ListByAccount returns all positions associated with a specific account.
	ListByAccount(ctx context.Context, tenantID, accountID string) ([]Position, error)

	// ListDueIncome returns all positions that have pending income events due before the specified time.
	ListDueIncome(ctx context.Context, before time.Time) ([]Position, error)

	// Update modifies an existing position's details, such as quantity or income schedule.
	Update(ctx context.Context, tenantID, id string, in UpdatePositionInput) (*Position, error)

	// Delete performs a soft delete on a position, marking it as inactive but retaining
	// historical data for portfolio summaries and audit logs.
	Delete(ctx context.Context, tenantID, id string) error
}

// PositionSnapshotRepository defines persistence for holding history.
type PositionSnapshotRepository interface {
	// Create records a new snapshot of a position's state at a specific point in time.
	Create(ctx context.Context, tenantID string, in CreatePositionSnapshotInput) (*PositionSnapshot, error)

	// ListByPosition returns the historical snapshots for a specific position, ordered by date.
	ListByPosition(ctx context.Context, tenantID, positionID string) ([]PositionSnapshot, error)

	// ListByTenantSince returns all position snapshots for a tenant since a given date, used for portfolio history.
	ListByTenantSince(ctx context.Context, tenantID string, since time.Time) ([]PositionSnapshot, error)
}

// PositionIncomeEventRepository defines persistence for the receivables ledger.
type PositionIncomeEventRepository interface {
	// Create adds a new income event to the ledger for a specific position.
	Create(ctx context.Context, tenantID string, in CreatePositionIncomeEventInput) (*PositionIncomeEvent, error)

	// GetByID retrieves a specific income event by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*PositionIncomeEvent, error)

	// ListByPosition returns all income events associated with a specific position.
	ListByTenant(ctx context.Context, tenantID string) ([]PositionIncomeEvent, error)

	// ListPending returns all income events for a tenant that are currently pending, used by the income scheduler.
	ListPending(ctx context.Context, tenantID string) ([]PositionIncomeEvent, error)

	// UpdateStatus changes the status of an income event, such as marking it as received or cancelled.
	UpdateStatus(ctx context.Context, tenantID, id string, status ReceivableStatus, receivedAt *time.Time) (*PositionIncomeEvent, error)
}

// PortfolioSnapshotRepository defines persistence for aggregate portfolio history.
type PortfolioSnapshotRepository interface {
	// Create records a new portfolio snapshot for a tenant at a specific date.
	Create(ctx context.Context, tenantID string, in CreatePortfolioSnapshotInput) (*PortfolioSnapshot, error)

	// GetByDate retrieves a portfolio snapshot for a tenant by the snapshot date, used for historical views.
	GetByDate(ctx context.Context, tenantID string, date time.Time) (*PortfolioSnapshot, error)

	// ListByTenant returns all portfolio snapshots for a tenant, ordered by date, used for historical trends.
	ListByTenant(ctx context.Context, tenantID string) ([]PortfolioSnapshot, error)
}
