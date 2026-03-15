package mocks

import (
	"context"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// AssetRepository is a mock implementation of domain.AssetRepository.
type AssetRepository struct {
	mock.Mock
}

var _ domain.AssetRepository = (*AssetRepository)(nil)

func (m *AssetRepository) Create(ctx context.Context, input domain.CreateAssetInput) (*domain.Asset, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *AssetRepository) GetByID(ctx context.Context, id string) (*domain.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *AssetRepository) GetByTicker(ctx context.Context, ticker string) (*domain.Asset, error) {
	args := m.Called(ctx, ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *AssetRepository) List(ctx context.Context, params domain.ListAssetsParams) ([]domain.Asset, error) {
	args := m.Called(ctx, params)
	var r0 []domain.Asset
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Asset)
	}
	return r0, args.Error(1)
}

func (m *AssetRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AssetRepository) GetLastPrice(ctx context.Context, id string) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

// TenantAssetConfigRepository is a mock implementation of domain.TenantAssetConfigRepository.
type TenantAssetConfigRepository struct {
	mock.Mock
}

var _ domain.TenantAssetConfigRepository = (*TenantAssetConfigRepository)(nil)

func (m *TenantAssetConfigRepository) Upsert(ctx context.Context, tenantID string, input domain.UpsertTenantAssetConfigInput) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantAssetConfig), args.Error(1)
}

func (m *TenantAssetConfigRepository) GetByAssetID(ctx context.Context, tenantID, assetID string) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantAssetConfig), args.Error(1)
}

func (m *TenantAssetConfigRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.TenantAssetConfig
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.TenantAssetConfig)
	}
	return r0, args.Error(1)
}

func (m *TenantAssetConfigRepository) Delete(ctx context.Context, tenantID, assetID string) error {
	args := m.Called(ctx, tenantID, assetID)
	return args.Error(0)
}
