package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	accountID := "acc_1"
	categoryID := "cat_1"
	userID := "user_1"
	input := domain.CreateTransactionInput{
		OccurredAt:  time.Now(),
		AccountID:   accountID,
		CategoryID:  categoryID,
		UserID:      userID,
		Description: "Lunch",
		Type:        domain.TransactionTypeExpense,
		AmountCents: 1500,
	}

	t.Run("Success_Expense", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		account := &domain.Account{ID: accountID, TenantID: tenantID}
		category := &domain.Category{ID: categoryID, TenantID: tenantID, Type: domain.CategoryTypeExpense}
		tx := &domain.Transaction{ID: "tx_1", AccountID: accountID, AmountCents: 1500, Type: domain.TransactionTypeExpense}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(account, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(category, nil)
		txRepo.On("Create", ctx, tenantID, input).Return(tx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, accountID, int64(-1500)).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, accountRepo, categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
		require.Equal(t, tx, res)
	})

	t.Run("Success_Income", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		incomeInput := input
		incomeInput.Type = domain.TransactionTypeIncome
		incomeInput.Description = "Salary"

		account := &domain.Account{ID: accountID}
		category := &domain.Category{ID: categoryID, Type: domain.CategoryTypeIncome}
		tx := &domain.Transaction{ID: "tx_2", Type: domain.TransactionTypeIncome, AmountCents: 1500}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(account, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(category, nil)
		txRepo.On("Create", ctx, tenantID, incomeInput).Return(tx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, accountID, int64(1500)).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, accountRepo, categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, incomeInput)

		require.NoError(t, err)
		require.Equal(t, tx, res)
	})

	t.Run("Error_CategoryTypeMismatch", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)

		account := &domain.Account{ID: accountID}
		category := &domain.Category{ID: categoryID, Type: domain.CategoryTypeIncome}

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(account, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(category, nil)

		svc := service.NewTransactionService(nil, accountRepo, categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		require.Nil(t, res)
	})

	t.Run("Error_AccountLookup", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, errors.New("db error"))

		svc := service.NewTransactionService(nil, accountRepo, nil, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to verify account")
		require.Nil(t, res)
	})

	t.Run("Error_AccountNotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, nil)

		svc := service.NewTransactionService(nil, accountRepo, nil, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Contains(t, err.Error(), "account not found")
		require.Nil(t, res)
	})

	t.Run("Error_CategoryNotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(nil, nil)

		svc := service.NewTransactionService(nil, accountRepo, categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Contains(t, err.Error(), "category not found")
		require.Nil(t, res)
	})

	t.Run("Error_CategoryLookup", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return((*domain.Category)(nil), errors.New("db error"))

		svc := service.NewTransactionService(nil, accountRepo, categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Error_RepoCreate", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(&domain.Category{ID: categoryID, Type: domain.CategoryTypeExpense}, nil)
		txRepo.On("Create", ctx, tenantID, input).Return(nil, errors.New("db error"))

		svc := service.NewTransactionService(txRepo, accountRepo, categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create transaction record")
		require.Nil(t, res)
	})

	t.Run("Error_BalanceUpdate", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(&domain.Category{ID: categoryID, Type: domain.CategoryTypeExpense}, nil)
		txRepo.On("Create", ctx, tenantID, input).Return(&domain.Transaction{ID: "tx_1", AccountID: accountID, CategoryID: categoryID, Type: input.Type, AmountCents: input.AmountCents}, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, accountID, int64(-1500)).Return(errors.New("balance error"))

		svc := service.NewTransactionService(txRepo, accountRepo, categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("AuditError_Create", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(&domain.Category{ID: categoryID, Type: domain.CategoryTypeExpense}, nil)
		txRepo.On("Create", ctx, tenantID, input).Return(&domain.Transaction{ID: "tx_1", AccountID: accountID, CategoryID: categoryID, Type: input.Type, AmountCents: input.AmountCents}, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, accountID, int64(-1500)).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewTransactionService(txRepo, accountRepo, categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err) // Audit failure shouldn't fail the primary action in current implementation
		require.NotNil(t, res)
	})
}

func TestTransactionService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	txID := "tx_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		expected := &domain.Transaction{ID: txID}
		txRepo.On("GetByID", ctx, tenantID, txID).Return(expected, nil)

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		res, err := svc.GetByID(ctx, tenantID, txID)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		txRepo.On("GetByID", ctx, tenantID, txID).Return((*domain.Transaction)(nil), errors.New("db error"))

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		res, err := svc.GetByID(ctx, tenantID, txID)

		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestTransactionService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	params := domain.ListTransactionsParams{Limit: 10}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		expected := []domain.Transaction{{ID: "tx_1"}}
		txRepo.On("List", ctx, tenantID, params).Return(expected, nil)

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		res, err := svc.List(ctx, tenantID, params)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		txRepo.On("List", ctx, tenantID, params).Return(([]domain.Transaction)(nil), errors.New("db error"))

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		res, err := svc.List(ctx, tenantID, params)

		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestTransactionService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	txID := "tx_1"
	newAmount := int64(2000)
	input := domain.UpdateTransactionInput{AmountCents: &newAmount}

	t.Run("Success_AmountChanged", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		oldTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}
		newTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 2000, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(oldTx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(-500)).Return(nil)
		txRepo.On("Update", ctx, tenantID, txID, input).Return(newTx, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, accountRepo, nil, auditRepo)
		res, err := svc.Update(ctx, tenantID, txID, input)

		require.NoError(t, err)
		require.Equal(t, newTx, res)
	})

	t.Run("Success_AmountNotChanged", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		desc := "New Description"
		inputNoAmount := domain.UpdateTransactionInput{Description: &desc}

		oldTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}
		newTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense, Description: desc}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(oldTx, nil)
		txRepo.On("Update", ctx, tenantID, txID, inputNoAmount).Return(newTx, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, nil, nil, auditRepo)
		res, err := svc.Update(ctx, tenantID, txID, inputNoAmount)

		require.NoError(t, err)
		require.Equal(t, newTx, res)
	})

	t.Run("Success_CategoryIDChanged", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		newCatID := "cat_2"
		inputCat := domain.UpdateTransactionInput{CategoryID: &newCatID}

		oldTx := &domain.Transaction{ID: txID, AccountID: "acc_1", CategoryID: "cat_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}
		newTx := &domain.Transaction{ID: txID, AccountID: "acc_1", CategoryID: "cat_2", AmountCents: 1500, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(oldTx, nil)
		txRepo.On("Update", ctx, tenantID, txID, inputCat).Return(newTx, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, nil, nil, auditRepo)
		res, err := svc.Update(ctx, tenantID, txID, inputCat)

		require.NoError(t, err)
		require.Equal(t, newTx, res)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		txRepo.On("GetByID", ctx, tenantID, txID).Return((*domain.Transaction)(nil), nil)

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		res, err := svc.Update(ctx, tenantID, txID, input)

		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, res)
	})

	t.Run("RepoError_Lookup", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		txRepo.On("GetByID", ctx, tenantID, txID).Return((*domain.Transaction)(nil), errors.New("db error"))

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		res, err := svc.Update(ctx, tenantID, txID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("RepoError_Update", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)

		txRepo.On("GetByID", ctx, tenantID, txID).Return(&domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(-500)).Return(nil)
		txRepo.On("Update", ctx, tenantID, txID, input).Return((*domain.Transaction)(nil), errors.New("update fail"))

		svc := service.NewTransactionService(txRepo, accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, txID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("BalanceError_Update", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)

		oldTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(oldTx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(-500)).Return(errors.New("balance fail"))

		svc := service.NewTransactionService(txRepo, accountRepo, nil, nil)
		res, err := svc.Update(ctx, tenantID, txID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("AuditError_Update", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accountRepo := new(mocks.AccountRepository)

		oldTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}
		newTx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 2000, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(oldTx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(-500)).Return(nil)
		txRepo.On("Update", ctx, tenantID, txID, input).Return(newTx, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewTransactionService(txRepo, accountRepo, nil, auditRepo)
		res, err := svc.Update(ctx, tenantID, txID, input)

		require.NoError(t, err)
		require.Equal(t, newTx, res)
	})
}

func TestTransactionService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	txID := "tx_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		tx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(tx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(1500)).Return(nil)
		txRepo.On("Delete", ctx, tenantID, txID).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, accountRepo, nil, auditRepo)
		err := svc.Delete(ctx, tenantID, txID)

		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		txRepo.On("GetByID", ctx, tenantID, txID).Return((*domain.Transaction)(nil), nil)

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		err := svc.Delete(ctx, tenantID, txID)

		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("DeleteRepoError", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		tx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(tx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(1500)).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)
		txRepo.On("Delete", ctx, tenantID, txID).Return(errors.New("db error"))

		svc := service.NewTransactionService(txRepo, accountRepo, nil, auditRepo)
		err := svc.Delete(ctx, tenantID, txID)

		require.Error(t, err)
	})

	t.Run("RepoError_Lookup", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		txRepo.On("GetByID", ctx, tenantID, txID).Return((*domain.Transaction)(nil), errors.New("db error"))

		svc := service.NewTransactionService(txRepo, nil, nil, nil)
		err := svc.Delete(ctx, tenantID, txID)

		require.Error(t, err)
	})

	t.Run("BalanceError_Delete", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)

		tx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(tx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(1500)).Return(errors.New("balance fail"))

		svc := service.NewTransactionService(txRepo, accountRepo, nil, nil)
		err := svc.Delete(ctx, tenantID, txID)

		require.Error(t, err)
	})

	t.Run("AuditError_Delete", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		auditRepo := new(mocks.AuditRepository)

		tx := &domain.Transaction{ID: txID, AccountID: "acc_1", AmountCents: 1500, Type: domain.TransactionTypeExpense}

		txRepo.On("GetByID", ctx, tenantID, txID).Return(tx, nil)
		accountRepo.On("UpdateBalance", ctx, tenantID, "acc_1", int64(1500)).Return(nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))
		txRepo.On("Delete", ctx, tenantID, txID).Return(nil)

		svc := service.NewTransactionService(txRepo, accountRepo, nil, auditRepo)
		err := svc.Delete(ctx, tenantID, txID)

		require.NoError(t, err)
	})
}

