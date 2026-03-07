package sqlc

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockQuerier is a mock implementation of the Querier interface.
// It is placed in this package so it can be reused across all repository tests.
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockQuerier) CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) (AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(AuditLog), args.Error(1)
}

func (m *MockQuerier) CreateCategory(ctx context.Context, arg CreateCategoryParams) (Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Category), args.Error(1)
}

func (m *MockQuerier) CreateOTPRequest(ctx context.Context, arg CreateOTPRequestParams) (OtpRequest, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(OtpRequest), args.Error(1)
}

func (m *MockQuerier) CreateTenant(ctx context.Context, arg CreateTenantParams) (Tenant, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Tenant), args.Error(1)
}

func (m *MockQuerier) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockQuerier) DeleteExpiredOTPs(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockQuerier) GetAccountByID(ctx context.Context, arg GetAccountByIDParams) (Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockQuerier) GetActiveOTPByEmail(ctx context.Context, email string) (OtpRequest, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(OtpRequest), args.Error(1)
}

func (m *MockQuerier) GetCategoryByID(ctx context.Context, arg GetCategoryByIDParams) (Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Category), args.Error(1)
}

func (m *MockQuerier) GetTenantByID(ctx context.Context, id string) (Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Tenant), args.Error(1)
}

func (m *MockQuerier) GetTransactionByID(ctx context.Context, arg GetTransactionByIDParams) (Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockQuerier) GetUserByEmail(ctx context.Context, arg GetUserByEmailParams) (User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockQuerier) GetUserByID(ctx context.Context, arg GetUserByIDParams) (User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockQuerier) ListAccountsByTenant(ctx context.Context, tenantID string) ([]Account, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]Account), args.Error(1)
}

func (m *MockQuerier) ListAuditLogsByEntity(ctx context.Context, arg ListAuditLogsByEntityParams) ([]AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]AuditLog), args.Error(1)
}

func (m *MockQuerier) ListAuditLogsByTenant(ctx context.Context, arg ListAuditLogsByTenantParams) ([]AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]AuditLog), args.Error(1)
}

func (m *MockQuerier) ListCategoriesByTenant(ctx context.Context, tenantID string) ([]Category, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]Category), args.Error(1)
}

func (m *MockQuerier) ListChildCategories(ctx context.Context, arg ListChildCategoriesParams) ([]Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]Category), args.Error(1)
}

func (m *MockQuerier) ListRootCategoriesByTenant(ctx context.Context, tenantID string) ([]Category, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]Category), args.Error(1)
}

func (m *MockQuerier) ListTenants(ctx context.Context) ([]Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Tenant), args.Error(1)
}

func (m *MockQuerier) ListTransactionsByTenant(ctx context.Context, arg ListTransactionsByTenantParams) ([]Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]Transaction), args.Error(1)
}

func (m *MockQuerier) ListUsersByTenant(ctx context.Context, tenantID string) ([]User, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockQuerier) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) SoftDeleteAccount(ctx context.Context, arg SoftDeleteAccountParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) SoftDeleteCategory(ctx context.Context, arg SoftDeleteCategoryParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) SoftDeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) SoftDeleteTransaction(ctx context.Context, arg SoftDeleteTransactionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) SoftDeleteUser(ctx context.Context, arg SoftDeleteUserParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockQuerier) UpdateAccountBalance(ctx context.Context, arg UpdateAccountBalanceParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Category), args.Error(1)
}

func (m *MockQuerier) UpdateTenant(ctx context.Context, arg UpdateTenantParams) (Tenant, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Tenant), args.Error(1)
}

func (m *MockQuerier) UpdateTransaction(ctx context.Context, arg UpdateTransactionParams) (Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockQuerier) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockQuerier) UpdateUserLastLogin(ctx context.Context, arg UpdateUserLastLoginParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

var _ Querier = (*MockQuerier)(nil)
