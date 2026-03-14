package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/stretchr/testify/mock"
)

// TenantService is a testify/mock implementation of domain.TenantService.
type TenantService struct {
	mock.Mock
}

func (m *TenantService) Create(ctx context.Context, input domain.CreateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantService.Create: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TenantService.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TenantService) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantService.GetByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TenantService.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TenantService) List(ctx context.Context) ([]domain.Tenant, error) {
	args := m.Called(ctx)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantService.List: %w", e)
	}

	res, ok := args.Get(0).([]domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TenantService.List: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TenantService) Update(ctx context.Context, id string, input domain.UpdateTenantInput) (*domain.Tenant, error) {
	args := m.Called(ctx, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantService.Update: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TenantService.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TenantService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock TenantService.Delete: %w", e)
	}
	return nil
}

func (m *TenantService) InviteUser(ctx context.Context, tenantID string, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TenantService.InviteUser: %w", e)
	}

	res, ok := args.Get(0).(*domain.User)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TenantService.InviteUser: unexpected type %T", args.Get(0))
	}
	return res, err
}

var _ domain.TenantService = (*TenantService)(nil)

// AccountService is a testify/mock implementation of domain.AccountService.
type AccountService struct {
	mock.Mock
}

func (m *AccountService) Create(ctx context.Context, tenantID string, input domain.CreateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountService.Create: %w", e)
	}

	res, ok := args.Get(0).(*domain.Account)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AccountService.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountService) GetByID(ctx context.Context, tenantID, id string) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountService.GetByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.Account)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AccountService.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountService.ListByTenant: %w", e)
	}

	res, ok := args.Get(0).([]domain.Account)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AccountService.ListByTenant: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountService) ListByUser(ctx context.Context, tenantID, userID string) ([]domain.Account, error) {
	args := m.Called(ctx, tenantID, userID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountService.ListByUser: %w", e)
	}

	res, ok := args.Get(0).([]domain.Account)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AccountService.ListByUser: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountService) Update(ctx context.Context, tenantID, id string, input domain.UpdateAccountInput) (*domain.Account, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AccountService.Update: %w", e)
	}

	res, ok := args.Get(0).(*domain.Account)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AccountService.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AccountService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AccountService.Delete: %w", e)
	}
	return nil
}

var _ domain.AccountService = (*AccountService)(nil)

// CategoryService is a testify/mock implementation of domain.CategoryService.
type CategoryService struct {
	mock.Mock
}

func (m *CategoryService) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryService.Create: %w", e)
	}

	res, ok := args.Get(0).(*domain.Category)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock CategoryService.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryService) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryService.GetByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.Category)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock CategoryService.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryService.ListByTenant: %w", e)
	}

	res, ok := args.Get(0).([]domain.Category)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock CategoryService.ListByTenant: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryService) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	args := m.Called(ctx, tenantID, parentID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryService.ListChildren: %w", e)
	}

	res, ok := args.Get(0).([]domain.Category)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock CategoryService.ListChildren: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryService) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock CategoryService.Update: %w", e)
	}

	res, ok := args.Get(0).(*domain.Category)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock CategoryService.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *CategoryService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock CategoryService.Delete: %w", e)
	}
	return nil
}

var _ domain.CategoryService = (*CategoryService)(nil)

// TransactionService is a testify/mock implementation of domain.TransactionService.
type TransactionService struct {
	mock.Mock
}

func (m *TransactionService) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionService.Create: %w", e)
	}

	res, ok := args.Get(0).(*domain.Transaction)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TransactionService.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionService) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionService.GetByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.Transaction)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TransactionService.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionService) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	args := m.Called(ctx, tenantID, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionService.List: %w", e)
	}

	res, ok := args.Get(0).([]domain.Transaction)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TransactionService.List: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionService) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock TransactionService.Update: %w", e)
	}

	res, ok := args.Get(0).(*domain.Transaction)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock TransactionService.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *TransactionService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock TransactionService.Delete: %w", e)
	}
	return nil
}

var _ domain.TransactionService = (*TransactionService)(nil)

// AdminService is a testify/mock implementation of domain.AdminService.
type AdminService struct {
	mock.Mock
}

