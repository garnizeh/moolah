package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupInvestmentService(t *testing.T) (
	domain.InvestmentService,
	*mocks.PositionRepository,
	*mocks.PositionIncomeEventRepository,
	*mocks.AssetRepository,
	*mocks.AccountRepository,
	*mocks.TransactionRepository,
	*mocks.AuditRepository,
) {
	t.Helper()

	posRepo := &mocks.PositionRepository{}
	incRepo := &mocks.PositionIncomeEventRepository{}
	assetRepo := &mocks.AssetRepository{}
	tenantCfgRepo := &mocks.TenantAssetConfigRepository{}
	accRepo := &mocks.AccountRepository{}
	txRepo := &mocks.TransactionRepository{}
	auditRepo := &mocks.AuditRepository{}
	converter := &mocks.CurrencyConverter{}

	svc := NewInvestmentService(
		posRepo, incRepo, assetRepo, tenantCfgRepo, accRepo, txRepo, auditRepo, converter,
	)

	return svc, posRepo, incRepo, assetRepo, accRepo, txRepo, auditRepo
}

func TestInvestmentService_CreatePosition_Validation(t *testing.T) {
	t.Parallel()

	tenantID := "01J7K7V8N6Y0Q1Z2X3C4V5B6N7"
	accountID := "01J7K7V8N6Y0Q1Z2X3C4V5B6N8"
	assetID := "01J7K7V8N6Y0Q1Z2X3C4V5B6N9"

	t.Run("fails when account not found", func(t *testing.T) {
		t.Parallel()
		svc, _, _, _, accRepo, _, _ := setupInvestmentService(t)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(nil, domain.ErrNotFound)

		input := domain.CreatePositionInput{
			AccountID:    accountID,
			AssetID:      assetID,
			Quantity:     10.5,
			AvgCostCents: 1500,
		}

		_, err := svc.CreatePosition(context.Background(), tenantID, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.Contains(t, err.Error(), accountID)
	})

	t.Run("fails when account is not of type investment", func(t *testing.T) {
		t.Parallel()
		svc, _, _, _, accRepo, _, _ := setupInvestmentService(t)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{
			ID:   accountID,
			Type: domain.AccountTypeChecking,
		}, nil)

		input := domain.CreatePositionInput{
			AccountID: accountID,
		}

		_, err := svc.CreatePosition(context.Background(), tenantID, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.Contains(t, err.Error(), "is not investment")
	})

	t.Run("fails when asset not found", func(t *testing.T) {
		t.Parallel()
		svc, _, _, assetRepo, accRepo, _, _ := setupInvestmentService(t)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{
			ID:   accountID,
			Type: domain.AccountTypeInvestment,
		}, nil)

		assetRepo.On("GetByID", mock.Anything, assetID).Return(nil, domain.ErrNotFound)

		input := domain.CreatePositionInput{
			AccountID: accountID,
			AssetID:   assetID,
		}

		_, err := svc.CreatePosition(context.Background(), tenantID, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.Contains(t, err.Error(), assetID)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, assetRepo, accRepo, _, auditRepo := setupInvestmentService(t)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{
			ID:   accountID,
			Type: domain.AccountTypeInvestment,
		}, nil)

		assetRepo.On("GetByID", mock.Anything, assetID).Return(&domain.Asset{
			ID:     assetID,
			Ticker: "AAPL",
		}, nil)

		posRepo.On("Create", mock.Anything, tenantID, mock.MatchedBy(func(p *domain.Position) bool {
			return p.AccountID == accountID && p.AssetID == assetID && p.Quantity == 10.5
		})).Return(&domain.Position{ID: "pos-1", AccountID: accountID, AssetID: assetID, Quantity: 10.5}, nil)

		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionCreate && in.EntityType == "position"
		})).Return(&domain.AuditLog{}, nil)

		openedAt := time.Now()
		input := domain.CreatePositionInput{
			AccountID:    accountID,
			AssetID:      assetID,
			Quantity:     10.5,
			AvgCostCents: 15000,
			OpenedAt:     &openedAt,
		}

		pos, err := svc.CreatePosition(context.Background(), tenantID, input)

		require.NoError(t, err)
		assert.NotNil(t, pos)
		assert.Equal(t, "pos-1", pos.ID)
	})

	t.Run("fails on repository error", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, assetRepo, accRepo, _, _ := setupInvestmentService(t)

		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{Type: domain.AccountTypeInvestment}, nil)
		assetRepo.On("GetByID", mock.Anything, assetID).Return(&domain.Asset{}, nil)
		posRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("db error"))

		_, err := svc.CreatePosition(context.Background(), tenantID, domain.CreatePositionInput{
			AccountID: accountID,
			AssetID:   assetID,
			Quantity:  10,
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create position")
	})

	t.Run("fails on account check error", func(t *testing.T) {
		t.Parallel()
		svc, _, _, _, accRepo, _, _ := setupInvestmentService(t)
		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(nil, errors.New("error"))

		_, err := svc.CreatePosition(context.Background(), tenantID, domain.CreatePositionInput{AccountID: accountID})
		require.Error(t, err)
	})

	t.Run("fails on asset check error", func(t *testing.T) {
		t.Parallel()
		svc, _, _, assetRepo, accRepo, _, _ := setupInvestmentService(t)
		accRepo.On("GetByID", mock.Anything, tenantID, accountID).Return(&domain.Account{Type: domain.AccountTypeInvestment}, nil)
		assetRepo.On("GetByID", mock.Anything, assetID).Return(nil, errors.New("error"))

		_, err := svc.CreatePosition(context.Background(), tenantID, domain.CreatePositionInput{AccountID: accountID, AssetID: assetID})
		require.Error(t, err)
	})
}

