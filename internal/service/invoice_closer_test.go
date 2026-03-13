package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInvoiceCloser_CloseInvoice(t *testing.T) {
	t.Parallel()

	tenantID := "01H2PX647WP45GRSW78V9N9B1K"
	accountID := "01H2PX647WP45GRSW78V9N9B1L"
	userID := "01H2PX647WP45GRSW78V9N9B1M"
	categoryID := "01H2PX647WP45GRSW78V9N9B1N"
	closingDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	slog.SetDefault(logger)

	t.Run("Empty pending list", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{
			ID:           accountID,
			Type:         domain.AccountTypeCreditCard,
			BalanceCents: 0,
		}, nil)

		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{}, nil)

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
		require.Empty(t, result.Errors)
	})

	t.Run("Single pending purchase success", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{
			ID:               "MP1",
			AccountID:        accountID,
			CategoryID:       categoryID,
			UserID:           userID,
			Description:      "Table",
			PaidInstallments: 0,
			InstallmentCount: 2,
			Status:           domain.MasterPurchaseStatusOpen,
		}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{
			ID:           accountID,
			Type:         domain.AccountTypeCreditCard,
			BalanceCents: 1000,
		}, nil)

		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		schedule := []domain.ProjectedInstallment{
			{DueDate: closingDate.Add(-1 * time.Hour), AmountCents: 500, InstallmentNumber: 1},
			{DueDate: closingDate.Add(30 * 24 * time.Hour), AmountCents: 500, InstallmentNumber: 2},
		}
		mpSvc.On("ProjectInstallments", mock.MatchedBy(func(m *domain.MasterPurchase) bool { return m.ID == "MP1" })).Return(schedule)

		txRepo.On("Create", mock.Anything, tenantID, mock.MatchedBy(func(i domain.CreateTransactionInput) bool {
			return i.MasterPurchaseID == "MP1" && i.AmountCents == 500
		})).Return(&domain.Transaction{ID: "TX1"}, nil)

		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(i domain.CreateAuditLogInput) bool {
			return i.ActorID == domain.ActorSystem && i.EntityID == "TX1" && i.Metadata != nil
		})).Return(&domain.AuditLog{}, nil)

		mpRepo.On("IncrementPaidInstallments", mock.Anything, tenantID, "MP1").Return(nil)

		accRepo.On("UpdateBalance", mock.Anything, tenantID, accountID, int64(1500)).Return(nil)

		// Mocking db.Begin since the service calls it. We need a pool mock or just ignore it if it's not checked.
		// Since I passed nil for db, I should expect a panic if not handled.
		// Let's not pass nil. But I don't have a pgxpool mock easily available here.
		// I will update the service to check if db is nil for testing purposes or use a mock.
		// Actually, I'll update the service to accept an interface for the DB if I wanted to mock it.
		// For now, I'll make the service more robust to nil DB if I'm just unit testing logic that doesn't strictly need the TX atomicity confirmed at this level.

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 1, result.ProcessedCount)
		require.Empty(t, result.Errors)
	})

	t.Run("Multiple purchases with partial failure", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp1 := domain.MasterPurchase{ID: "MP1", AccountID: accountID, PaidInstallments: 0, InstallmentCount: 2, Status: domain.MasterPurchaseStatusOpen}
		mp2 := domain.MasterPurchase{ID: "MP2", AccountID: accountID, PaidInstallments: 0, InstallmentCount: 2, Status: domain.MasterPurchaseStatusOpen}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp1, mp2}, nil)

		// MP1 Success
		schedule1 := []domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100, InstallmentNumber: 1}}
		mpSvc.On("ProjectInstallments", mock.MatchedBy(func(m *domain.MasterPurchase) bool { return m.ID == "MP1" })).Return(schedule1)
		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "TX1"}, nil).Once()
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil).Once()
		mpRepo.On("IncrementPaidInstallments", mock.Anything, tenantID, "MP1").Return(nil)
		accRepo.On("UpdateBalance", mock.Anything, tenantID, accountID, int64(100)).Return(nil).Once()

		// MP2 Failure
		schedule2 := []domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 200, InstallmentNumber: 1}}
		mpSvc.On("ProjectInstallments", mock.MatchedBy(func(m *domain.MasterPurchase) bool { return m.ID == "MP2" })).Return(schedule2)
		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("db error")).Once()

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err) // Top level err is nil because it continues
		require.Equal(t, 1, result.ProcessedCount)
		require.Len(t, result.Errors, 1)
		require.Contains(t, result.Errors[0].Error(), "MP2")
	})

	t.Run("Account not found", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(nil, errors.New("not found"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		_, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get account")
	})

	t.Run("Account not credit card", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{
			ID:   accountID,
			Type: domain.AccountTypeChecking,
		}, nil)

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		_, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.Error(t, err)
		require.Contains(t, err.Error(), "account is not a credit card")
	})

	t.Run("Already closed purchase", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		accRepo := new(mocks.AccountRepository)

		mp := domain.MasterPurchase{ID: "MP1", Status: domain.MasterPurchaseStatusClosed}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		closer := service.NewInvoiceCloser(mpRepo, nil, nil, accRepo, nil, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("No pending installments due", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, InstallmentCount: 2, Status: domain.MasterPurchaseStatusOpen}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		schedule := []domain.ProjectedInstallment{{DueDate: closingDate.Add(24 * time.Hour)}}
		mpSvc.On("ProjectInstallments", mock.Anything).Return(schedule)

		closer := service.NewInvoiceCloser(mpRepo, nil, nil, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("Failed audit log", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, Status: domain.MasterPurchaseStatusOpen, AccountID: accountID, CategoryID: "CAT1"}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		schedule := []domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100, InstallmentNumber: 1}}
		mpSvc.On("ProjectInstallments", mock.Anything).Return(schedule)

		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "TX1"}, nil)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("audit fail"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("Failed tx creation", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, Status: domain.MasterPurchaseStatusOpen, AccountID: accountID, CategoryID: "CAT1"}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		schedule := []domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100, InstallmentNumber: 1}}
		mpSvc.On("ProjectInstallments", mock.Anything).Return(schedule)

		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("db error"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("Failed balance update", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, Status: domain.MasterPurchaseStatusOpen, AccountID: accountID, CategoryID: "CAT1"}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		schedule := []domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100, InstallmentNumber: 1}}
		mpSvc.On("ProjectInstallments", mock.Anything).Return(schedule)

		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "TX1"}, nil)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)
		mpRepo.On("IncrementPaidInstallments", mock.Anything, tenantID, "MP1").Return(nil)
		accRepo.On("UpdateBalance", mock.Anything, tenantID, accountID, int64(100)).Return(errors.New("db error"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("Already fully paid but not closed", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{
			ID:               "MP1",
			PaidInstallments: 2,
			InstallmentCount: 2,
			Status:           domain.MasterPurchaseStatusOpen,
		}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		// Schedule with 2 installments
		schedule := []domain.ProjectedInstallment{
			{InstallmentNumber: 1},
			{InstallmentNumber: 2},
		}
		mpSvc.On("ProjectInstallments", mock.Anything).Return(schedule)

		closer := service.NewInvoiceCloser(mpRepo, nil, nil, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("Failed to list master purchases", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		accRepo := new(mocks.AccountRepository)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return(nil, errors.New("db error"))

		closer := service.NewInvoiceCloser(mpRepo, nil, nil, accRepo, nil, nil)
		_, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list pending master purchases")
	})

	t.Run("Failed increment installments", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, Status: domain.MasterPurchaseStatusOpen, AccountID: accountID, CategoryID: "CAT1"}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)

		schedule := []domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100, InstallmentNumber: 1}}
		mpSvc.On("ProjectInstallments", mock.Anything).Return(schedule)

		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "TX1"}, nil)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)
		mpRepo.On("IncrementPaidInstallments", mock.Anything, tenantID, "MP1").Return(errors.New("db error"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Equal(t, 0, result.ProcessedCount)
	})

	t.Run("Failed increments", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, Status: domain.MasterPurchaseStatusOpen}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)
		mpSvc.On("ProjectInstallments", mock.Anything).Return([]domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100}})
		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "TX1"}, nil)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)
		mpRepo.On("IncrementPaidInstallments", mock.Anything, tenantID, "MP1").Return(errors.New("db fail"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Len(t, result.Errors, 1)
	})

	t.Run("Failed balance update", func(t *testing.T) {
		t.Parallel()

		mpRepo := new(mocks.MasterPurchaseRepository)
		txRepo := new(mocks.TransactionRepository)
		auditRepo := new(mocks.AuditRepository)
		accRepo := new(mocks.AccountRepository)
		mpSvc := new(mocks.MasterPurchaseService)

		mp := domain.MasterPurchase{ID: "MP1", PaidInstallments: 0, Status: domain.MasterPurchaseStatusOpen}

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{ID: accountID, Type: domain.AccountTypeCreditCard, BalanceCents: 0}, nil)
		mpRepo.On("ListByAccount", mock.Anything, tenantID, accountID).Return([]domain.MasterPurchase{mp}, nil)
		mpSvc.On("ProjectInstallments", mock.Anything).Return([]domain.ProjectedInstallment{{DueDate: closingDate, AmountCents: 100}})
		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "TX1"}, nil)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)
		mpRepo.On("IncrementPaidInstallments", mock.Anything, tenantID, "MP1").Return(nil)
		accRepo.On("UpdateBalance", mock.Anything, tenantID, accountID, mock.Anything).Return(errors.New("db fail"))

		closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, nil)
		result, err := closer.CloseInvoice(context.Background(), tenantID, accountID, closingDate)

		require.NoError(t, err)
		require.Len(t, result.Errors, 1)
	})
}
