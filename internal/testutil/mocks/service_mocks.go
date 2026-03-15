package mocks

import (
	"context"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// AccountService is a mock implementation of domain.AccountService.
type AccountService struct {
	mock.Mock
}

var _ domain.AccountService = (*AccountService)(nil)

func (m *AccountService) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *AccountService) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *AccountService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.Account
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Account)
	}
	return r0, args.Error(1)
}

func (m *AccountService) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID, userID)
	var r0 []domain.Account
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Account)
	}
	return r0, args.Error(1)
}

func (m *AccountService) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *AccountService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// CategoryService is a mock implementation of domain.CategoryService.
type CategoryService struct {
	mock.Mock
}

var _ domain.CategoryService = (*CategoryService)(nil)

func (m *CategoryService) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryService) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.Category
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Category)
	}
	return r0, args.Error(1)
}

func (m *CategoryService) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID, parentID)
	var r0 []domain.Category
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Category)
	}
	return r0, args.Error(1)
}

func (m *CategoryService) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// TransactionService is a mock implementation of domain.TransactionService.
type TransactionService struct {
	mock.Mock
}

var _ domain.TransactionService = (*TransactionService)(nil)

func (m *TransactionService) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *TransactionService) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *TransactionService) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	args := m.Called(ctx, tenantID, params)
	var r0 []domain.Transaction
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Transaction)
	}
	return r0, args.Error(1)
}

func (m *TransactionService) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *TransactionService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

// AuthService is a mock implementation of domain.AuthService.
type AuthService struct {
	mock.Mock
}

var _ domain.AuthService = (*AuthService)(nil)

func (m *AuthService) RequestOTP(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *AuthService) VerifyOTP(ctx context.Context, email, code string) (*domain.TokenPair, error) {
	args := m.Called(ctx, email, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenPair), args.Error(1)
}

func (m *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenPair), args.Error(1)
}

// TenantService is a mock implementation of domain.TenantService.
type TenantService struct {
	mock.Mock
}

var _ domain.TenantService = (*TenantService)(nil)

func (m *TenantService) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantService) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantService) List(ctx context.Context) ([]domain.Tenant, error) {
	args := m.Called(ctx)
	var r0 []domain.Tenant
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Tenant)
	}
	return r0, args.Error(1)
}

func (m *TenantService) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TenantService) InviteUser(ctx context.Context, tenantID string, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// AdminService is a mock implementation of domain.AdminService.
type AdminService struct {
	mock.Mock
}

var _ domain.AdminService = (*AdminService)(nil)

func (m *AdminService) ListAllTenants(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	var r0 []domain.Tenant
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.Tenant)
	}
	return r0, args.Error(1)
}

func (m *AdminService) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *AdminService) UpdateTenantPlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	args := m.Called(ctx, id, plan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *AdminService) SuspendTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AdminService) RestoreTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AdminService) HardDeleteTenant(ctx context.Context, id, confirmationToken string) error {
	args := m.Called(ctx, id, confirmationToken)
	return args.Error(0)
}

func (m *AdminService) ListAllUsers(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	var r0 []domain.User
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.User)
	}
	return r0, args.Error(1)
}

func (m *AdminService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *AdminService) ForceDeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AdminService) ListAuditLogs(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, params)
	var r0 []domain.AuditLog
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.AuditLog)
	}
	return r0, args.Error(1)
}

// InvoiceCloser is a mock implementation of domain.InvoiceCloser.
type InvoiceCloser struct {
	mock.Mock
}

var _ domain.InvoiceCloser = (*InvoiceCloser)(nil)

func (m *InvoiceCloser) CloseInvoice(ctx context.Context, tenantID, accountID string, closingDate time.Time) (domain.CloseInvoiceResult, error) {
	args := m.Called(ctx, tenantID, accountID, closingDate)
	return args.Get(0).(domain.CloseInvoiceResult), args.Error(1)
}

// CurrencyConverter is a mock implementation of domain.CurrencyConverter.
type CurrencyConverter struct {
	mock.Mock
}

var _ domain.CurrencyConverter = (*CurrencyConverter)(nil)

func (m *CurrencyConverter) Convert(ctx context.Context, amountCents int64, fromCurrency, toCurrency string) (int64, error) {
	args := m.Called(ctx, amountCents, fromCurrency, toCurrency)
	return args.Get(0).(int64), args.Error(1)
}

// MasterPurchaseService is a mock implementation of domain.MasterPurchaseService.
type MasterPurchaseService struct {
	mock.Mock
}

var _ domain.MasterPurchaseService = (*MasterPurchaseService)(nil)

func (m *MasterPurchaseService) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MasterPurchase), args.Error(1)
}

func (m *MasterPurchaseService) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MasterPurchase), args.Error(1)
}

func (m *MasterPurchaseService) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID)
	var r0 []domain.MasterPurchase
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.MasterPurchase)
	}
	return r0, args.Error(1)
}

func (m *MasterPurchaseService) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, accountID)
	var r0 []domain.MasterPurchase
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.MasterPurchase)
	}
	return r0, args.Error(1)
}

func (m *MasterPurchaseService) ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment {
	args := m.Called(mp)
	var r0 []domain.ProjectedInstallment
	if arg := args.Get(0); arg != nil {
		r0 = arg.([]domain.ProjectedInstallment)
	}
	return r0
}

func (m *MasterPurchaseService) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MasterPurchase), args.Error(1)
}

func (m *MasterPurchaseService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}
