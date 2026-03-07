package mocks

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"

	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
)

// Querier is a testify/mock implementation of sqlc.Querier.
// It centralises the mock so all repository, service, and handler tests can import it.
type Querier struct {
	mock.Mock
}

func (m *Querier) CreateAccount(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Account{}, fmt.Errorf("mock querier CreateAccount: %w", err)
	}
	return args.Get(0).(sqlc.Account), nil
}

func (m *Querier) CreateAuditLog(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.AuditLog{}, fmt.Errorf("mock querier CreateAuditLog: %w", err)
	}
	return args.Get(0).(sqlc.AuditLog), nil
}

func (m *Querier) CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Category{}, fmt.Errorf("mock querier CreateCategory: %w", err)
	}
	return args.Get(0).(sqlc.Category), nil
}

func (m *Querier) CreateOTPRequest(ctx context.Context, arg sqlc.CreateOTPRequestParams) (sqlc.OtpRequest, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.OtpRequest{}, fmt.Errorf("mock querier CreateOTPRequest: %w", err)
	}
	return args.Get(0).(sqlc.OtpRequest), nil
}

func (m *Querier) CreateTenant(ctx context.Context, arg sqlc.CreateTenantParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Tenant{}, fmt.Errorf("mock querier CreateTenant: %w", err)
	}
	return args.Get(0).(sqlc.Tenant), nil
}

func (m *Querier) CreateTransaction(ctx context.Context, arg sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Transaction{}, fmt.Errorf("mock querier CreateTransaction: %w", err)
	}
	return args.Get(0).(sqlc.Transaction), nil
}

func (m *Querier) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.User{}, fmt.Errorf("mock querier CreateUser: %w", err)
	}
	return args.Get(0).(sqlc.User), nil
}

func (m *Querier) DeleteExpiredOTPs(ctx context.Context) error {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier DeleteExpiredOTPs: %w", err)
	}
	return nil
}

func (m *Querier) GetAccountByID(ctx context.Context, arg sqlc.GetAccountByIDParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Account{}, fmt.Errorf("mock querier GetAccountByID: %w", err)
	}
	return args.Get(0).(sqlc.Account), nil
}

func (m *Querier) GetActiveOTPByEmail(ctx context.Context, email string) (sqlc.OtpRequest, error) {
	args := m.Called(ctx, email)
	if err := args.Error(1); err != nil {
		return sqlc.OtpRequest{}, fmt.Errorf("mock querier GetActiveOTPByEmail: %w", err)
	}
	return args.Get(0).(sqlc.OtpRequest), nil
}

func (m *Querier) GetCategoryByID(ctx context.Context, arg sqlc.GetCategoryByIDParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Category{}, fmt.Errorf("mock querier GetCategoryByID: %w", err)
	}
	return args.Get(0).(sqlc.Category), nil
}

func (m *Querier) GetTenantByID(ctx context.Context, id string) (sqlc.Tenant, error) {
	args := m.Called(ctx, id)
	if err := args.Error(1); err != nil {
		return sqlc.Tenant{}, fmt.Errorf("mock querier GetTenantByID: %w", err)
	}
	return args.Get(0).(sqlc.Tenant), nil
}

func (m *Querier) GetTransactionByID(ctx context.Context, arg sqlc.GetTransactionByIDParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Transaction{}, fmt.Errorf("mock querier GetTransactionByID: %w", err)
	}
	return args.Get(0).(sqlc.Transaction), nil
}

func (m *Querier) GetUserByEmail(ctx context.Context, arg sqlc.GetUserByEmailParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.User{}, fmt.Errorf("mock querier GetUserByEmail: %w", err)
	}
	return args.Get(0).(sqlc.User), nil
}

func (m *Querier) GetUserByID(ctx context.Context, arg sqlc.GetUserByIDParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.User{}, fmt.Errorf("mock querier GetUserByID: %w", err)
	}
	return args.Get(0).(sqlc.User), nil
}

func (m *Querier) ListAccountsByTenant(ctx context.Context, tenantID string) ([]sqlc.Account, error) {
	args := m.Called(ctx, tenantID)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListAccountsByTenant: %w", err)
	}
	return args.Get(0).([]sqlc.Account), nil
}

func (m *Querier) ListAuditLogsByEntity(ctx context.Context, arg sqlc.ListAuditLogsByEntityParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListAuditLogsByEntity: %w", err)
	}
	return args.Get(0).([]sqlc.AuditLog), nil
}

