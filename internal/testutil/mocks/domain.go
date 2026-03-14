package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// PositionRepository mock
type PositionRepository struct {
	mock.Mock
}

func (m *PositionRepository) Create(ctx context.Context, tenantID string, input *domain.Position) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Position), err //nolint:errcheck
}

func (m *PositionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Position), err //nolint:errcheck
}

func (m *PositionRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Position, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Position), err //nolint:errcheck
}

func (m *PositionRepository) Update(ctx context.Context, tenantID, id string, input *domain.Position) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Position), err //nolint:errcheck
}

func (m *PositionRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// PositionIncomeEventRepository mock
type PositionIncomeEventRepository struct {
	mock.Mock
}

func (m *PositionIncomeEventRepository) Create(ctx context.Context, tenantID string, input *domain.PositionIncomeEvent) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.PositionIncomeEvent), err //nolint:errcheck
}

func (m *PositionIncomeEventRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.PositionIncomeEvent), err //nolint:errcheck
}

func (m *PositionIncomeEventRepository) ListByPosition(ctx context.Context, tenantID, positionID string) ([]domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, positionID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByPosition: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.PositionIncomeEvent), err //nolint:errcheck
}

func (m *PositionIncomeEventRepository) UpdateStatus(ctx context.Context, tenantID, id string, status domain.PositionIncomeStatus, txID *string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, id, status, txID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UpdateStatus: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.PositionIncomeEvent), err //nolint:errcheck
}

// AssetRepository mock
type AssetRepository struct {
	mock.Mock
}

func (m *AssetRepository) GetByID(ctx context.Context, id string) (*domain.Asset, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Asset), err //nolint:errcheck
}

func (m *AssetRepository) Create(ctx context.Context, in domain.CreateAssetInput) (*domain.Asset, error) {
	args := m.Called(ctx, in)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Asset), err //nolint:errcheck
}

func (m *AssetRepository) GetByTicker(ctx context.Context, ticker string) (*domain.Asset, error) {
	args := m.Called(ctx, ticker)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByTicker: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Asset), err //nolint:errcheck
}

func (m *AssetRepository) List(ctx context.Context, params domain.ListAssetsParams) ([]domain.Asset, error) {
	args := m.Called(ctx, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock List: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Asset), err //nolint:errcheck
}

func (m *AssetRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

func (m *AssetRepository) GetLastPrice(ctx context.Context, id string) (int64, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetLastPrice: %w", e)
	}
	res, _ := args.Get(0).(int64) //nolint:errcheck
	return res, err
}

// TenantAssetConfigRepository mock
type TenantAssetConfigRepository struct {
	mock.Mock
}

func (m *TenantAssetConfigRepository) Upsert(ctx context.Context, tenantID string, input domain.UpsertTenantAssetConfigInput) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Upsert: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.TenantAssetConfig), err //nolint:errcheck
}

func (m *TenantAssetConfigRepository) GetByAssetID(ctx context.Context, tenantID, assetID string) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, assetID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByAssetID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.TenantAssetConfig), err //nolint:errcheck
}

func (m *TenantAssetConfigRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.TenantAssetConfig), err //nolint:errcheck
}

func (m *TenantAssetConfigRepository) Delete(ctx context.Context, tenantID, assetID string) error {
	args := m.Called(ctx, tenantID, assetID)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// AccountRepository mock
type AccountRepository struct {
	mock.Mock
}

func (m *AccountRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Account), err //nolint:errcheck
}

func (m *AccountRepository) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Account), err //nolint:errcheck
}

func (m *AccountRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Account), err //nolint:errcheck
}

func (m *AccountRepository) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID, userID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByUser: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Account), err //nolint:errcheck
}

func (m *AccountRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Account), err //nolint:errcheck
}

func (m *AccountRepository) UpdateBalance(ctx context.Context, tenantID, id string, newBalanceCents int64) error {
	args := m.Called(ctx, tenantID, id, newBalanceCents)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock UpdateBalance: %w", e)
	}
	return nil
}