func TestTransactionService_CalculateDelta_Coverage(t *testing.T) {
	t.Parallel()

	// Since calculateDelta is internal, we trigger the default case by providing an invalid TransactionType
	// via a Create call (it checks type later when calculating delta).
	ctx := context.Background()
	tenantID := "tenant_1"
	accountID := "acc_1"
	categoryID := "cat_1"
	input := domain.CreateTransactionInput{
		AccountID:   accountID,
		CategoryID:  categoryID,
		Type:        domain.TransactionType("INVALID"),
		AmountCents: 1000,
	}

	t.Run("DefaultCase", func(t *testing.T) {
		t.Parallel()
		txRepo := new(mocks.TransactionRepository)
		accountRepo := new(mocks.AccountRepository)
		categoryRepo := new(mocks.CategoryRepository)

		// Mock account/category lookups to pass validation
		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(&domain.Account{ID: accountID}, nil)
		categoryRepo.On("GetByID", ctx, tenantID, categoryID).Return(&domain.Category{ID: categoryID, Type: domain.CategoryType("INVALID")}, nil)

		// Create needs to be called to reach calculateDelta
		txRepo.On("Create", ctx, tenantID, input).Return(&domain.Transaction{ID: "tx_1", Type: input.Type, AmountCents: input.AmountCents}, nil)
		// delta for INVALID should be 0
		accountRepo.On("UpdateBalance", ctx, tenantID, accountID, int64(0)).Return(nil)

		auditRepo := new(mocks.AuditRepository)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewTransactionService(txRepo, accountRepo, categoryRepo, auditRepo)
		_, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
	})
}