func (m *Querier) ListAuditLogsByTenant(ctx context.Context, arg sqlc.ListAuditLogsByTenantParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListAuditLogsByTenant: %w", err)
	}
	return args.Get(0).([]sqlc.AuditLog), nil
}

func (m *Querier) ListCategoriesByTenant(ctx context.Context, tenantID string) ([]sqlc.Category, error) {
	args := m.Called(ctx, tenantID)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListCategoriesByTenant: %w", err)
	}
	return args.Get(0).([]sqlc.Category), nil
}

func (m *Querier) ListChildCategories(ctx context.Context, arg sqlc.ListChildCategoriesParams) ([]sqlc.Category, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListChildCategories: %w", err)
	}
	return args.Get(0).([]sqlc.Category), nil
}

func (m *Querier) ListRootCategoriesByTenant(ctx context.Context, tenantID string) ([]sqlc.Category, error) {
	args := m.Called(ctx, tenantID)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListRootCategoriesByTenant: %w", err)
	}
	return args.Get(0).([]sqlc.Category), nil
}

func (m *Querier) ListTenants(ctx context.Context) ([]sqlc.Tenant, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListTenants: %w", err)
	}
	return args.Get(0).([]sqlc.Tenant), nil
}

func (m *Querier) ListTransactionsByTenant(ctx context.Context, arg sqlc.ListTransactionsByTenantParams) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListTransactionsByTenant: %w", err)
	}
	return args.Get(0).([]sqlc.Transaction), nil
}

func (m *Querier) ListUsersByTenant(ctx context.Context, tenantID string) ([]sqlc.User, error) {
	args := m.Called(ctx, tenantID)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock querier ListUsersByTenant: %w", err)
	}
	return args.Get(0).([]sqlc.User), nil
}

func (m *Querier) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier MarkOTPUsed: %w", err)
	}
	return nil
}

func (m *Querier) SoftDeleteAccount(ctx context.Context, arg sqlc.SoftDeleteAccountParams) error {
	args := m.Called(ctx, arg)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier SoftDeleteAccount: %w", err)
	}
	return nil
}

func (m *Querier) SoftDeleteCategory(ctx context.Context, arg sqlc.SoftDeleteCategoryParams) error {
	args := m.Called(ctx, arg)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier SoftDeleteCategory: %w", err)
	}
	return nil
}

func (m *Querier) SoftDeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier SoftDeleteTenant: %w", err)
	}
	return nil
}

func (m *Querier) SoftDeleteTransaction(ctx context.Context, arg sqlc.SoftDeleteTransactionParams) error {
	args := m.Called(ctx, arg)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier SoftDeleteTransaction: %w", err)
	}
	return nil
}

func (m *Querier) SoftDeleteUser(ctx context.Context, arg sqlc.SoftDeleteUserParams) error {
	args := m.Called(ctx, arg)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier SoftDeleteUser: %w", err)
	}
	return nil
}

func (m *Querier) UpdateAccount(ctx context.Context, arg sqlc.UpdateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Account{}, fmt.Errorf("mock querier UpdateAccount: %w", err)
	}
	return args.Get(0).(sqlc.Account), nil
}

func (m *Querier) UpdateAccountBalance(ctx context.Context, arg sqlc.UpdateAccountBalanceParams) error {
	args := m.Called(ctx, arg)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier UpdateAccountBalance: %w", err)
	}
	return nil
}

func (m *Querier) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Category{}, fmt.Errorf("mock querier UpdateCategory: %w", err)
	}
	return args.Get(0).(sqlc.Category), nil
}

func (m *Querier) UpdateTenant(ctx context.Context, arg sqlc.UpdateTenantParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Tenant{}, fmt.Errorf("mock querier UpdateTenant: %w", err)
	}
	return args.Get(0).(sqlc.Tenant), nil
}

func (m *Querier) UpdateTransaction(ctx context.Context, arg sqlc.UpdateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.Transaction{}, fmt.Errorf("mock querier UpdateTransaction: %w", err)
	}
	return args.Get(0).(sqlc.Transaction), nil
}

func (m *Querier) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	if err := args.Error(1); err != nil {
		return sqlc.User{}, fmt.Errorf("mock querier UpdateUser: %w", err)
	}
	return args.Get(0).(sqlc.User), nil
}

func (m *Querier) UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error {
	args := m.Called(ctx, arg)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock querier UpdateUserLastLogin: %w", err)
	}
	return nil
}

var _ sqlc.Querier = (*Querier)(nil)