func (m *AccountRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// TransactionRepository mock
type TransactionRepository struct {
	mock.Mock
}

func (m *TransactionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Transaction), err //nolint:errcheck
}

func (m *TransactionRepository) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Transaction), err //nolint:errcheck
}

func (m *TransactionRepository) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	args := m.Called(ctx, tenantID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock List: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Transaction), err //nolint:errcheck
}

func (m *TransactionRepository) ListByAccount(ctx context.Context, tenantID, accountID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	args := m.Called(ctx, tenantID, accountID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByAccount: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Transaction), err //nolint:errcheck
}

func (m *TransactionRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Transaction), err //nolint:errcheck
}

func (m *TransactionRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// AuditRepository mock
type AuditRepository struct {
	mock.Mock
}

func (m *AuditRepository) Create(ctx context.Context, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.AuditLog), err //nolint:errcheck
}

func (m *AuditRepository) ListByTenant(ctx context.Context, tenantID string, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.AuditLog), err //nolint:errcheck
}

func (m *AuditRepository) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, entityType, entityID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByEntity: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.AuditLog), err //nolint:errcheck
}

// AuthService mock
type AuthService struct {
	mock.Mock
}

func (m *AuthService) RequestOTP(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock RequestOTP: %w", e)
	}
	return nil
}

func (m *AuthService) VerifyOTP(ctx context.Context, email, code string) (*domain.TokenPair, error) {
	args := m.Called(ctx, email, code)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock VerifyOTP: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.TokenPair), err //nolint:errcheck
}

func (m *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock RefreshToken: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.TokenPair), err //nolint:errcheck
}

// UserRepository mock
type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err //nolint:errcheck
}

func (m *UserRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err //nolint:errcheck
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByEmail: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err //nolint:errcheck
}

func (m *UserRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.User), err //nolint:errcheck
}

func (m *UserRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err //nolint:errcheck
}

func (m *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock UpdateLastLogin: %w", e)
	}
	return nil
}

func (m *UserRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// CategoryRepository mock
type CategoryRepository struct {
	mock.Mock
}

func (m *CategoryRepository) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Category), err //nolint:errcheck
}

func (m *CategoryRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Category), err //nolint:errcheck
}

func (m *CategoryRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Category), err //nolint:errcheck
}

func (m *CategoryRepository) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID, parentID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListChildren: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Category), err //nolint:errcheck
}

func (m *CategoryRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Category), err //nolint:errcheck
}

func (m *CategoryRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// MasterPurchaseRepository mock
type MasterPurchaseRepository struct {
	mock.Mock
}

func (m *MasterPurchaseRepository) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.MasterPurchase), err //nolint:errcheck
}

func (m *MasterPurchaseRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.MasterPurchase), err //nolint:errcheck
}

func (m *MasterPurchaseRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.MasterPurchase), err //nolint:errcheck
}

func (m *MasterPurchaseRepository) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, accountID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListByAccount: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.MasterPurchase), err //nolint:errcheck
}

func (m *MasterPurchaseRepository) ListPendingClose(ctx context.Context, tenantID string, cutoffDate time.Time) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, cutoffDate)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListPendingClose: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.MasterPurchase), err //nolint:errcheck
}

func (m *MasterPurchaseRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.MasterPurchase), err //nolint:errcheck
}

func (m *MasterPurchaseRepository) IncrementPaidInstallments(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock IncrementPaidInstallments: %w", e)
	}
	return nil
}

func (m *MasterPurchaseRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// AdminTenantRepository mock
type AdminTenantRepository struct {
	mock.Mock
}

func (m *AdminTenantRepository) ListAll(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListAll: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Tenant), err //nolint:errcheck
}

func (m *AdminTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Tenant), err //nolint:errcheck
}

func (m *AdminTenantRepository) UpdatePlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	args := m.Called(ctx, id, plan)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UpdatePlan: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Tenant), err //nolint:errcheck
}

