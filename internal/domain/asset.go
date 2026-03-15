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
	// Create adds a new asset to the global catalogue.
	Create(ctx context.Context, input CreateAssetInput) (*Asset, error)

	// GetByID retrieves an asset by its unique ID.
	GetByID(ctx context.Context, id string) (*Asset, error)

	// GetByTicker retrieves an asset by its ticker symbol.
	GetByTicker(ctx context.Context, ticker string) (*Asset, error)

	// List returns assets matching the optional filters, with pagination support.
	List(ctx context.Context, params ListAssetsParams) ([]Asset, error)

	// Delete removes an asset from the global catalogue.
	Delete(ctx context.Context, id string) error

	// GetLastPrice retrieves the most recent price for an asset, used in portfolio summaries.
	GetLastPrice(ctx context.Context, id string) (int64, error) // Added for portfolio summary
}

// TenantAssetConfigRepository defines persistence for per-tenant asset overrides.
type TenantAssetConfigRepository interface {
	// Upsert creates or updates a tenant's asset configuration for a specific asset.
	Upsert(ctx context.Context, tenantID string, input UpsertTenantAssetConfigInput) (*TenantAssetConfig, error)

	// GetByAssetID retrieves a tenant's asset configuration for a specific asset.
	GetByAssetID(ctx context.Context, tenantID, assetID string) (*TenantAssetConfig, error)

	// ListByTenant returns all asset configurations for a given tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]TenantAssetConfig, error)

	// Delete removes a tenant's asset configuration for a specific asset.
	Delete(ctx context.Context, tenantID, assetID string) error
}

// InvestmentService orchestrates portfolio tracking.
type InvestmentService interface {
	// CreatePosition manages the lifecycle of an investment position, including asset reference and quantity.
	CreatePosition(ctx context.Context, tenantID string, in CreatePositionInput) (*Position, error)

	// GetPosition retrieves a specific position by ID, ensuring tenant isolation.
	GetPosition(ctx context.Context, tenantID, id string) (*Position, error)

	// ListPositions returns all positions for a tenant, with optional filters for active/inactive.
	ListPositions(ctx context.Context, tenantID string) ([]Position, error)

	// ListPositionsByAccount returns all positions associated with a specific account.
	ListPositionsByAccount(ctx context.Context, tenantID, accountID string) ([]Position, error)

	// UpdatePosition allows modifying certain fields of a position, such as quantity or last price,
	// while enforcing business rules.
	UpdatePosition(ctx context.Context, tenantID, id string, in UpdatePositionInput) (*Position, error)

	// DeletePosition performs a soft delete on a position, marking it as inactive but retaining
	// historical data for portfolio summaries and audit logs.
	DeletePosition(ctx context.Context, tenantID, id string) error

	// MarkIncomeReceived and CancelIncome manage the lifecycle of income events related to positions,
	// such as dividends or interest payments.
	MarkIncomeReceived(ctx context.Context, tenantID, eventID string) (*PositionIncomeEvent, error)

	// CancelIncome reverses a previously marked income event, which may be necessary in cases
	// of erroneous marking or changes in the underlying position.
	CancelIncome(ctx context.Context, tenantID, eventID string) (*PositionIncomeEvent, error)

	// ListIncomeEvents returns all income events for a tenant, with optional status filtering.
	ListIncomeEvents(ctx context.Context, tenantID, status string) ([]PositionIncomeEvent, error)

	// GetPortfolioSummary aggregates the current value of all positions for a tenant,
	// applying the latest asset prices and tenant-specific configurations to provide
	// a comprehensive view of the portfolio's worth and composition.
	GetPortfolioSummary(ctx context.Context, tenantID string) (*PortfolioSummary, error)

	// TakeSnapshot captures the state of a tenant's portfolio at a specific point in time,
	TakeSnapshot(ctx context.Context, tenantID string) (*PortfolioSnapshot, error)

	// CreateAsses is for managing the global asset catalogue, which is referenced by positions and
	// used in portfolio summaries.
	CreateAsset(ctx context.Context, input CreateAssetInput) (*Asset, error)

	// GetAssetByID retrieves a global asset by its ID.
	GetAssetByID(ctx context.Context, id string) (*Asset, error)

	// ListAssets returns global assets with optional filtering, used for reference in
	// position creation and portfolio summaries.
	ListAssets(ctx context.Context, params ListAssetsParams) ([]Asset, error)

	// DeleteAsset removes an asset from the global catalogue, which may have implications
	// for existing positions and summaries, so it should be used with caution.
	DeleteAsset(ctx context.Context, id string) error

	// UpsertTenantAssetConfig allows tenants to override certain fields of global assets, enabling
	// customization of asset names, currencies, or details for their specific portfolio tracking needs.
	UpsertTenantAssetConfig(ctx context.Context, tenantID string, input UpsertTenantAssetConfigInput) (*TenantAssetConfig, error)

	// GetTenantAssetConfig retrieves a tenant's specific configuration for a given asset, which is used to
	// apply overrides in portfolio summaries and position details.
	GetTenantAssetConfig(ctx context.Context, tenantID, assetID string) (*TenantAssetConfig, error)

	// ListTenantAssetConfigs returns all asset configurations for a tenant, which can be used to display
	// the tenant's customized asset settings in the UI or apply them in portfolio calculations.
	ListTenantAssetConfigs(ctx context.Context, tenantID string) ([]TenantAssetConfig, error)

	// DeleteTenantAssetConfig removes a tenant's specific configuration for an asset,
	// causing the tenant to fall back to the global asset values.
	DeleteTenantAssetConfig(ctx context.Context, tenantID, assetID string) error

	// GetAssetWithTenantConfig retrieves the global asset details along with any tenant-specific overrides,
	// providing a complete view of the asset as it applies to the tenant's portfolio.
	// This is particularly useful for portfolio summaries and position details where both
	// global and tenant-specific information is needed.
	GetAssetWithTenantConfig(ctx context.Context, tenantID, id string) (*Asset, error)
}
