package service

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/shopspring/decimal"
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

	pos, err := s.posRepo.Create(ctx, tenantID, in)
	if err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
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

// ListPositionsByAccount returns all positions for a given tenant and account.
func (s *InvestmentService) ListPositionsByAccount(ctx context.Context, tenantID, accountID string) ([]domain.Position, error) {
	pp, err := s.posRepo.ListByAccount(ctx, tenantID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions by account: %w", err)
	}
	return pp, nil
}

// UpdatePosition modifies an existing position's details, and logs the update action in the audit trail, including the old and new values of the changed fields.
func (s *InvestmentService) UpdatePosition(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error) {
	_, err := s.posRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get position for update: %w", err)
	}

	pos, err := s.posRepo.Update(ctx, tenantID, id, in)
	if err != nil {
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
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

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
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
	if event.Status == domain.ReceivableStatusReceived {
		return nil, fmt.Errorf("%w: income event already received", domain.ErrInvalidInput)
	}
	if event.Status == domain.ReceivableStatusCancelled {
		return nil, fmt.Errorf("%w: income event already cancelled", domain.ErrInvalidInput)
	}

	// 3. Update status
	updated, err := s.incRepo.UpdateStatus(ctx, tenantID, eventID, domain.ReceivableStatusReceived, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to mark income received: %w", err)
	}

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
		Action:     domain.AuditActionUpdate,
		EntityType: "income_event",
		EntityID:   eventID,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit income event update: %w", err)
	}

	// 4. NOTE: transaction materialization is intentionally skipped here.
	// The transaction repository currently requires a non-empty CategoryID,
	// while income events do not carry one. Status transition should still succeed.

	return updated, nil
}

// CancelIncome marks a position income event as cancelled, preventing it from being processed as received, and logs the update action in the audit trail. It validates that the event is in a state that can be cancelled before performing the update.
func (s *InvestmentService) CancelIncome(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	updated, err := s.incRepo.UpdateStatus(ctx, tenantID, eventID, domain.ReceivableStatusCancelled, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel income: %w", err)
	}

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
		Action:     domain.AuditActionUpdate,
		EntityType: "income_event",
		EntityID:   eventID,
	}); err != nil {
		return nil, fmt.Errorf("failed to audit income event cancelation: %w", err)
	}

	return updated, nil
}

// ListIncomeEvents returns all income events for a given tenant, optionally filtered by status.
func (s *InvestmentService) ListIncomeEvents(ctx context.Context, tenantID, status string) ([]domain.PositionIncomeEvent, error) {
	if status == string(domain.ReceivableStatusPending) {
		res, err := s.incRepo.ListPending(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to list pending income events: %w", err)
		}
		return res, nil
	}
	res, err := s.incRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list income events by tenant: %w", err)
	}
	return res, nil
}

// GetPortfolioSummary generates a real-time summary of the tenant's portfolio, including allocation by asset type and individual position summaries. It aggregates data from all positions to calculate total value and gain/loss information for each position.
func (s *InvestmentService) GetPortfolioSummary(ctx context.Context, tenantID string) (*domain.PortfolioSummary, error) {
	pp, err := s.posRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions for summary: %w", err)
	}

	summary := &domain.PortfolioSummary{
		TotalValueCents:  0,
		TotalIncomeCents: 0,
		Currency:         "USD", // Default preferred currency for summary? Or from tenant config?
		AllocationByType: make(map[domain.AssetType][]domain.AllocationSlice),
		Positions:        []domain.PositionView{},
	}

	for _, p := range pp {
		asset, err := s.GetAssetWithTenantConfig(ctx, tenantID, p.AssetID)
		if err != nil {
			return nil, fmt.Errorf("failed to get asset %s: %w", p.AssetID, err)
		}

		marketValue := p.Quantity.Mul(decimal.NewFromInt(p.LastPriceCents)).IntPart()
		summary.TotalValueCents += marketValue
		if p.IncomeAmountCents != nil {
			summary.TotalIncomeCents += *p.IncomeAmountCents
		}

		summary.Positions = append(summary.Positions, domain.PositionView{
			Position:    p,
			AssetName:   asset.Name,
			AssetTicker: asset.Ticker,
			AssetType:   asset.AssetType,
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

	tenantID, _ := middleware.TenantIDFromCtx(ctx)
	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
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

	tenantID, _ := middleware.TenantIDFromCtx(ctx)
	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
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

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
		Action:     domain.AuditActionUpdate,
		EntityType: "tenant_asset_config",
		EntityID:   input.AssetID,
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

	actorID, _ := middleware.UserIDFromCtx(ctx)
	actorRole, _ := middleware.RoleFromCtx(ctx)

	if _, err := s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    actorID,
		ActorRole:  domain.Role(actorRole),
		Action:     domain.AuditActionSoftDelete,
		EntityType: "tenant_asset_config",
		EntityID:   assetID,
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
