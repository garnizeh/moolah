package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// AuthRepository is a testify/mock implementation of domain.AuthRepository.
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

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.OTPRequest)
	if !ok {
		return nil, fmt.Errorf("mock CreateOTPRequest: unexpected type %T", args.Get(0))
	}

	return res, err
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

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.OTPRequest)
	if !ok {
		return nil, fmt.Errorf("mock GetActiveOTPRequest: unexpected type %T", args.Get(0))
	}

	return res, err
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

var _ domain.AuthRepository = (*AuthRepository)(nil)

// UserRepository is a testify/mock implementation of domain.UserRepository.
type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.User)
	if !ok {
		return nil, fmt.Errorf("mock UserRepository.Create: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *UserRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.User)
	if !ok {
		return nil, fmt.Errorf("mock UserRepository.GetByID: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.GetByEmail: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.User)
	if !ok {
		return nil, fmt.Errorf("mock UserRepository.GetByEmail: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *UserRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.User, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).([]domain.User)
	if !ok {
		return nil, fmt.Errorf("mock UserRepository.ListByTenant: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *UserRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock UserRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.User)
	if !ok {
		return nil, fmt.Errorf("mock UserRepository.Update: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock UserRepository.UpdateLastLogin: %w", e)
	}
	return nil
}

func (m *UserRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock UserRepository.Delete: %w", e)
	}
	return nil
}

var _ domain.UserRepository = (*UserRepository)(nil)

// AuditRepository is a testify/mock implementation of domain.AuditRepository.
type AuditRepository struct {
	mock.Mock
}

func (m *AuditRepository) Create(ctx context.Context, input domain.CreateAuditLogInput) (*domain.AuditLog, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuditRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.AuditLog)
	if !ok {
		return nil, fmt.Errorf("mock AuditRepository.Create: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *AuditRepository) ListByTenant(ctx context.Context, tenantID string, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuditRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).([]domain.AuditLog)
	if !ok {
		return nil, fmt.Errorf("mock AuditRepository.ListByTenant: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *AuditRepository) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]domain.AuditLog, error) {
	args := m.Called(ctx, tenantID, entityType, entityID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AuditRepository.ListByEntity: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).([]domain.AuditLog)
	if !ok {
		return nil, fmt.Errorf("mock AuditRepository.ListByEntity: unexpected type %T", args.Get(0))
	}

	return res, err
}

var _ domain.AuditRepository = (*AuditRepository)(nil)

// AdminTenantRepository is a mock implementation of domain.AdminTenantRepository.
type AdminTenantRepository struct {
	mock.Mock
}

func (m *AdminTenantRepository) ListAll(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminTenantRepository.ListAll: %w", e)
	}

	res, ok := args.Get(0).([]domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminTenantRepository.ListAll: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminTenantRepository.GetByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminTenantRepository.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminTenantRepository) UpdatePlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	args := m.Called(ctx, id, plan)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminTenantRepository.UpdatePlan: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminTenantRepository.UpdatePlan: unexpected type %T", args.Get(0))
	}
	return res, err
}

// AuthService is a mock implementation of domain.AuthService.
type AuthService struct {
	mock.Mock
}

func (m *AuthService) RequestOTP(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	err := args.Error(0)
	if err != nil {
		return fmt.Errorf("mock RequestOTP: %w", err)
	}
	return nil
}

func (m *AuthService) VerifyOTP(ctx context.Context, email, code string) (*domain.TokenPair, error) {
	args := m.Called(ctx, email, code)
	var err error
	if e := args.Error(1); e != nil {
		err = e
	}

	res, ok := args.Get(0).(*domain.TokenPair)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock VerifyOTP: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	var err error
	if e := args.Error(1); e != nil {
		err = e
	}

	res, ok := args.Get(0).(*domain.TokenPair)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock RefreshToken: unexpected type %T", args.Get(0))
	}
	return res, err
}

var _ domain.AuthService = (*AuthService)(nil)

func (m *AdminTenantRepository) Suspend(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminTenantRepository.Suspend: %w", e)
	}
	return nil
}

func (m *AdminTenantRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminTenantRepository.Restore: %w", e)
	}
	return nil
}

func (m *AdminTenantRepository) HardDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminTenantRepository.HardDelete: %w", e)
	}
	return nil
}

var _ domain.AdminTenantRepository = (*AdminTenantRepository)(nil)

// AdminUserRepository is a mock implementation of domain.AdminUserRepository.
type AdminUserRepository struct {
	mock.Mock
}

func (m *AdminUserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminUserRepository.ListAll: %w", e)
	}

	res, ok := args.Get(0).([]domain.User)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminUserRepository.ListAll: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminUserRepository.GetByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.User)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminUserRepository.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminUserRepository) ForceDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminUserRepository.ForceDelete: %w", e)
	}
	return nil
}

var _ domain.AdminUserRepository = (*AdminUserRepository)(nil)

// AdminAuditRepository is a mock implementation of domain.AdminAuditRepository.
type AdminAuditRepository struct {
	mock.Mock
}

