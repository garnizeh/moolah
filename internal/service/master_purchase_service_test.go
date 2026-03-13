package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMasterPurchaseService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	input := domain.CreateMasterPurchaseInput{
		AccountID:            "acc_id",
		CategoryID:           "cat_id",
		UserID:               "user_id",
		Description:          "iPhone",
		TotalAmountCents:     120000,
		InstallmentCount:     12,
		ClosingDay:           10,
		FirstInstallmentDate: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		accRepo := new(mocks.AccountRepository)
		catRepo := new(mocks.CategoryRepository)

		accRepo.On("GetByID", ctx, tenantID, input.AccountID).Return(&domain.Account{
			ID:   input.AccountID,
			Type: domain.AccountTypeCreditCard,
		}, nil)

		catRepo.On("GetByID", ctx, tenantID, input.CategoryID).Return(&domain.Category{
			ID:   input.CategoryID,
			Type: domain.CategoryTypeExpense,
		}, nil)

		mp := &domain.MasterPurchase{ID: "mp_id", Description: input.Description}
		mpRepo.On("Create", ctx, tenantID, input).Return(mp, nil)

		svc := service.NewMasterPurchaseService(mpRepo, accRepo, catRepo)
		got, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
		assert.Equal(t, mp, got)
	})

	t.Run("error - account fetch", func(t *testing.T) {
		t.Parallel()
		accRepo := new(mocks.AccountRepository)
		accRepo.On("GetByID", ctx, tenantID, input.AccountID).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(nil, accRepo, nil)
		got, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify account")
		assert.Nil(t, got)
	})

	t.Run("error - not credit card", func(t *testing.T) {
		t.Parallel()
		accRepo := new(mocks.AccountRepository)

		accRepo.On("GetByID", ctx, tenantID, input.AccountID).Return(&domain.Account{
			ID:   input.AccountID,
			Type: domain.AccountTypeChecking,
		}, nil)

		svc := service.NewMasterPurchaseService(nil, accRepo, nil)
		got, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a credit card")
		assert.Nil(t, got)
	})

	t.Run("error - category fetch", func(t *testing.T) {
		t.Parallel()
		accRepo := new(mocks.AccountRepository)
		catRepo := new(mocks.CategoryRepository)

		accRepo.On("GetByID", ctx, tenantID, input.AccountID).Return(&domain.Account{
			ID:   input.AccountID,
			Type: domain.AccountTypeCreditCard,
		}, nil)
		catRepo.On("GetByID", ctx, tenantID, input.CategoryID).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(nil, accRepo, catRepo)
		got, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify category")
		assert.Nil(t, got)
	})

	t.Run("error - category not expense", func(t *testing.T) {
		t.Parallel()
		accRepo := new(mocks.AccountRepository)
		catRepo := new(mocks.CategoryRepository)

		accRepo.On("GetByID", ctx, tenantID, input.AccountID).Return(&domain.Account{
			ID:   input.AccountID,
			Type: domain.AccountTypeCreditCard,
		}, nil)
		catRepo.On("GetByID", ctx, tenantID, input.CategoryID).Return(&domain.Category{
			ID:   input.CategoryID,
			Type: domain.CategoryTypeIncome,
		}, nil)

		svc := service.NewMasterPurchaseService(nil, accRepo, catRepo)
		got, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be of type expense")
		assert.Nil(t, got)
	})

	t.Run("error - repo create", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		accRepo := new(mocks.AccountRepository)
		catRepo := new(mocks.CategoryRepository)

		accRepo.On("GetByID", ctx, tenantID, input.AccountID).Return(&domain.Account{
			ID:   input.AccountID,
			Type: domain.AccountTypeCreditCard,
		}, nil)
		catRepo.On("GetByID", ctx, tenantID, input.CategoryID).Return(&domain.Category{
			ID:   input.CategoryID,
			Type: domain.CategoryTypeExpense,
		}, nil)
		mpRepo.On("Create", ctx, tenantID, input).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(mpRepo, accRepo, catRepo)
		got, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create record")
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseService_ProjectInstallments(t *testing.T) {
	t.Parallel()

	svc := service.NewMasterPurchaseService(nil, nil, nil)
	startDate, err := time.Parse("2006-01-02", "2023-01-01")
	require.NoError(t, err)

	testCases := []struct {
		name             string
		expectedAmounts  []int64
		totalAmountCents int64
		installmentCount int32
	}{
		{
			name:             "1200 / 3 (evenly divisible)",
			totalAmountCents: 1200,
			installmentCount: 3,
			expectedAmounts:  []int64{400, 400, 400},
		},
		{
			name:             "1000 / 3 (remainder 1)",
			totalAmountCents: 1000,
			installmentCount: 3,
			expectedAmounts:  []int64{333, 333, 334},
		},
		{
			name:             "1001 / 3 (remainder 2)",
			totalAmountCents: 1001,
			installmentCount: 3,
			expectedAmounts:  []int64{333, 333, 335},
		},
		{
			name:             "1 / 2 (min base 0)",
			totalAmountCents: 1,
			installmentCount: 2,
			expectedAmounts:  []int64{0, 1},
		},
		{
			name:             "99 / 2",
			totalAmountCents: 99,
			installmentCount: 2,
			expectedAmounts:  []int64{49, 50},
		},
		{
			name:             "10000 / 12",
			totalAmountCents: 10000,
			installmentCount: 12,
			expectedAmounts:  []int64{833, 833, 833, 833, 833, 833, 833, 833, 833, 833, 833, 837},
		},
		{
			name:             "1 / 48",
			totalAmountCents: 1,
			installmentCount: 48,
			expectedAmounts:  append(make([]int64, 47), 1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mp := &domain.MasterPurchase{
				TotalAmountCents:     tc.totalAmountCents,
				InstallmentCount:     tc.installmentCount,
				FirstInstallmentDate: startDate,
			}

			installments := svc.ProjectInstallments(mp)
			require.Len(t, installments, int(tc.installmentCount))

			var sum int64
			for i, inst := range installments {
				assert.Equal(t, tc.expectedAmounts[i], inst.AmountCents, "installment %d amount mismatch", i+1)
				sum += inst.AmountCents
				expectedDate := startDate.AddDate(0, i, 0)
				assert.True(t, expectedDate.Equal(inst.DueDate), "installment %d due date mismatch", i+1)
			}
			assert.Equal(t, tc.totalAmountCents, sum, "total sum mismatch")
		})
	}

	t.Run("zero installments", func(t *testing.T) {
		t.Parallel()
		mp := &domain.MasterPurchase{InstallmentCount: 0}
		installments := svc.ProjectInstallments(mp)
		assert.Nil(t, installments)
	})
}

func TestMasterPurchaseService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	mpID := "mp_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mp := &domain.MasterPurchase{ID: mpID}
		mpRepo.On("GetByID", ctx, tenantID, mpID).Return(mp, nil)

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		got, err := svc.GetByID(ctx, tenantID, mpID)

		require.NoError(t, err)
		assert.Equal(t, mp, got)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mpRepo.On("GetByID", ctx, tenantID, mpID).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		got, err := svc.GetByID(ctx, tenantID, mpID)

		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseService_ListByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mps := []domain.MasterPurchase{{ID: "1"}, {ID: "2"}}
		mpRepo.On("ListByTenant", ctx, tenantID).Return(mps, nil)

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		got, err := svc.ListByTenant(ctx, tenantID)

		require.NoError(t, err)
		assert.Equal(t, mps, got)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mpRepo.On("ListByTenant", ctx, tenantID).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		got, err := svc.ListByTenant(ctx, tenantID)

		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseService_ListByAccount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	accountID := "acc_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mps := []domain.MasterPurchase{{ID: "1"}}
		mpRepo.On("ListByAccount", ctx, tenantID, accountID).Return(mps, nil)

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		got, err := svc.ListByAccount(ctx, tenantID, accountID)

		require.NoError(t, err)
		assert.Equal(t, mps, got)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mpRepo.On("ListByAccount", ctx, tenantID, accountID).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		got, err := svc.ListByAccount(ctx, tenantID, accountID)

		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	mpID := "mp_id"
	catID := "new_cat"
	input := domain.UpdateMasterPurchaseInput{
		CategoryID: &catID,
	}

	t.Run("success with category check", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		catRepo := new(mocks.CategoryRepository)

		catRepo.On("GetByID", ctx, tenantID, catID).Return(&domain.Category{ID: catID, Type: domain.CategoryTypeExpense}, nil)
		mpRepo.On("Update", ctx, tenantID, mpID, input).Return(&domain.MasterPurchase{ID: mpID}, nil)

		svc := service.NewMasterPurchaseService(mpRepo, nil, catRepo)
		got, err := svc.Update(ctx, tenantID, mpID, input)

		require.NoError(t, err)
		assert.NotNil(t, got)
	})

	t.Run("repo update error", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		catRepo := new(mocks.CategoryRepository)
		// We use a different input or just mock the category call since input has CategoryID
		catRepo.On("GetByID", ctx, tenantID, *input.CategoryID).Return(&domain.Category{Type: domain.CategoryTypeExpense}, nil)
		mpRepo.On("Update", ctx, tenantID, mpID, input).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(mpRepo, nil, catRepo)
		got, err := svc.Update(ctx, tenantID, mpID, input)

		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("error - category fetch fail", func(t *testing.T) {
		t.Parallel()
		catRepo := new(mocks.CategoryRepository)
		catRepo.On("GetByID", ctx, tenantID, catID).Return(nil, errors.New("db error"))

		svc := service.NewMasterPurchaseService(nil, nil, catRepo)
		got, err := svc.Update(ctx, tenantID, mpID, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify new category")
		assert.Nil(t, got)
	})

	t.Run("error - category not found", func(t *testing.T) {
		t.Parallel()
		catRepo := new(mocks.CategoryRepository)

		catRepo.On("GetByID", ctx, tenantID, catID).Return(nil, nil)

		svc := service.NewMasterPurchaseService(nil, nil, catRepo)
		got, err := svc.Update(ctx, tenantID, mpID, input)

		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, got)
	})

	t.Run("error - category not expense", func(t *testing.T) {
		t.Parallel()
		catRepo := new(mocks.CategoryRepository)

		catRepo.On("GetByID", ctx, tenantID, catID).Return(&domain.Category{ID: catID, Type: domain.CategoryTypeIncome}, nil)

		svc := service.NewMasterPurchaseService(nil, nil, catRepo)
		got, err := svc.Update(ctx, tenantID, mpID, input)

		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestMasterPurchaseService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	mpID := "mp_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mpRepo.On("Delete", ctx, tenantID, mpID).Return(nil)

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		err := svc.Delete(ctx, tenantID, mpID)

		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mpRepo := new(mocks.MasterPurchaseRepository)
		mpRepo.On("Delete", ctx, tenantID, mpID).Return(errors.New("db error"))

		svc := service.NewMasterPurchaseService(mpRepo, nil, nil)
		err := svc.Delete(ctx, tenantID, mpID)

		require.Error(t, err)
	})
}