func TestInvestmentService_GetPosition(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"
	posID := "pos-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("GetByID", mock.Anything, tenantID, posID).Return(&domain.Position{ID: posID}, nil)

		pos, err := svc.GetPosition(context.Background(), tenantID, posID)
		require.NoError(t, err)
		assert.Equal(t, posID, pos.ID)
	})
}

func TestInvestmentService_ListPositions(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("ListByTenant", mock.Anything, tenantID).Return([]domain.Position{{ID: "1"}}, nil)

		list, err := svc.ListPositions(context.Background(), tenantID)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})
}

func TestInvestmentService_UpdatePosition(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"
	posID := "pos-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, auditRepo := setupInvestmentService(t)

		qty := 20.0
		cost := int64(12000)
		status := domain.PositionStatusClosed
		closedAt := time.Now()
		desc := "notes"

		existing := &domain.Position{ID: posID, Quantity: 10}
		posRepo.On("GetByID", mock.Anything, tenantID, posID).Return(existing, nil)
		posRepo.On("Update", mock.Anything, tenantID, posID, mock.Anything).Return(existing, nil)
		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionUpdate
		})).Return(&domain.AuditLog{}, nil)

		input := domain.UpdatePositionInput{Quantity: &qty, AvgCostCents: &cost, Notes: &desc, Status: &status, ClosedAt: &closedAt}
		pos, err := svc.UpdatePosition(context.Background(), tenantID, posID, input)

		require.NoError(t, err)
		assert.InDelta(t, 20.0, pos.Quantity, 0.001)
		assert.Equal(t, int64(12000), pos.AvgCostCents)
		assert.Equal(t, "notes", *pos.Notes)
		assert.Equal(t, status, pos.Status)
	})

	t.Run("fails when position not found", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("GetByID", mock.Anything, tenantID, posID).Return(nil, domain.ErrNotFound)

		_, err := svc.UpdatePosition(context.Background(), tenantID, posID, domain.UpdatePositionInput{})
		require.Error(t, err)
	})

	t.Run("fails on repo update error", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("GetByID", mock.Anything, tenantID, posID).Return(&domain.Position{}, nil)
		posRepo.On("Update", mock.Anything, tenantID, posID, mock.Anything).Return(nil, errors.New("error"))

		_, err := svc.UpdatePosition(context.Background(), tenantID, posID, domain.UpdatePositionInput{})
		require.Error(t, err)
	})
}

func TestInvestmentService_DeletePosition(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"
	posID := "pos-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, auditRepo := setupInvestmentService(t)
		posRepo.On("GetByID", mock.Anything, tenantID, posID).Return(&domain.Position{}, nil)
		posRepo.On("Delete", mock.Anything, tenantID, posID).Return(nil)
		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionSoftDelete
		})).Return(&domain.AuditLog{}, nil)

		err := svc.DeletePosition(context.Background(), tenantID, posID)
		require.NoError(t, err)
	})

	t.Run("fails on repo error", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("GetByID", mock.Anything, tenantID, posID).Return(&domain.Position{}, nil)
		posRepo.On("Delete", mock.Anything, tenantID, posID).Return(errors.New("error"))

		err := svc.DeletePosition(context.Background(), tenantID, posID)
		require.Error(t, err)
	})
}