func (m *AdminAuditRepository) ListAll(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminAuditRepository.ListAll: %w", e)
	}

	res, ok := args.Get(0).([]domain.AuditLog)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminAuditRepository.ListAll: unexpected type %T", args.Get(0))
	}
	return res, err
}

var _ domain.AdminAuditRepository = (*AdminAuditRepository)(nil)

// AccountRepository is a testify/mock implementation of domain.AccountRepository.
type AccountRepository struct {
	mock.Mock
}

func (m *AccountRepository) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Account)
	if !ok {
		return nil, fmt.Errorf("mock AccountRepository.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Account)
	if !ok {
		return nil, fmt.Errorf("mock AccountRepository.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).([]domain.Account)
	if !ok {
		return nil, fmt.Errorf("mock AccountRepository.ListByTenant: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountRepository) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID, userID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountRepository.ListByUser: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).([]domain.Account)
	if !ok {
		return nil, fmt.Errorf("mock AccountRepository.ListByUser: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Account)
	if !ok {
		return nil, fmt.Errorf("mock AccountRepository.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountRepository) UpdateBalance(ctx context.Context, tenantID, id string, newBalanceCents int64) error {
	args := m.Called(ctx, tenantID, id, newBalanceCents)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AccountRepository.UpdateBalance: %w", e)
	}
	return nil
}

func (m *AccountRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AccountRepository.Delete: %w", e)
	}
	return nil
}

var _ domain.AccountRepository = (*AccountRepository)(nil)

// CategoryRepository is a testify/mock implementation of domain.CategoryRepository.
type CategoryRepository struct {
	mock.Mock
}

func (m *CategoryRepository) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Category)
	if !ok {
		return nil, fmt.Errorf("mock CategoryRepository.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Category)
	if !ok {
		return nil, fmt.Errorf("mock CategoryRepository.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).([]domain.Category)
	if !ok {
		return nil, fmt.Errorf("mock CategoryRepository.ListByTenant: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryRepository) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID, parentID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryRepository.ListChildren: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).([]domain.Category)
	if !ok {
		return nil, fmt.Errorf("mock CategoryRepository.ListChildren: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Category)
	if !ok {
		return nil, fmt.Errorf("mock CategoryRepository.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock CategoryRepository.Delete: %w", e)
	}
	return nil
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

// TenantRepository is a testify/mock implementation of domain.TenantRepository.
type TenantRepository struct {
	mock.Mock
}

func (m *TenantRepository) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.Tenant)
	if !ok {
		return nil, fmt.Errorf("mock TenantRepository.Create: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *TenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.Tenant)
	if !ok {
		return nil, fmt.Errorf("mock TenantRepository.GetByID: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *TenantRepository) List(ctx context.Context) ([]domain.Tenant, error) {
	args := m.Called(ctx)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantRepository.List: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).([]domain.Tenant)
	if !ok {
		return nil, fmt.Errorf("mock TenantRepository.List: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *TenantRepository) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	// Type assertion with error handling to satisfy govet and errcheck
	res, ok := args.Get(0).(*domain.Tenant)
	if !ok {
		return nil, fmt.Errorf("mock TenantRepository.Update: unexpected type %T", args.Get(0))
	}

	return res, err
}

func (m *TenantRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock TenantRepository.Delete: %w", e)
	}
	return nil
}

var _ domain.TenantRepository = (*TenantRepository)(nil)

// TransactionRepository mock
type TransactionRepository struct {
	mock.Mock
}

func (m *TransactionRepository) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Transaction)
	if !ok {
		return nil, fmt.Errorf("mock TransactionRepository.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Transaction)
	if !ok {
		return nil, fmt.Errorf("mock TransactionRepository.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionRepository) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	args := m.Called(ctx, tenantID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionRepository.List: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).([]domain.Transaction)
	if !ok {
		return nil, fmt.Errorf("mock TransactionRepository.List: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}
	res, ok := args.Get(0).(*domain.Transaction)
	if !ok {
		return nil, fmt.Errorf("mock TransactionRepository.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock TransactionRepository.Delete: %w", e)
	}
	return nil
}

// MasterPurchaseRepository is a testify/mock implementation of domain.MasterPurchaseRepository.
type MasterPurchaseRepository struct {
	mock.Mock
}

func (m *MasterPurchaseRepository) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseRepository.Create: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseRepository.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseRepository.GetByID: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseRepository.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseRepository) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseRepository.ListByTenant: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).([]domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseRepository.ListByTenant: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseRepository) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, accountID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseRepository.ListByAccount: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).([]domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseRepository.ListByAccount: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseRepository) ListPendingClose(ctx context.Context, tenantID string, cutoffDate time.Time) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, cutoffDate)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseRepository.ListPendingClose: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).([]domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseRepository.ListPendingClose: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseRepository) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseRepository.Update: %w", e)
	}
	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseRepository.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseRepository) IncrementPaidInstallments(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock MasterPurchaseRepository.IncrementPaidInstallments: %w", e)
	}
	return nil
}

func (m *MasterPurchaseRepository) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock MasterPurchaseRepository.Delete: %w", e)
	}
	return nil
}

var _ domain.MasterPurchaseRepository = (*MasterPurchaseRepository)(nil)
