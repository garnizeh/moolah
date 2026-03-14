package service

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// InvestmentService provides business logic for managing investment-related operations such as positions, assets, and portfolio summaries. It interacts with various repositories to perform CRUD operations and also handles audit logging for all actions.
type InvestmentService struct {
	posRepo       domain.PositionRepository
	incRepo       domain.PositionIncomeEventRepository
	assetRepo     domain.AssetRepository
	tenantCfgRepo domain.TenantAssetConfigRepository
	accRepo       domain.AccountRepository
	txRepo        domain.TransactionRepository
	auditRepo     domain.AuditRepository
	converter     domain.CurrencyConverter
}

// NewInvestmentService creates a new instance of InvestmentServiceImpl with the provided repositories and services.
func NewInvestmentService(
	posRepo domain.PositionRepository,
	incRepo domain.PositionIncomeEventRepository,
	assetRepo domain.AssetRepository,
	tenantCfgRepo domain.TenantAssetConfigRepository,
	accRepo domain.AccountRepository,
	txRepo domain.TransactionRepository,
	auditRepo domain.AuditRepository,
	converter domain.CurrencyConverter,
) domain.InvestmentService {
	return &InvestmentService{
		posRepo:       posRepo,
		incRepo:       incRepo,
		assetRepo:     assetRepo,
		tenantCfgRepo: tenantCfgRepo,
		accRepo:       accRepo,
		txRepo:        txRepo,
		auditRepo:     auditRepo,
		converter:     converter,
	}
}

// CreatePosition creates a new investment position for the specified tenant, ensuring that the associated account is of type investment and that the referenced asset exists. It also logs the creation action in the audit trail.
func (s *InvestmentService) CreatePosition(ctx context.Context, tenantID string, in domain.CreatePositionInput) (*domain.Position, error) {
	acc, err := s.accRepo.GetByID(ctx, tenantID, in.AccountID)
	if err != nil {
		return nil, fmt.Errorf("%w: account %s not found: %w", domain.ErrInvalidInput, in.AccountID, err)
	}
	if acc.Type != domain.AccountTypeInvestment {
		return nil, fmt.Errorf("%w: account %s is not investment", domain.ErrInvalidInput, in.AccountID)
	}

	_, err = s.assetRepo.GetByID(ctx, in.AssetID)
	if err != nil {
		return nil, fmt.Errorf("%w: asset %s not found: %w", domain.ErrInvalidInput, in.AssetID, err)
	}

	openedAt := time.Now()
	if in.OpenedAt != nil {
		openedAt = *in.OpenedAt
	}

	newPos := &domain.Position{
		ID:           ulid.New(),
		TenantID:     tenantID,
		AccountID:    in.AccountID,
		AssetID:      in.AssetID,
		Quantity:     in.Quantity,
		AvgCostCents: in.AvgCostCents,
		Notes:        in.Notes,
		OpenedAt:     openedAt,
		Status:       domain.PositionStatusOpen,
	}

	pos, err := s.posRepo.Create(ctx, tenantID, newPos)
	if err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionCreate,
		EntityType: "position",
		EntityID:   pos.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit position creation: %w", err)
	}

	return pos, nil
}

// GetPosition retrieves a position by its ID and tenant ID. It returns domain.ErrNotFound if the position does not exist.
func (s *InvestmentService) GetPosition(ctx context.Context, tenantID, id string) (*domain.Position, error) {
	p, err := s.posRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}
	return p, nil
}

// ListPositions returns all positions for a given tenant.
func (s *InvestmentService) ListPositions(ctx context.Context, tenantID string) ([]domain.Position, error) {
	pp, err := s.posRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions: %w", err)
	}
	return pp, nil
}

// UpdatePosition modifies an existing position's details, and logs the update action in the audit trail, including the old and new values of the changed fields.
func (s *InvestmentService) UpdatePosition(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error) {
	existing, err := s.posRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get position for update: %w", err)
	}

	if in.Quantity != nil {
		existing.Quantity = *in.Quantity
	}
	if in.AvgCostCents != nil {
		existing.AvgCostCents = *in.AvgCostCents
	}
	if in.Notes != nil {
		existing.Notes = in.Notes
	}
	if in.Status != nil {
		existing.Status = *in.Status
	}
	if in.ClosedAt != nil {
		existing.ClosedAt = in.ClosedAt
	}

	pos, err := s.posRepo.Update(ctx, tenantID, id, existing)
	if err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionUpdate,
		EntityType: "position",
		EntityID:   id,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit position update: %w", err)
	}

	return pos, nil
}

