package service_test

import (
	"context"
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

	t.Run("Error_AccountNotFound", func(t *testing.T) {
		t.Parallel()
		accountRepo := new(mocks.AccountRepository)

		accountRepo.On("GetByID", ctx, tenantID, accountID).Return(nil, domain.ErrNotFound)

		svc := service.NewTransactionService(nil, accountRepo, nil, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
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
}
