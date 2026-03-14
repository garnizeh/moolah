package domain

import (
	"context"
	"time"
)

// PositionStatus defines the lifecycle of an investment position.
type PositionStatus string

const (
	PositionStatusOpen   PositionStatus = "open"
	PositionStatusClosed PositionStatus = "closed"
)

// Position represents an investment holding in a specific asset.
type Position struct {
	OpenedAt     time.Time      `json:"opened_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CreatedAt    time.Time      `json:"created_at"`
	Notes        *string        `json:"notes,omitempty"`
	Asset        *Asset         `json:"asset,omitempty"`
	DeletedAt    *time.Time     `json:"-"`
	ClosedAt     *time.Time     `json:"closed_at,omitempty"`
	ID           string         `json:"id"`
	TenantID     string         `json:"-"`
	AccountID    string         `json:"account_id"`
	AssetID      string         `json:"asset_id"`
	Status       PositionStatus `json:"status"`
	Quantity     float64        `json:"quantity"`
	AvgCostCents int64          `json:"avg_cost_cents"`
}

// CreatePositionInput defines the data required to open a new position.
type CreatePositionInput struct {
	OpenedAt     *time.Time `json:"opened_at,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	AccountID    string     `json:"account_id" validate:"required"`
	AssetID      string     `json:"asset_id" validate:"required"`
	Quantity     float64    `json:"quantity" validate:"required,gt=0"`
	AvgCostCents int64      `json:"avg_cost_cents" validate:"required,min=0"`
}

// UpdatePositionInput defines the data required to update an existing position.
type UpdatePositionInput struct {
	Quantity     *float64        `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	AvgCostCents *int64          `json:"avg_cost_cents,omitempty" validate:"omitempty,min=0"`
	Notes        *string         `json:"notes,omitempty"`
	Status       *PositionStatus `json:"status,omitempty"`
	ClosedAt     *time.Time      `json:"closed_at,omitempty"`
}

// PositionSnapshot records the value of a single position at a point in time.
type PositionSnapshot struct {
	SnapshotDate   time.Time `json:"snapshot_date"`
	CreatedAt      time.Time `json:"created_at"`
	ID             string    `json:"id"`
	TenantID       string    `json:"-"`
	PositionID     string    `json:"position_id"`
	Currency       string    `json:"currency"`
	Quantity       float64   `json:"quantity"`
	LastPriceCents int64     `json:"last_price_cents"`
}

// PositionIncomeStatus defines the status of a receivable income event.
type PositionIncomeStatus string

const (
	PositionIncomeStatusPending   PositionIncomeStatus = "pending"
	PositionIncomeStatusReceived  PositionIncomeStatus = "received"
	PositionIncomeStatusCancelled PositionIncomeStatus = "cancelled"
)

// PositionIncomeEvent tracks expected dividends, interest, or rentals.
type PositionIncomeEvent struct {
	EventDate     time.Time            `json:"event_date"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	TransactionID *string              `json:"transaction_id,omitempty"`
	ID            string               `json:"id"`
	TenantID      string               `json:"-"`
	PositionID    string               `json:"position_id"`
	AccountID     string               `json:"account_id"`
	CategoryID    string               `json:"category_id"`
	Currency      string               `json:"currency"`
	IncomeType    string               `json:"income_type"`
	Status        PositionIncomeStatus `json:"status"`
	AmountCents   int64                `json:"amount_cents"`
}

// PortfolioSnapshot groups multiple position snapshots for a tenant.
type PortfolioSnapshot struct {
	SnapshotDate    time.Time `json:"snapshot_date"`
	CreatedAt       time.Time `json:"created_at"`
	ID              string    `json:"id"`
	TenantID        string    `json:"-"`
	Currency        string    `json:"currency"`
	TotalValueCents int64     `json:"total_value_cents"`
}

// PortfolioSummary is a real-time view of all positions.
type PortfolioSummary struct {
	AllocationByType map[AssetType]int64 `json:"allocation_by_type"`
	TenantID         string              `json:"-"`
	Currency         string              `json:"currency"`
	Positions        []PositionSummary   `json:"positions"`
	TotalValueCents  int64               `json:"total_value_cents"`
}

// PositionSummary provides current value info for a specific position.
type PositionSummary struct {
	PositionID     string  `json:"position_id"`
	Ticker         string  `json:"ticker"`
	Name           string  `json:"name"`
	Quantity       float64 `json:"quantity"`
	GainLossPct    float64 `json:"gain_loss_pct"`
	LastPriceCents int64   `json:"last_price_cents"`
	ValueCents     int64   `json:"value_cents"`
	CostCents      int64   `json:"cost_cents"`
	GainLossCents  int64   `json:"gain_loss_cents"`
}

// PositionRepository is the database abstraction for the position table.
type PositionRepository interface {
	Create(ctx context.Context, tenantID string, input *Position) (*Position, error)
	GetByID(ctx context.Context, tenantID, id string) (*Position, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Position, error)
	Update(ctx context.Context, tenantID, id string, input *Position) (*Position, error)
	Delete(ctx context.Context, tenantID, id string) error
}

// PositionIncomeEventRepository is the database abstraction for the position_income_event table.
type PositionIncomeEventRepository interface {
	Create(ctx context.Context, tenantID string, input *PositionIncomeEvent) (*PositionIncomeEvent, error)
	GetByID(ctx context.Context, tenantID, id string) (*PositionIncomeEvent, error)
	ListByPosition(ctx context.Context, tenantID, positionID string) ([]PositionIncomeEvent, error)
	UpdateStatus(ctx context.Context, tenantID, id string, status PositionIncomeStatus, txID *string) (*PositionIncomeEvent, error)
}
