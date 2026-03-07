package mocks

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/stretchr/testify/mock"
)

type Querier struct {
	mock.Mock
}

func (m *Querier) AdminForceDeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) AdminGetTenantByID(ctx context.Context, id string) (sqlc.Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) AdminGetUserByID(ctx context.Context, id string) (sqlc.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) AdminHardDeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) AdminListAllAuditLogs(ctx context.Context, arg sqlc.AdminListAllAuditLogsParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]sqlc.AuditLog), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) AdminListAllTenants(ctx context.Context, withDeleted bool) ([]sqlc.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	return args.Get(0).([]sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) AdminListAllUsers(ctx context.Context) ([]sqlc.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) AdminRestoreTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) AdminSuspendTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) AdminUpdateTenantPlan(ctx context.Context, arg sqlc.AdminUpdateTenantPlanParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateAccount(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Account), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateAuditLog(ctx context.Context, arg sqlc.CreateAuditLogParams) (sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.AuditLog), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateCategory(ctx context.Context, arg sqlc.CreateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Category), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateOTPRequest(ctx context.Context, arg sqlc.CreateOTPRequestParams) (sqlc.OtpRequest, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.OtpRequest), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateTenant(ctx context.Context, arg sqlc.CreateTenantParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateTransaction(ctx context.Context, arg sqlc.CreateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Transaction), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) DeleteExpiredOTPs(ctx context.Context) error {
	args := m.Called(ctx)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) GetAccountByID(ctx context.Context, arg sqlc.GetAccountByIDParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Account), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) GetActiveOTPByEmail(ctx context.Context, email string) (sqlc.OtpRequest, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(sqlc.OtpRequest), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) GetCategoryByID(ctx context.Context, arg sqlc.GetCategoryByIDParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Category), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) GetTenantByID(ctx context.Context, id string) (sqlc.Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) GetTransactionByID(ctx context.Context, arg sqlc.GetTransactionByIDParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Transaction), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) GetUserByEmail(ctx context.Context, arg sqlc.GetUserByEmailParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) GetUserByID(ctx context.Context, arg sqlc.GetUserByIDParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListAccountsByTenant(ctx context.Context, tenantID string) ([]sqlc.Account, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]sqlc.Account), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListAccountsByUser(ctx context.Context, arg sqlc.ListAccountsByUserParams) ([]sqlc.Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]sqlc.Account), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListAuditLogsByEntity(ctx context.Context, arg sqlc.ListAuditLogsByEntityParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]sqlc.AuditLog), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListAuditLogsByTenant(ctx context.Context, arg sqlc.ListAuditLogsByTenantParams) ([]sqlc.AuditLog, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]sqlc.AuditLog), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListCategoriesByTenant(ctx context.Context, tenantID string) ([]sqlc.Category, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]sqlc.Category), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListChildCategories(ctx context.Context, arg sqlc.ListChildCategoriesParams) ([]sqlc.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]sqlc.Category), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListTenants(ctx context.Context) ([]sqlc.Tenant, error) {
	args := m.Called(ctx)
	return args.Get(0).([]sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListTransactionsByTenant(ctx context.Context, tenantID string) ([]sqlc.Transaction, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]sqlc.Transaction), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) ListUsersByTenant(ctx context.Context, tenantID string) ([]sqlc.User, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) MarkOTPUsed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) SoftDeleteAccount(ctx context.Context, arg sqlc.SoftDeleteAccountParams) error {
	args := m.Called(ctx, arg)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) SoftDeleteCategory(ctx context.Context, arg sqlc.SoftDeleteCategoryParams) error {
	args := m.Called(ctx, arg)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) SoftDeleteTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) SoftDeleteTransaction(ctx context.Context, arg sqlc.SoftDeleteTransactionParams) error {
	args := m.Called(ctx, arg)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) SoftDeleteUser(ctx context.Context, arg sqlc.SoftDeleteUserParams) error {
	args := m.Called(ctx, arg)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) UpdateAccount(ctx context.Context, arg sqlc.UpdateAccountParams) (sqlc.Account, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Account), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) UpdateAccountBalance(ctx context.Context, arg sqlc.UpdateAccountBalanceParams) error {
	args := m.Called(ctx, arg)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}

func (m *Querier) UpdateCategory(ctx context.Context, arg sqlc.UpdateCategoryParams) (sqlc.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Category), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) UpdateTenant(ctx context.Context, arg sqlc.UpdateTenantParams) (sqlc.Tenant, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Tenant), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) UpdateTransaction(ctx context.Context, arg sqlc.UpdateTransactionParams) (sqlc.Transaction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.Transaction), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(sqlc.User), fmt.Errorf("mock error: %w", args.Error(1))
}

func (m *Querier) UpdateUserLastLogin(ctx context.Context, arg sqlc.UpdateUserLastLoginParams) error {
	args := m.Called(ctx, arg)
	return fmt.Errorf("mock error: %w", fmt.Errorf("mock error: %w", args.Error(0)))
}