func (m *AdminService) ListAllTenants(ctx context.Context, withDeleted bool) ([]domain.Tenant, error) {
	args := m.Called(ctx, withDeleted)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminService.ListAllTenants: %w", e)
	}

	res, ok := args.Get(0).([]domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminService.ListAllTenants: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminService) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminService.GetTenantByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminService.GetTenantByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminService) UpdateTenantPlan(ctx context.Context, id string, plan domain.TenantPlan) (*domain.Tenant, error) {
	args := m.Called(ctx, id, plan)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminService.UpdateTenantPlan: %w", e)
	}

	res, ok := args.Get(0).(*domain.Tenant)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminService.UpdateTenantPlan: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminService) SuspendTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminService.SuspendTenant: %w", e)
	}
	return nil
}

func (m *AdminService) RestoreTenant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminService.RestoreTenant: %w", e)
	}
	return nil
}

func (m *AdminService) HardDeleteTenant(ctx context.Context, id, confirmationToken string) error {
	args := m.Called(ctx, id, confirmationToken)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminService.HardDeleteTenant: %w", e)
	}
	return nil
}

func (m *AdminService) ListAllUsers(ctx context.Context) ([]domain.User, error) {
	args := m.Called(ctx)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminService.ListAllUsers: %w", e)
	}

	res, ok := args.Get(0).([]domain.User)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminService.ListAllUsers: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminService.GetUserByID: %w", e)
	}

	res, ok := args.Get(0).(*domain.User)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminService.GetUserByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *AdminService) ForceDeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock AdminService.ForceDeleteUser: %w", e)
	}
	return nil
}

func (m *AdminService) ListAuditLogs(ctx context.Context, params domain.ListAuditLogsParams) ([]domain.AuditLog, error) {
	args := m.Called(ctx, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock AdminService.ListAuditLogs: %w", e)
	}

	res, ok := args.Get(0).([]domain.AuditLog)
	if !ok && args.Get(0) != nil {
		return nil, fmt.Errorf("mock AdminService.ListAuditLogs: unexpected type %T", args.Get(0))
	}
	return res, err
}

var _ domain.AdminService = (*AdminService)(nil)

// MasterPurchaseService is a testify/mock implementation of domain.MasterPurchaseService.
type MasterPurchaseService struct {
	mock.Mock
}

func (m *MasterPurchaseService) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseService.Create: %w", e)
	}

	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseService.Create: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseService) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseService.GetByID: %w", e)
	}

	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseService.GetByID: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseService) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseService.ListByTenant: %w", e)
	}

	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).([]domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseService.ListByTenant: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseService) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, accountID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseService.ListByAccount: %w", e)
	}

	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).([]domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseService.ListByAccount: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseService) ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment {
	args := m.Called(mp)

	if args.Get(0) == nil {
		return nil
	}

	res, ok := args.Get(0).([]domain.ProjectedInstallment)
	if !ok {
		return nil
	}
	return res
}

func (m *MasterPurchaseService) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	args := m.Called(ctx, tenantID, id, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock MasterPurchaseService.Update: %w", e)
	}

	if args.Get(0) == nil {
		return nil, err
	}

	res, ok := args.Get(0).(*domain.MasterPurchase)
	if !ok {
		return nil, fmt.Errorf("mock MasterPurchaseService.Update: unexpected type %T", args.Get(0))
	}
	return res, err
}

func (m *MasterPurchaseService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	if e := args.Error(0); e != nil {
		return fmt.Errorf("mock MasterPurchaseService.Delete: %w", e)
	}
	return nil
}

var _ domain.MasterPurchaseService = (*MasterPurchaseService)(nil)

// InvoiceCloser is a testify/mock implementation of domain.InvoiceCloser.
type InvoiceCloser struct {
	mock.Mock
}

func (m *InvoiceCloser) CloseInvoice(ctx context.Context, tenantID, accountID string, closingDate time.Time) (domain.CloseInvoiceResult, error) {
	args := m.Called(ctx, tenantID, accountID, closingDate)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvoiceCloser.CloseInvoice: %w", e)
	}

	res, ok := args.Get(0).(domain.CloseInvoiceResult)
	if !ok {
		return domain.CloseInvoiceResult{}, fmt.Errorf("mock InvoiceCloser.CloseInvoice: unexpected type %T", args.Get(0))
	}
	return res, err
}

var _ domain.InvoiceCloser = (*InvoiceCloser)(nil)

