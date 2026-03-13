package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	userID := "user_1"
	input := domain.CreateAccountInput{
		UserID:       userID,
		Name:         "Main Checking",
		Type:         domain.AccountTypeChecking,
		Currency:     "USD",
		InitialCents: 1000,
	}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		user := &domain.User{ID: userID, TenantID: tenantID, Role: domain.RoleAdmin}
		account := &domain.Account{
			ID:           "acc_1",
			TenantID:     tenantID,
			UserID:       userID,
			Name:         input.Name,
			Type:         input.Type,
			Currency:     input.Currency,
			BalanceCents: input.InitialCents,
		}

		userRepo.On("GetByID", ctx, tenantID, userID).Return(user, nil)
		accountRepo.On("Create", ctx, tenantID, input).Return(account, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewAccountService(accountRepo, userRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
		require.Equal(t, account, res)
	})

	t.Run("AuditErrorCreate", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		user := &domain.User{ID: userID, TenantID: tenantID, Role: domain.RoleAdmin}
		account := &domain.Account{ID: "acc_1", UserID: userID}

		userRepo.On("GetByID", ctx, tenantID, userID).Return(user, nil)
		accountRepo.On("Create", ctx, tenantID, input).Return(account, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewAccountService(accountRepo, userRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
		require.Equal(t, account, res)
	})

	t.Run("UserLookupError", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.UserRepository)
		userRepo.On("GetByID", ctx, tenantID, userID).Return(nil, errors.New("db error"))

		svc := service.NewAccountService(nil, userRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("UserNil", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.UserRepository)
		userRepo.On("GetByID", ctx, tenantID, userID).Return(nil, nil)

		svc := service.NewAccountService(nil, userRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		userRepo.On("GetByID", ctx, tenantID, userID).Return(nil, domain.ErrNotFound)

		svc := service.NewAccountService(accountRepo, userRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("RepoError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		user := &domain.User{ID: userID, TenantID: tenantID}
		userRepo.On("GetByID", ctx, tenantID, userID).Return(user, nil)
		accountRepo.On("Create", ctx, tenantID, input).Return(nil, errors.New("db error"))

		svc := service.NewAccountService(accountRepo, userRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestAccountService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	accountID := "acc_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		account := &domain.Account{ID: accountID, TenantID: tenantID}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(account, nil)

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.GetByID(ctx, tenantID, accountID)

		require.NoError(t, err)
		require.Equal(t, account, res)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, domain.ErrNotFound)

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.GetByID(ctx, tenantID, accountID)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, res)
	})
}

func TestAccountService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	userID := "user_1"

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accounts := []domain.Account{{ID: "acc_1", TenantID: tenantID}}

		accountRepo.On("ListByTenant", ctx, tenantID).Return(accounts, nil)

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.ListByTenant(ctx, tenantID)

		require.NoError(t, err)
		require.Equal(t, accounts, res)
	})

	t.Run("ListByTenant_Error", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("ListByTenant", ctx, tenantID).Return(([]domain.Account)(nil), errors.New("db error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.ListByTenant(ctx, tenantID)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("ListByUser", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accounts := []domain.Account{{ID: "acc_1", TenantID: tenantID, UserID: userID}}

		accountRepo.On("ListByUser", ctx, tenantID, userID).Return(accounts, nil)

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.ListByUser(ctx, tenantID, userID)

		require.NoError(t, err)
		require.Equal(t, accounts, res)
	})

	t.Run("ListByUser_Error", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("ListByUser", ctx, tenantID, userID).Return(([]domain.Account)(nil), errors.New("db error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.ListByUser(ctx, tenantID, userID)

		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestAccountService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	accountID := "acc_1"
	newName := "New Account Name"
	input := domain.UpdateAccountInput{Name: &newName}

	t.Run("LookupError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, errors.New("db error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, accountID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("UpdateError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{Name: "Old"}, nil)
		accountRepo.On("Update", ctx, tenantID, accountID, input).Return(nil, errors.New("update error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, accountID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		oldAccount := &domain.Account{ID: accountID, TenantID: tenantID, Name: "Old Name"}
		newAccount := &domain.Account{ID: accountID, TenantID: tenantID, Name: newName, UserID: "user_1"}
		newCurrency := "USD"
		inputWithBoth := domain.UpdateAccountInput{Name: &newName, Currency: &newCurrency}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(oldAccount, nil)
		accountRepo.On("Update", ctx, tenantID, accountID, inputWithBoth).Return(newAccount, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.Action == domain.AuditActionUpdate && input.EntityID == accountID
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAccountService(accountRepo, nil, auditRepo)
		res, err := svc.Update(ctx, tenantID, accountID, inputWithBoth)

		require.NoError(t, err)
		require.Equal(t, newAccount, res)
	})

	t.Run("AuditErrorCleanup", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		oldAccount := &domain.Account{ID: accountID, TenantID: tenantID, Name: "Old Name"}
		newAccount := &domain.Account{ID: accountID, TenantID: tenantID, Name: newName, UserID: "user_1"}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(oldAccount, nil)
		accountRepo.On("Update", ctx, tenantID, accountID, input).Return(newAccount, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewAccountService(accountRepo, nil, auditRepo)
		res, err := svc.Update(ctx, tenantID, accountID, input)

		require.NoError(t, err)
		require.Equal(t, newAccount, res)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, domain.ErrNotFound)

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, accountID, domain.UpdateAccountInput{})

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, res)
	})

	t.Run("UpdateError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		oldAccount := &domain.Account{ID: accountID, TenantID: tenantID, Name: "Old Name"}
		newName := "New Name"
		input := domain.UpdateAccountInput{Name: &newName}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(oldAccount, nil)
		accountRepo.On("Update", ctx, tenantID, accountID, input).Return((*domain.Account)(nil), errors.New("db error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, accountID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("LookupError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, errors.New("lookup error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, accountID, domain.UpdateAccountInput{})

		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestAccountService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	accountID := "acc_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		account := &domain.Account{ID: accountID, TenantID: tenantID}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(account, nil)
		accountRepo.On("Delete", ctx, tenantID, accountID).Return(nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.Action == domain.AuditActionSoftDelete && input.EntityID == accountID
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAccountService(accountRepo, nil, auditRepo)
		err := svc.Delete(ctx, tenantID, accountID)

		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, domain.ErrNotFound)

		svc := service.NewAccountService(accountRepo, nil, nil)
		err := svc.Delete(ctx, tenantID, accountID)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("DeleteError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID, UserID: "user_1"}, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)
		accountRepo.On("Delete", ctx, tenantID, accountID).Return(errors.New("delete error"))

		svc := service.NewAccountService(accountRepo, nil, auditRepo)
		err := svc.Delete(ctx, tenantID, accountID)

		require.Error(t, err)
	})

	t.Run("LookupError", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, errors.New("lookup error"))

		svc := service.NewAccountService(accountRepo, nil, nil)
		err := svc.Delete(ctx, tenantID, accountID)

		require.Error(t, err)
	})

	t.Run("AuditError_Delete", func(t *testing.T) {
		t.Parallel()

		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		accountRepo.On("Delete", ctx, tenantID, accountID).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewAccountService(accountRepo, nil, auditRepo)
		err := svc.Delete(ctx, tenantID, accountID)

		require.NoError(t, err) // We don't fail the whole request if audit fails
	})
}