func TestInvestmentService_MarkIncomeReceived_Validation(t *testing.T) {
	t.Parallel()

	tenantID := "01J7K7V8N6Y0Q1Z2X3C4V5B6N7"
	eventID := "01J7K7V8N6Y0Q1Z2X3C4V5B6N8"

	t.Run("fails when event already received", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, _, _ := setupInvestmentService(t)

		incRepo.On("GetByID", mock.Anything, tenantID, eventID).Return(&domain.PositionIncomeEvent{
			ID:     eventID,
			Status: domain.PositionIncomeStatusReceived,
		}, nil)

		_, err := svc.MarkIncomeReceived(context.Background(), tenantID, eventID)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.Contains(t, err.Error(), "already received")
	})

	t.Run("fails when event cancelled", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, _, _ := setupInvestmentService(t)

		incRepo.On("GetByID", mock.Anything, tenantID, eventID).Return(&domain.PositionIncomeEvent{
			ID:     eventID,
			Status: domain.PositionIncomeStatusCancelled,
		}, nil)

		_, err := svc.MarkIncomeReceived(context.Background(), tenantID, eventID)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		assert.Contains(t, err.Error(), "already cancelled")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, txRepo, auditRepo := setupInvestmentService(t)

		event := &domain.PositionIncomeEvent{
			ID:          eventID,
			PositionID:  "pos-1",
			AmountCents: 1000,
			Status:      domain.PositionIncomeStatusPending,
		}

		incRepo.On("GetByID", mock.Anything, tenantID, eventID).Return(event, nil)
		incRepo.On("UpdateStatus", mock.Anything, tenantID, eventID, domain.PositionIncomeStatusReceived, mock.Anything).Return(event, nil)
		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.Transaction{ID: "tx-1"}, nil)
		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionUpdate
		})).Return(&domain.AuditLog{}, nil)

		updated, err := svc.MarkIncomeReceived(context.Background(), tenantID, eventID)

		require.NoError(t, err)
		assert.NotNil(t, updated)
	})

	t.Run("fails when event not found", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, _, _ := setupInvestmentService(t)
		incRepo.On("GetByID", mock.Anything, tenantID, eventID).Return(nil, domain.ErrNotFound)

		_, err := svc.MarkIncomeReceived(context.Background(), tenantID, eventID)
		require.Error(t, err)
	})

	t.Run("fails on repo update error", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, _, _ := setupInvestmentService(t)
		incRepo.On("GetByID", mock.Anything, tenantID, eventID).Return(&domain.PositionIncomeEvent{Status: domain.PositionIncomeStatusPending}, nil)
		incRepo.On("UpdateStatus", mock.Anything, tenantID, eventID, domain.PositionIncomeStatusReceived, mock.Anything).Return(nil, errors.New("err"))

		_, err := svc.MarkIncomeReceived(context.Background(), tenantID, eventID)
		require.Error(t, err)
	})

	t.Run("fails on transaction create error", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, txRepo, auditRepo := setupInvestmentService(t)
		incRepo.On("GetByID", mock.Anything, tenantID, eventID).Return(&domain.PositionIncomeEvent{Status: domain.PositionIncomeStatusPending}, nil)
		incRepo.On("UpdateStatus", mock.Anything, tenantID, eventID, domain.PositionIncomeStatusReceived, mock.Anything).Return(&domain.PositionIncomeEvent{}, nil)
		txRepo.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("err"))
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)

		_, err := svc.MarkIncomeReceived(context.Background(), tenantID, eventID)
		require.Error(t, err)
	})
}

func TestInvestmentService_CancelIncome(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"
	eventID := "event-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, _, auditRepo := setupInvestmentService(t)
		incRepo.On("UpdateStatus", mock.Anything, tenantID, eventID, domain.PositionIncomeStatusCancelled, mock.Anything).Return(&domain.PositionIncomeEvent{}, nil)
		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionUpdate
		})).Return(&domain.AuditLog{}, nil)

		_, err := svc.CancelIncome(context.Background(), tenantID, eventID)
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		svc, _, incRepo, _, _, _, _ := setupInvestmentService(t)
		incRepo.On("UpdateStatus", mock.Anything, tenantID, eventID, domain.PositionIncomeStatusCancelled, mock.Anything).Return(nil, errors.New("error"))

		_, err := svc.CancelIncome(context.Background(), tenantID, eventID)
		require.Error(t, err)
	})
}

func TestInvestmentService_GetPortfolioSummary(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)

		posRepo.On("ListByTenant", mock.Anything, tenantID).Return([]domain.Position{
			{ID: "p1", AssetID: "a1", Quantity: 10, AvgCostCents: 1000},
			{ID: "p2", AssetID: "a1", Quantity: 5, AvgCostCents: 2000},
		}, nil)

		summary, err := svc.GetPortfolioSummary(context.Background(), tenantID)
		require.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, int64(0), summary.TotalValueCents)
		assert.Len(t, summary.Positions, 2)
	})

	t.Run("continues on asset error", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)

		posRepo.On("ListByTenant", mock.Anything, tenantID).Return([]domain.Position{
			{ID: "p1", AssetID: "a1", Quantity: 10},
			{ID: "p2", AssetID: "a2", Quantity: 5},
		}, nil)

		summary, err := svc.GetPortfolioSummary(context.Background(), tenantID)
		require.NoError(t, err)
		assert.Len(t, summary.Positions, 2)
	})

	t.Run("fails when list fails", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("ListByTenant", mock.Anything, tenantID).Return(nil, errors.New("err"))

		_, err := svc.GetPortfolioSummary(context.Background(), tenantID)
		require.Error(t, err)
	})
}

func TestInvestmentService_TakeSnapshot(t *testing.T) {
	t.Parallel()
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)

		posRepo.On("ListByTenant", mock.Anything, tenantID).Return([]domain.Position{
			{ID: "p1", AssetID: "a1", Quantity: 10},
		}, nil)

		_, err := svc.TakeSnapshot(context.Background(), tenantID)
		require.NoError(t, err)
	})

	t.Run("failure on summary error", func(t *testing.T) {
		t.Parallel()
		svc, posRepo, _, _, _, _, _ := setupInvestmentService(t)
		posRepo.On("ListByTenant", mock.Anything, tenantID).Return(nil, errors.New("err"))

		_, err := svc.TakeSnapshot(context.Background(), tenantID)
		require.Error(t, err)
	})
}
