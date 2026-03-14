package domain

import (
	"context"
	"time"
)

// AssetType categorises the investment instrument.
type AssetType string

const (
	AssetTypeStock        AssetType = "stock"
	AssetTypeBond         AssetType = "bond"
	AssetTypeFund         AssetType = "fund"
	AssetTypeCrypto       AssetType = "crypto"
	AssetTypeRealEstate   AssetType = "real_estate"
	AssetTypeIncomeSource AssetType = "income_source"
)

// Asset is a global, admin-managed reference record — no tenant_id.
type Asset struct {
	CreatedAt time.Time `json:"created_at"`
	ISIN      *string   `json:"isin,omitempty"`
	Details   *string   `json:"details,omitempty"`
	ID        string    `json:"id"`
	Ticker    string    `json:"ticker"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	AssetType AssetType `json:"asset_type"`
}

// CreateAssetInput defines the data required to create a new global asset.
type CreateAssetInput struct {
	ISIN      *string   `json:"isin,omitempty" validate:"omitempty,max=12"`
	Details   *string   `json:"details,omitempty"`
	Ticker    string    `json:"ticker" validate:"required,max=20"`
	Name      string    `json:"name" validate:"required,max=200"`
	Currency  string    `json:"currency" validate:"required,len=3"`
	AssetType AssetType `json:"asset_type" validate:"required"`
}

// ListAssetsParams defines filtering/pagination for global assets.
type ListAssetsParams struct {
	Currency  *string    `json:"currency,omitempty"`
	AssetType *AssetType `json:"asset_type,omitempty"`
	Limit     int32      `json:"limit,omitempty"`
	Offset    int32      `json:"offset,omitempty"`
}

// TenantAssetConfig holds a tenant's sparse overrides for a global asset.
// Fields that are nil fall back to the global asset value.
type TenantAssetConfig struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
	Name      *string    `json:"name,omitempty"`     // overrides Asset.Name
	Currency  *string    `json:"currency,omitempty"` // overrides Asset.Currency
	Details   *string    `json:"details,omitempty"`  // overrides Asset.Details
	ID        string     `json:"id"`
	TenantID  string     `json:"-"`
	AssetID   string     `json:"asset_id"`
}

// UpsertTenantAssetConfigInput defines the data for tenant-specific overrides.
type UpsertTenantAssetConfigInput struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,max=200"`
	Currency *string `json:"currency,omitempty" validate:"omitempty,len=3"`
	Details  *string `json:"details,omitempty"`
	AssetID  string  `json:"asset_id" validate:"required"`
}

// AssetRepository defines persistence operations for the global asset catalogue.
type AssetRepository interface {
	Create(ctx context.Context, input CreateAssetInput) (*Asset, error)
	GetByID(ctx context.Context, id string) (*Asset, error)
	GetByTicker(ctx context.Context, ticker string) (*Asset, error)
	List(ctx context.Context, params ListAssetsParams) ([]Asset, error)
	Delete(ctx context.Context, id string) error
	GetLastPrice(ctx context.Context, id string) (int64, error) // Added for portfolio summary
}

// TenantAssetConfigRepository defines persistence for per-tenant asset overrides.
type TenantAssetConfigRepository interface {
	Upsert(ctx context.Context, tenantID string, input UpsertTenantAssetConfigInput) (*TenantAssetConfig, error)
	GetByAssetID(ctx context.Context, tenantID, assetID string) (*TenantAssetConfig, error)
	ListByTenant(ctx context.Context, tenantID string) ([]TenantAssetConfig, error)
	Delete(ctx context.Context, tenantID, assetID string) error
}

// InvestmentService orchestrates portfolio tracking.
type InvestmentService interface {
	CreatePosition(ctx context.Context, tenantID string, in CreatePositionInput) (*Position, error)
	GetPosition(ctx context.Context, tenantID, id string) (*Position, error)
	ListPositions(ctx context.Context, tenantID string) ([]Position, error)
	UpdatePosition(ctx context.Context, tenantID, id string, in UpdatePositionInput) (*Position, error)
	DeletePosition(ctx context.Context, tenantID, id string) error

	// Receivable lifecycle
	MarkIncomeReceived(ctx context.Context, tenantID, eventID string) (*PositionIncomeEvent, error)
	CancelIncome(ctx context.Context, tenantID, eventID string) (*PositionIncomeEvent, error)

	// Portfolio summary
	GetPortfolioSummary(ctx context.Context, tenantID string) (*PortfolioSummary, error)
	TakeSnapshot(ctx context.Context, tenantID string) (*PortfolioSnapshot, error)

	// Asset Catalogue
	CreateAsset(ctx context.Context, input CreateAssetInput) (*Asset, error)
	GetAssetByID(ctx context.Context, id string) (*Asset, error)
	ListAssets(ctx context.Context, params ListAssetsParams) ([]Asset, error)
	DeleteAsset(ctx context.Context, id string) error

	// Tenant Asset Configuration
	UpsertTenantAssetConfig(ctx context.Context, tenantID string, input UpsertTenantAssetConfigInput) (*TenantAssetConfig, error)
	GetTenantAssetConfig(ctx context.Context, tenantID, assetID string) (*TenantAssetConfig, error)
	ListTenantAssetConfigs(ctx context.Context, tenantID string) ([]TenantAssetConfig, error)
	DeleteTenantAssetConfig(ctx context.Context, tenantID, assetID string) error
	GetAssetWithTenantConfig(ctx context.Context, tenantID, id string) (*Asset, error)
}

// CurrencyConverter normalises values across multiple tickers.
type CurrencyConverter interface {
	Convert(ctx context.Context, amount int64, from, to string) (int64, error)
}