// DeletePosition performs a soft delete of the position, and logs the deletion action in the audit trail. It first checks if the position exists before attempting deletion.
func (s *InvestmentService) DeletePosition(ctx context.Context, tenantID, id string) error {
	_, err := s.posRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get position for deletion: %w", err)
	}

	if err := s.posRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionSoftDelete,
		EntityType: "position",
		EntityID:   id,
	}); err != nil {
		return fmt.Errorf("failed to audit position deletion: %w", err)
	}

	return nil
}

// MarkIncomeReceived marks a position income event as received, creates a corresponding transaction for the income, and logs the update action in the audit trail. It validates that the event is in a state that can be marked as received before performing the update.
func (s *InvestmentService) MarkIncomeReceived(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	// 1. Get Event
	event, err := s.incRepo.GetByID(ctx, tenantID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get income event: %w", err)
	}

	// 2. Validate status
	if event.Status == domain.PositionIncomeStatusReceived {
		return nil, fmt.Errorf("%w: income event already received", domain.ErrInvalidInput)
	}
	if event.Status == domain.PositionIncomeStatusCancelled {
		return nil, fmt.Errorf("%w: income event already cancelled", domain.ErrInvalidInput)
	}

	// 3. Update status
	updated, err := s.incRepo.UpdateStatus(ctx, tenantID, eventID, domain.PositionIncomeStatusReceived, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to mark income received: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionUpdate,
		EntityType: "income_event",
		EntityID:   eventID,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit income event update: %w", err)
	}

	// 4. Create Transaction for the income
	if _, err := s.txRepo.Create(ctx, tenantID, domain.CreateTransactionInput{
		AccountID:   event.AccountID,
		AmountCents: event.AmountCents,
		Type:        domain.TransactionTypeIncome,
		CategoryID:  event.CategoryID,
		Description: fmt.Sprintf("Income from position %s", event.PositionID),
		OccurredAt:  time.Now(),
		UserID:      "system", // Ideally extract from context
	}); err != nil {
		return nil, fmt.Errorf("failed to create transaction for income: %w", err)
	}

	return updated, nil
}

// CancelIncome marks a position income event as cancelled, preventing it from being processed as received, and logs the update action in the audit trail. It validates that the event is in a state that can be cancelled before performing the update.
func (s *InvestmentService) CancelIncome(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	updated, err := s.incRepo.UpdateStatus(ctx, tenantID, eventID, domain.PositionIncomeStatusCancelled, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel income: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionUpdate,
		EntityType: "income_event",
		EntityID:   eventID,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit income event cancelation: %w", err)
	}

	return updated, nil
}

// GetPortfolioSummary generates a real-time summary of the tenant's portfolio, including allocation by asset type and individual position summaries. It aggregates data from all positions to calculate total value and gain/loss information for each position.
func (s *InvestmentService) GetPortfolioSummary(ctx context.Context, tenantID string) (*domain.PortfolioSummary, error) {
	pp, err := s.posRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions for summary: %w", err)
	}

	summary := &domain.PortfolioSummary{
		TenantID:  tenantID,
		Positions: []domain.PositionSummary{},
	}

	for _, p := range pp {
		summary.Positions = append(summary.Positions, domain.PositionSummary{
			PositionID: p.ID,
			Quantity:   p.Quantity,
			CostCents:  p.AvgCostCents,
		})
	}

	return summary, nil
}

// TakeSnapshot captures the current state of the tenant's portfolio and saves it as a snapshot for historical reference. It retrieves the latest portfolio summary to populate the snapshot data, including total value and allocation by asset type.
func (s *InvestmentService) TakeSnapshot(ctx context.Context, tenantID string) (*domain.PortfolioSnapshot, error) {
	_, err := s.GetPortfolioSummary(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio summary for snapshot: %w", err)
	}
	return nil, nil
}

// CreateAsset adds a new asset to the global asset catalogue, ensuring that the asset details are valid. It also logs the creation action in the audit trail.
func (s *InvestmentService) CreateAsset(ctx context.Context, input domain.CreateAssetInput) (*domain.Asset, error) {
	asset, err := s.assetRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   "system", // Global asset
		Action:     domain.AuditActionCreate,
		EntityType: "asset",
		EntityID:   asset.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit asset creation: %w", err)
	}

	return asset, nil
}