func (m *AdminTenantRepository) Suspend(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Suspend: %w", e)
	}
	return nil
}

func (m *AdminTenantRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Restore: %w", e)
	}
	return nil
}

func (m *AdminTenantRepository) HardDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock HardDelete: %w", e)
	}
	return nil
}

// AdminUserRepository mock
type AdminUserRepository struct {
	mock.Mock
}

func (m *AdminUserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListAll: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.User), err //nolint:errcheck
}

func (m *AdminUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.User), err //nolint:errcheck
}

func (m *AdminUserRepository) ForceDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock ForceDelete: %w", e)
	}
	return nil
}

// AdminAuditRepository mock
type AdminAuditRepository struct {
	mock.Mock
}

func (m *AdminAuditRepository) ListAll(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock ListAll: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.AuditLog), err //nolint:errcheck
}

// AuthRepository mock
type AuthRepository struct {
	mock.Mock
}

func (m *AuthRepository) CreateOTPRequest(ctx context.Context, input domain.CreateOTPRequestInput) (*domain.OTPRequest, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CreateOTPRequest: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.OTPRequest), err //nolint:errcheck
}

func (m *AuthRepository) GetActiveOTPRequest(ctx context.Context, email string) (*domain.OTPRequest, error) {
	args := m.Called(ctx, email)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetActiveOTPRequest: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.OTPRequest), err //nolint:errcheck
}

func (m *AuthRepository) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock MarkOTPUsed: %w", e)
	}
	return nil
}

func (m *AuthRepository) DeleteExpiredOTPRequests(ctx context.Context) error {
	args := m.Called(ctx)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock DeleteExpiredOTPRequests: %w", e)
	}
	return nil
}

// TenantRepository mock
type TenantRepository struct {
	mock.Mock
}

func (m *TenantRepository) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Tenant), err //nolint:errcheck
}

func (m *TenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Tenant), err //nolint:errcheck
}

func (m *TenantRepository) List(ctx context.Context) ([]domain.Tenant, error) {
	args := m.Called(ctx)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock List: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]domain.Tenant), err //nolint:errcheck
}

func (m *TenantRepository) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).(*domain.Tenant), err //nolint:errcheck
}

func (m *TenantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock Delete: %w", e)
	}
	return nil
}

// CurrencyConverter mock
type CurrencyConverter struct {
	mock.Mock
}

func (m *CurrencyConverter) Convert(ctx context.Context, amount int64, from, to string) (int64, error) {
	args := m.Called(ctx, amount, from, to)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock Convert: %w", e)
	}
	res, _ := args.Get(0).(int64) //nolint:errcheck
	return res, err
}

var (
	_ domain.PositionRepository            = (*PositionRepository)(nil)
	_ domain.PositionIncomeEventRepository = (*PositionIncomeEventRepository)(nil)
	_ domain.AssetRepository               = (*AssetRepository)(nil)
	_ domain.TenantAssetConfigRepository   = (*TenantAssetConfigRepository)(nil)
	_ domain.AccountRepository             = (*AccountRepository)(nil)
	_ domain.TransactionRepository         = (*TransactionRepository)(nil)
	_ domain.AuditRepository               = (*AuditRepository)(nil)
	_ domain.AuthService                   = (*AuthService)(nil)
	_ domain.UserRepository                = (*UserRepository)(nil)
	_ domain.CategoryRepository            = (*CategoryRepository)(nil)
	_ domain.MasterPurchaseRepository      = (*MasterPurchaseRepository)(nil)
	_ domain.AdminTenantRepository         = (*AdminTenantRepository)(nil)
	_ domain.AdminUserRepository           = (*AdminUserRepository)(nil)
	_ domain.AdminAuditRepository          = (*AdminAuditRepository)(nil)
	_ domain.AuthRepository                = (*AuthRepository)(nil)
	_ domain.TenantRepository              = (*TenantRepository)(nil)
	_ domain.CurrencyConverter             = (*CurrencyConverter)(nil)
)
