package mocks

import (
	"context"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// InvestmentService is a mock implementation of domain.InvestmentService.
type InvestmentService struct {
	mock.Mock
}

var _ domain.InvestmentService = (*InvestmentService)(nil)

func (m *InvestmentService) CreatePosition(ctx context.Context, tenantID string, in domain.CreatePositionInput) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Position), args.Error(1)
}

func (m *InvestmentService) GetPosition(ctx context.Context, tenantID, id string) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Position), args.Error(1)
}

func (m *InvestmentService) ListPositions(ctx context.Context, tenantID string) ([]domain.Position, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.Position
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Position)
	}
	return r0, args.Error(1)
}

func (m *InvestmentService) ListPositionsByAccount(ctx context.Context, tenantID, accountID string) ([]domain.Position, error) {
	args := m.Called(ctx, tenantID, accountID)
	var r0 []domain.Position
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Position)
	}
	return r0, args.Error(1)
}

func (m *InvestmentService) UpdatePosition(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Position), args.Error(1)
}

func (m *InvestmentService) DeletePosition(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *InvestmentService) MarkIncomeReceived(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PositionIncomeEvent), args.Error(1)
}

func (m *InvestmentService) CancelIncome(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PositionIncomeEvent), args.Error(1)
}

func (m *InvestmentService) ListIncomeEvents(ctx context.Context, tenantID, status string) ([]domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, status)
	var r0 []domain.PositionIncomeEvent
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.PositionIncomeEvent)
	}
	return r0, args.Error(1)
}

func (m *InvestmentService) GetPortfolioSummary(ctx context.Context, tenantID string) (*domain.PortfolioSummary, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PortfolioSummary), args.Error(1)
}

func (m *InvestmentService) TakeSnapshot(ctx context.Context, tenantID string) (*domain.PortfolioSnapshot, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PortfolioSnapshot), args.Error(1)
}

func (m *InvestmentService) CreateAsset(ctx context.Context, input domain.CreateAssetInput) (*domain.Asset, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *InvestmentService) GetAssetByID(ctx context.Context, id string) (*domain.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *InvestmentService) ListAssets(ctx context.Context, params domain.ListAssetsParams) ([]domain.Asset, error) {
	args := m.Called(ctx, params)
	var r0 []domain.Asset
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Asset)
	}
	return r0, args.Error(1)
}

func (m *InvestmentService) DeleteAsset(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *InvestmentService) UpsertTenantAssetConfig(ctx context.Context, tenantID string, input domain.UpsertTenantAssetConfigInput) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantAssetConfig), args.Error(1)
}

func (m *InvestmentService) GetTenantAssetConfig(ctx context.Context, tenantID, assetID string) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantAssetConfig), args.Error(1)
}

func (m *InvestmentService) ListTenantAssetConfigs(ctx context.Context, tenantID string) ([]domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.TenantAssetConfig
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.TenantAssetConfig)
	}
	return r0, args.Error(1)
}

func (m *InvestmentService) DeleteTenantAssetConfig(ctx context.Context, tenantID, assetID string) error {
	args := m.Called(ctx, tenantID, assetID)
	return args.Error(0)
}

func (m *InvestmentService) GetAssetWithTenantConfig(ctx context.Context, tenantID, id string) (*domain.Asset, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}