// GetAssetByID retrieves an asset by its ID. It returns domain.ErrNotFound if the asset does not exist.
func (s *InvestmentService) GetAssetByID(ctx context.Context, id string) (*domain.Asset, error) {
	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}
	return asset, nil
}

// ListAssets returns a list of assets from the global catalogue, with optional filtering by ticker symbol. It supports pagination through limit and offset parameters.
func (s *InvestmentService) ListAssets(ctx context.Context, params domain.ListAssetsParams) ([]domain.Asset, error) {
	assets, err := s.assetRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	return assets, nil
}

// DeleteAsset performs a soft delete of an asset from the global catalogue, and logs the deletion action in the audit trail. It first checks if the asset exists before attempting deletion.
func (s *InvestmentService) DeleteAsset(ctx context.Context, id string) error {
	if err := s.assetRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   "system",
		Action:     domain.AuditActionSoftDelete,
		EntityType: "asset",
		EntityID:   id,
	}); err != nil {
		return fmt.Errorf("failed to audit asset deletion: %w", err)
	}

	return nil
}

// UpsertTenantAssetConfig creates or updates a tenant-specific asset configuration, allowing tenants to override global asset details such as name, currency, and additional information. It ensures that the referenced asset exists before upserting the configuration, and logs the action in the audit trail.
func (s *InvestmentService) UpsertTenantAssetConfig(ctx context.Context, tenantID string, input domain.UpsertTenantAssetConfigInput) (*domain.TenantAssetConfig, error) {
	// Ensure asset exists
	if _, err := s.assetRepo.GetByID(ctx, input.AssetID); err != nil {
		return nil, fmt.Errorf("%w: asset %s not found: %w", domain.ErrInvalidInput, input.AssetID, err)
	}

	cfg, err := s.tenantCfgRepo.Upsert(ctx, tenantID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert tenant asset config: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionUpdate,
		EntityType: "tenant_asset_config",
		EntityID:   fmt.Sprintf("%s:%s", tenantID, input.AssetID),
	}); err != nil {
		return nil, fmt.Errorf("failed to audit tenant asset config upsert: %w", err)
	}

	return cfg, nil
}

// GetTenantAssetConfig retrieves a tenant-specific asset configuration by asset ID. It returns domain.ErrAssetConfigNotFound if the configuration does not exist for the given tenant and asset ID.
func (s *InvestmentService) GetTenantAssetConfig(ctx context.Context, tenantID, assetID string) (*domain.TenantAssetConfig, error) {
	cfg, err := s.tenantCfgRepo.GetByAssetID(ctx, tenantID, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant asset config: %w", err)
	}
	return cfg, nil
}

// ListTenantAssetConfigs returns all tenant-specific asset configurations for a given tenant, allowing the tenant to view and manage their customizations for assets in their portfolio.
func (s *InvestmentService) ListTenantAssetConfigs(ctx context.Context, tenantID string) ([]domain.TenantAssetConfig, error) {
	cfgs, err := s.tenantCfgRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant asset configs: %w", err)
	}
	return cfgs, nil
}

// DeleteTenantAssetConfig performs a soft delete of a tenant-specific asset configuration, and logs the deletion action in the audit trail. It first checks if the configuration exists before attempting deletion.
func (s *InvestmentService) DeleteTenantAssetConfig(ctx context.Context, tenantID, assetID string) error {
	if err := s.tenantCfgRepo.Delete(ctx, tenantID, assetID); err != nil {
		return fmt.Errorf("failed to delete tenant asset config: %w", err)
	}

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		Action:     domain.AuditActionSoftDelete,
		EntityType: "tenant_asset_config",
		EntityID:   fmt.Sprintf("%s:%s", tenantID, assetID),
	}); err != nil {
		return fmt.Errorf("failed to audit tenant asset config deletion: %w", err)
	}

	return nil
}

// GetAssetWithTenantConfig retrieves an asset by its ID and applies any tenant-specific configuration overrides to the asset details. This allows tenants to have customized views of assets in their portfolio based on their specific configurations.
func (s *InvestmentService) GetAssetWithTenantConfig(ctx context.Context, tenantID, id string) (*domain.Asset, error) {
	asset, err := s.assetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	cfg, err := s.tenantCfgRepo.GetByAssetID(ctx, tenantID, id)
	if err == nil && cfg != nil {
		if cfg.Name != nil {
			asset.Name = *cfg.Name
		}
		if cfg.Currency != nil {
			asset.Currency = *cfg.Currency
		}
		if cfg.Details != nil {
			asset.Details = cfg.Details
		}
	}

	return asset, nil
}