// InvestmentService is a testify/mock implementation of domain.InvestmentService.
type InvestmentService struct {
	mock.Mock
}

func (m *InvestmentService) CreatePosition(ctx context.Context, tenantID string, in domain.CreatePositionInput) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, in)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.CreatePosition: %w", e)
	}
	res, _ := args.Get(0).(*domain.Position) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) GetPosition(ctx context.Context, tenantID, id string) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.GetPosition: %w", e)
	}
	res, _ := args.Get(0).(*domain.Position) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) ListPositions(ctx context.Context, tenantID string) ([]domain.Position, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.ListPositions: %w", e)
	}
	res, _ := args.Get(0).([]domain.Position) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) UpdatePosition(ctx context.Context, tenantID, id string, in domain.UpdatePositionInput) (*domain.Position, error) {
	args := m.Called(ctx, tenantID, id, in)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.UpdatePosition: %w", e)
	}
	res, _ := args.Get(0).(*domain.Position) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) DeletePosition(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0) //nolint:wrapcheck //nolint:wrapcheck
}

func (m *InvestmentService) MarkIncomeReceived(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.MarkIncomeReceived: %w", e)
	}
	res, _ := args.Get(0).(*domain.PositionIncomeEvent) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) CancelIncome(ctx context.Context, tenantID, eventID string) (*domain.PositionIncomeEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.CancelIncome: %w", e)
	}
	res, _ := args.Get(0).(*domain.PositionIncomeEvent) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) GetPortfolioSummary(ctx context.Context, tenantID string) (*domain.PortfolioSummary, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.GetPortfolioSummary: %w", e)
	}
	res, _ := args.Get(0).(*domain.PortfolioSummary) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) TakeSnapshot(ctx context.Context, tenantID string) (*domain.PortfolioSnapshot, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.TakeSnapshot: %w", e)
	}
	res, _ := args.Get(0).(*domain.PortfolioSnapshot) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) CreateAsset(ctx context.Context, input domain.CreateAssetInput) (*domain.Asset, error) {
	args := m.Called(ctx, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.CreateAsset: %w", e)
	}
	res, _ := args.Get(0).(*domain.Asset) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) GetAssetByID(ctx context.Context, id string) (*domain.Asset, error) {
	args := m.Called(ctx, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.GetAssetByID: %w", e)
	}
	res, _ := args.Get(0).(*domain.Asset) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) ListAssets(ctx context.Context, params domain.ListAssetsParams) ([]domain.Asset, error) {
	args := m.Called(ctx, params)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.ListAssets: %w", e)
	}
	res, _ := args.Get(0).([]domain.Asset) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) DeleteAsset(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0) //nolint:wrapcheck //nolint:wrapcheck
}

func (m *InvestmentService) UpsertTenantAssetConfig(ctx context.Context, tenantID string, input domain.UpsertTenantAssetConfigInput) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, input)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.UpsertTenantAssetConfig: %w", e)
	}
	res, _ := args.Get(0).(*domain.TenantAssetConfig) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) GetTenantAssetConfig(ctx context.Context, tenantID, assetID string) (*domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID, assetID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.GetTenantAssetConfig: %w", e)
	}
	res, _ := args.Get(0).(*domain.TenantAssetConfig) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) ListTenantAssetConfigs(ctx context.Context, tenantID string) ([]domain.TenantAssetConfig, error) {
	args := m.Called(ctx, tenantID)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.ListTenantAssetConfigs: %w", e)
	}
	res, _ := args.Get(0).([]domain.TenantAssetConfig) //nolint:errcheck
	return res, err
}

func (m *InvestmentService) DeleteTenantAssetConfig(ctx context.Context, tenantID, assetID string) error {
	args := m.Called(ctx, tenantID, assetID)
	return args.Error(0) //nolint:wrapcheck //nolint:wrapcheck
}

func (m *InvestmentService) GetAssetWithTenantConfig(ctx context.Context, tenantID, id string) (*domain.Asset, error) {
	args := m.Called(ctx, tenantID, id)
	var err error
	if e := args.Error(1); e != nil {
		err = fmt.Errorf("mock InvestmentService.GetAssetWithTenantConfig: %w", e)
	}
	res, _ := args.Get(0).(*domain.Asset) //nolint:errcheck
	return res, err
}

var _ domain.InvestmentService = (*InvestmentService)(nil)
