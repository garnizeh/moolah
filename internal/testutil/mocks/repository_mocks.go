package mocks

import (
	"context"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// UserRepository is a mock implementation of domain.UserRepository.
type UserRepository struct {
	mock.Mock
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (m *UserRepository) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.User
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.User)
	}
	return r0, args.Error(1)
}

func (m *UserRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *UserRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// AccountRepository is a mock implementation of domain.AccountRepository.
type AccountRepository struct {
	mock.Mock
}

var _ domain.AccountRepository = (*AccountRepository)(nil)

func (m *AccountRepository) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *AccountRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *AccountRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.Account
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Account)
	}
	return r0, args.Error(1)
}

func (m *AccountRepository) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID, userID)
	var r0 []domain.Account
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Account)
	}
	return r0, args.Error(1)
}

func (m *AccountRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *AccountRepository) UpdateBalance(ctx context.Context, tenantID, id string, newBalanceCents int64) error {
	args := m.Called(ctx, tenantID, id, newBalanceCents)
	return args.Error(0)
}

func (m *AccountRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// CategoryRepository is a mock implementation of domain.CategoryRepository.
type CategoryRepository struct {
	mock.Mock
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

func (m *CategoryRepository) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.Category
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Category)
	}
	return r0, args.Error(1)
}

func (m *CategoryRepository) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID, parentID)
	var r0 []domain.Category
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Category)
	}
	return r0, args.Error(1)
}

func (m *CategoryRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// TransactionRepository is a mock implementation of domain.TransactionRepository.
type TransactionRepository struct {
	mock.Mock
}

var _ domain.TransactionRepository = (*TransactionRepository)(nil)

func (m *TransactionRepository) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *TransactionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *TransactionRepository) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	args := m.Called(ctx, tenantID, params)
	var r0 []domain.Transaction
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Transaction)
	}
	return r0, args.Error(1)
}

func (m *TransactionRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *TransactionRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// AuditRepository is a mock implementation of domain.AuditRepository.
type AuditRepository struct {
	mock.Mock
}

var _ domain.AuditRepository = (*AuditRepository)(nil)

func (m *AuditRepository) Create(ctx context.Context, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuditLog), args.Error(1)
}

func (m *AuditRepository) ListByTenant(ctx context.Context, tenantID string, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, params)
	var r0 []domain.AuditLog
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.AuditLog)
	}
	return r0, args.Error(1)
}

func (m *AuditRepository) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, entityType, entityID)
	var r0 []domain.AuditLog
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.AuditLog)
	}
	return r0, args.Error(1)
}

// IdempotencyStore is a mock implementation of domain.IdempotencyStore.
type IdempotencyStore struct {
	mock.Mock
}

var _ domain.IdempotencyStore = (*IdempotencyStore)(nil)

func (m *IdempotencyStore) Get(ctx context.Context, key string) (*domain.CachedResponse, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CachedResponse), args.Error(1)
}

func (m *IdempotencyStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *IdempotencyStore) SetResponse(ctx context.Context, key string, resp domain.CachedResponse, ttl time.Duration) error {
	args := m.Called(ctx, key, resp, ttl)
	return args.Error(0)
}

// AuthRepository is a mock implementation of domain.AuthRepository.
type AuthRepository struct {
	mock.Mock
}

var _ domain.AuthRepository = (*AuthRepository)(nil)

func (m *AuthRepository) CreateOTPRequest(ctx context.Context, input domain.CreateOTPRequestInput) (*domain.OTPRequest, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPRequest), args.Error(1)
}

func (m *AuthRepository) GetActiveOTPRequest(ctx context.Context, email string) (*domain.OTPRequest, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OTPRequest), args.Error(1)
}

func (m *AuthRepository) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AuthRepository) DeleteExpiredOTPRequests(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TenantRepository is a mock implementation of domain.TenantRepository.
type TenantRepository struct {
	mock.Mock
}

var _ domain.TenantRepository = (*TenantRepository)(nil)

func (m *TenantRepository) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantRepository) List(ctx context.Context) ([]domain.Tenant, error) {
	args := m.Called(ctx)
	var r0 []domain.Tenant
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Tenant)
	}
	return r0, args.Error(1)
}

func (m *TenantRepository) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MasterPurchaseRepository is a mock implementation of domain.MasterPurchaseRepository.
type MasterPurchaseRepository struct {
	mock.Mock
}

var _ domain.MasterPurchaseRepository = (*MasterPurchaseRepository)(nil)

func (m *MasterPurchaseRepository) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MasterPurchase), args.Error(1)
}

func (m *MasterPurchaseRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MasterPurchase), args.Error(1)
}

func (m *MasterPurchaseRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.MasterPurchase
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.MasterPurchase)
	}
	return r0, args.Error(1)
}

func (m *MasterPurchaseRepository) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, accountID)
	var r0 []domain.MasterPurchase
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.MasterPurchase)
	}
	return r0, args.Error(1)
}

func (m *MasterPurchaseRepository) ListPendingClose(ctx context.Context, tenantID string, cutoffDate time.Time) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, cutoffDate)
	var r0 []domain.MasterPurchase
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.MasterPurchase)
	}
	return r0, args.Error(1)
}

func (m *MasterPurchaseRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MasterPurchase), args.Error(1)
}

func (m *MasterPurchaseRepository) IncrementPaidInstallments(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// Mailer is a mock implementation of domain.Mailer.
type Mailer struct {
	mock.Mock
}

var _ domain.Mailer = (*Mailer)(nil)

func (m *Mailer) SendOTP(ctx context.Context, to, code string) error {
	args := m.Called(ctx, to, code)
	return args.Error(0)
}

func (m *MasterPurchaseRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// AdminTenantRepository is a mock implementation of domain.AdminTenantRepository.
type AdminTenantRepository struct {
	mock.Mock
}

var _ domain.AdminTenantRepository = (*AdminTenantRepository)(nil)

func (m *AdminTenantRepository) ListAll(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	var r0 []domain.Tenant
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Tenant)
	}
	return r0, args.Error(1)
}

func (m *AdminTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *AdminTenantRepository) UpdatePlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	args := m.Called(ctx, id, plan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *AdminTenantRepository) Suspend(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AdminTenantRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AdminTenantRepository) HardDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// AdminUserRepository is a mock implementation of domain.AdminUserRepository.
type AdminUserRepository struct {
	mock.Mock
}

var _ domain.AdminUserRepository = (*AdminUserRepository)(nil)

func (m *AdminUserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	var r0 []domain.User
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.User)
	}
	return r0, args.Error(1)
}

func (m *AdminUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *AdminUserRepository) ForceDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// AdminAuditRepository is a mock implementation of domain.AdminAuditRepository.
type AdminAuditRepository struct {
	mock.Mock
}

var _ domain.AdminAuditRepository = (*AdminAuditRepository)(nil)

func (m *AdminAuditRepository) ListAll(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, params)
	var r0 []domain.AuditLog
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.AuditLog)
	}
	return r0, args.Error(1)
}
