package mocks

import (
	"context"

	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/stretchr/testify/mock"
)

type Querier struct {
	mock.Mock
}

func (m *Querier) AdminForceDeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) AdminGetTenantByID(ctx context.Context, id string) (sqlc.Tenant, error) {
	args := m.Called(ctx, id)
	var r0 sqlc.Tenant
	if rf, ok := args.Get(0).(sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) AdminGetUserByID(ctx context.Context, id string) (sqlc.User, error) {
	args := m.Called(ctx, id)
	var r0 sqlc.User
	if rf, ok := args.Get(0).(sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) AdminHardDeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) AdminListAllAuditLogs(ctx context.Context, arg sqlc.AdminListAllAuditLogsParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	var r0 []sqlc.AuditLog
	if rf, ok := args.Get(0).([]sqlc.AuditLog); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) AdminListAllTenants(ctx context.Context, withDeleted bool) ([]sqlc.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	var r0 []sqlc.Tenant
	if rf, ok := args.Get(0).([]sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) AdminListAllUsers(ctx context.Context) ([]sqlc.User, error) {
	args := m.Called(ctx)
	var r0 []sqlc.User
	if rf, ok := args.Get(0).([]sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) AdminRestoreTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) AdminSuspendTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) AdminUpdateTenantPlan(ctx context.Context, arg sqlc.AdminUpdateTenantPlanParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Tenant
	if rf, ok := args.Get(0).(sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateAccount(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Account
	if rf, ok := args.Get(0).(sqlc.Account); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateAuditLog(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.AuditLog
	if rf, ok := args.Get(0).(sqlc.AuditLog); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Category
	if rf, ok := args.Get(0).(sqlc.Category); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateOTPRequest(ctx context.Context, arg sqlc.CreateOTPRequestParams) (sqlc.OtpRequest, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.OtpRequest
	if rf, ok := args.Get(0).(sqlc.OtpRequest); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateTenant(ctx context.Context, arg sqlc.CreateTenantParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Tenant
	if rf, ok := args.Get(0).(sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateTransaction(ctx context.Context, arg sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Transaction
	if rf, ok := args.Get(0).(sqlc.Transaction); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.User
	if rf, ok := args.Get(0).(sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) DeleteExpiredOTPs(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) GetAccountByID(ctx context.Context, arg sqlc.GetAccountByIDParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Account
	if rf, ok := args.Get(0).(sqlc.Account); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) GetActiveOTPByEmail(ctx context.Context, email string) (sqlc.OtpRequest, error) {
	args := m.Called(ctx, email)
	var r0 sqlc.OtpRequest
	if rf, ok := args.Get(0).(sqlc.OtpRequest); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) GetCategoryByID(ctx context.Context, arg sqlc.GetCategoryByIDParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Category
	if rf, ok := args.Get(0).(sqlc.Category); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) GetTenantByID(ctx context.Context, id string) (sqlc.Tenant, error) {
	args := m.Called(ctx, id)
	var r0 sqlc.Tenant
	if rf, ok := args.Get(0).(sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) GetTransactionByID(ctx context.Context, arg sqlc.GetTransactionByIDParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Transaction
	if rf, ok := args.Get(0).(sqlc.Transaction); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) GetUserByEmail(ctx context.Context, arg sqlc.GetUserByEmailParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.User
	if rf, ok := args.Get(0).(sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) GetUserByID(ctx context.Context, arg sqlc.GetUserByIDParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.User
	if rf, ok := args.Get(0).(sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListAccountsByTenant(ctx context.Context, tenantID string) ([]sqlc.Account, error) {
	args := m.Called(ctx, tenantID)
	var r0 []sqlc.Account
	if rf, ok := args.Get(0).([]sqlc.Account); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListAccountsByUser(ctx context.Context, arg sqlc.ListAccountsByUserParams) ([]sqlc.Account, error) {
	args := m.Called(ctx, arg)
	var r0 []sqlc.Account
	if rf, ok := args.Get(0).([]sqlc.Account); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListAuditLogsByEntity(ctx context.Context, arg sqlc.ListAuditLogsByEntityParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	var r0 []sqlc.AuditLog
	if rf, ok := args.Get(0).([]sqlc.AuditLog); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListAuditLogsByTenant(ctx context.Context, arg sqlc.ListAuditLogsByTenantParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	var r0 []sqlc.AuditLog
	if rf, ok := args.Get(0).([]sqlc.AuditLog); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListCategoriesByTenant(ctx context.Context, tenantID string) ([]sqlc.Category, error) {
	args := m.Called(ctx, tenantID)
	var r0 []sqlc.Category
	if rf, ok := args.Get(0).([]sqlc.Category); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListChildCategories(ctx context.Context, arg sqlc.ListChildCategoriesParams) ([]sqlc.Category, error) {
	args := m.Called(ctx, arg)
	var r0 []sqlc.Category
	if rf, ok := args.Get(0).([]sqlc.Category); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListTenants(ctx context.Context) ([]sqlc.Tenant, error) {
	args := m.Called(ctx)
	var r0 []sqlc.Tenant
	if rf, ok := args.Get(0).([]sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListTransactionsByTenant(ctx context.Context, arg sqlc.ListTransactionsByTenantParams) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	var r0 []sqlc.Transaction
	if rf, ok := args.Get(0).([]sqlc.Transaction); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) ListUsersByTenant(ctx context.Context, tenantID string) ([]sqlc.User, error) {
	args := m.Called(ctx, tenantID)
	var r0 []sqlc.User
	if rf, ok := args.Get(0).([]sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) SoftDeleteAccount(ctx context.Context, arg sqlc.SoftDeleteAccountParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) SoftDeleteCategory(ctx context.Context, arg sqlc.SoftDeleteCategoryParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) SoftDeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) SoftDeleteTransaction(ctx context.Context, arg sqlc.SoftDeleteTransactionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) SoftDeleteUser(ctx context.Context, arg sqlc.SoftDeleteUserParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) UpdateAccount(ctx context.Context, arg sqlc.UpdateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Account
	if rf, ok := args.Get(0).(sqlc.Account); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) UpdateAccountBalance(ctx context.Context, arg sqlc.UpdateAccountBalanceParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0) //nolint:wrapcheck
}

func (m *Querier) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Category
	if rf, ok := args.Get(0).(sqlc.Category); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) UpdateTenant(ctx context.Context, arg sqlc.UpdateTenantParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Tenant
	if rf, ok := args.Get(0).(sqlc.Tenant); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) UpdateTransaction(ctx context.Context, arg sqlc.UpdateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.Transaction
	if rf, ok := args.Get(0).(sqlc.Transaction); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	var r0 sqlc.User
	if rf, ok := args.Get(0).(sqlc.User); ok {
		r0 = rf
	}
	return r0, args.Error(1) //nolint:wrapcheck
}

func (m *Querier) UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0) //nolint:wrapcheck
}
