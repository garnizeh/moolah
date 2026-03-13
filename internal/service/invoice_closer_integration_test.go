//go:build integration

package service_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/internal/testutil/seeds"
)

func TestInvoiceCloser_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	// Setup Repositories
	tenantRepo := repository.NewTenantRepository(db.Queries)
	userRepo := repository.NewUserRepository(db.Queries)
	accRepo := repository.NewAccountRepository(db.Queries)
	catRepo := repository.NewCategoryRepository(db.Queries)
	mpRepo := repository.NewMasterPurchaseRepository(db.Queries)
	txRepo := repository.NewTransactionRepository(db.Queries)
	auditRepo := repository.NewAuditRepository(db.Queries)

	// Setup Service
	mpSvc := service.NewMasterPurchaseService(mpRepo, accRepo, catRepo)
	closer := service.NewInvoiceCloser(mpRepo, txRepo, auditRepo, accRepo, mpSvc, db.Pool)

	t.Run("Scenario 1 — Single instalment materialised", func(t *testing.T) {
		t.Parallel()
		today := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

		scenarioTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Scenario 1"})
		require.NoError(t, err)
		scenarioUser, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: scenarioTenant.ID,
			Email:    "s1@example.com",
			Name:     "S1 User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)
		scenarioCategory, err := catRepo.Create(ctx, scenarioTenant.ID, domain.CreateCategoryInput{
			Name: "Gifts",
			Type: domain.CategoryTypeExpense,
		})
		require.NoError(t, err)
		scenarioAccount, err := accRepo.Create(ctx, scenarioTenant.ID, domain.CreateAccountInput{
			UserID:       scenarioUser.ID,
			Name:         "Visa Credit",
			Type:         domain.AccountTypeCreditCard,
			Currency:     "USD",
			InitialCents: 0,
		})
		require.NoError(t, err)

		mp := seeds.SeedMasterPurchase(t, ctx, db.Queries, scenarioTenant.ID, scenarioAccount.ID, scenarioCategory.ID, scenarioUser.ID, 1200, 3, today)

		res, err := closer.CloseInvoice(ctx, scenarioTenant.ID, scenarioAccount.ID, today)
		require.NoError(t, err)
		require.Equal(t, 1, res.ProcessedCount)

		txs, err := txRepo.List(ctx, scenarioTenant.ID, domain.ListTransactionsParams{AccountID: scenarioAccount.ID})
		require.NoError(t, err)

		var mpTx *domain.Transaction
		for _, tx := range txs {
			if tx.MasterPurchaseID == mp.ID {
				mpTx = &tx
				break
			}
		}
		require.NotNil(t, mpTx)
		require.Equal(t, int64(400), mpTx.AmountCents)

		updatedMP, err := mpRepo.GetByID(ctx, scenarioTenant.ID, mp.ID)
		require.NoError(t, err)
		require.Equal(t, int32(1), updatedMP.PaidInstallments)
		require.Equal(t, domain.MasterPurchaseStatusOpen, updatedMP.Status)

		logs, err := auditRepo.ListByTenant(ctx, scenarioTenant.ID, domain.ListAuditLogsParams{ActorID: domain.ActorSystem})
		require.NoError(t, err)
		require.NotEmpty(t, logs)
	})

	t.Run("Scenario 2 — Final instalment closes purchase", func(t *testing.T) {
		t.Parallel()
		today := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

		scenarioTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Scenario 2"})
		require.NoError(t, err)
		scenarioUser, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: scenarioTenant.ID,
			Email:    "s2@example.com",
			Name:     "S2 User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		scenarioCategory := seeds.CreateCategory(t, ctx, db.Queries, scenarioTenant.ID)

		scenarioAccount, err := accRepo.Create(ctx, scenarioTenant.ID, domain.CreateAccountInput{
			UserID:       scenarioUser.ID,
			Name:         "Mastercard Credit",
			Type:         domain.AccountTypeCreditCard,
			Currency:     "USD",
			InitialCents: 0,
		})
		require.NoError(t, err)

		persistedCategory, err := catRepo.GetByID(ctx, scenarioTenant.ID, scenarioCategory.ID)
		require.NoError(t, err)

		mp := seeds.SeedMasterPurchase(t, ctx, db.Queries, scenarioTenant.ID, scenarioAccount.ID, persistedCategory.ID, scenarioUser.ID, 1000, 3, today.AddDate(0, -2, 0))
		seeds.SetMasterPurchasePaidInstallments(t, ctx, db.Queries, scenarioTenant.ID, mp.ID, 2, domain.MasterPurchaseStatusOpen)

		res, err := closer.CloseInvoice(ctx, scenarioTenant.ID, scenarioAccount.ID, today)
		require.NoError(t, err)
		require.Equal(t, 1, res.ProcessedCount)

		txs, err := txRepo.List(ctx, scenarioTenant.ID, domain.ListTransactionsParams{AccountID: scenarioAccount.ID})
		require.NoError(t, err)

		var lastTx *domain.Transaction
		for _, tx := range txs {
			if tx.MasterPurchaseID == mp.ID {
				lastTx = &tx
			}
		}
		require.NotNil(t, lastTx)
		require.Equal(t, int64(334), lastTx.AmountCents)

		updatedMP, err := mpRepo.GetByID(ctx, scenarioTenant.ID, mp.ID)
		require.NoError(t, err)
		require.Equal(t, int32(3), updatedMP.PaidInstallments)
		require.Equal(t, domain.MasterPurchaseStatusClosed, updatedMP.Status)
	})

	t.Run("Scenario 3 — No pending purchases", func(t *testing.T) {
		t.Parallel()
		today := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

		scenarioTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Scenario 3"})
		require.NoError(t, err)
		scenarioUser, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: scenarioTenant.ID,
			Email:    "s3@example.com",
			Name:     "S3 User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		scenarioCategory := seeds.CreateCategory(t, ctx, db.Queries, scenarioTenant.ID)

		scenarioAccount, err := accRepo.Create(ctx, scenarioTenant.ID, domain.CreateAccountInput{
			UserID:       scenarioUser.ID,
			Name:         "Amex Credit",
			Type:         domain.AccountTypeCreditCard,
			Currency:     "USD",
			InitialCents: 0,
		})
		require.NoError(t, err)

		persistedCategory, err := catRepo.GetByID(ctx, scenarioTenant.ID, scenarioCategory.ID)
		require.NoError(t, err)

		mp := seeds.SeedMasterPurchase(t, ctx, db.Queries, scenarioTenant.ID, scenarioAccount.ID, persistedCategory.ID, scenarioUser.ID, 1000, 2, today)
		seeds.SetMasterPurchasePaidInstallments(t, ctx, db.Queries, scenarioTenant.ID, mp.ID, 2, domain.MasterPurchaseStatusClosed)

		res, err := closer.CloseInvoice(ctx, scenarioTenant.ID, scenarioAccount.ID, today)
		require.NoError(t, err)
		require.Equal(t, 0, res.ProcessedCount)
	})

	t.Run("Scenario 4 — Cross-tenant isolation", func(t *testing.T) {
		t.Parallel()
		today := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

		tenantA, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Tenant A isolation"})
		require.NoError(t, err)
		userA, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenantA.ID,
			Email:    "testA@example.com",
			Name:     "User A",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)
		categoryA, err := catRepo.Create(ctx, tenantA.ID, domain.CreateCategoryInput{
			Name: "Gifts",
			Type: domain.CategoryTypeExpense,
		})
		require.NoError(t, err)
		accountA, err := accRepo.Create(ctx, tenantA.ID, domain.CreateAccountInput{
			UserID:       userA.ID,
			Name:         "Visa A",
			Type:         domain.AccountTypeCreditCard,
			Currency:     "USD",
			InitialCents: 0,
		})
		require.NoError(t, err)

		tenantB, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Tenant B isolation"})
		require.NoError(t, err)
		userB, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenantB.ID,
			Email:    "testB@example.com",
			Name:     "User B",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)
		categoryB, err := catRepo.Create(ctx, tenantB.ID, domain.CreateCategoryInput{
			Name: "Gifts",
			Type: domain.CategoryTypeExpense,
		})
		require.NoError(t, err)
		accountB, err := accRepo.Create(ctx, tenantB.ID, domain.CreateAccountInput{
			UserID:       userB.ID,
			Name:         "Visa B",
			Type:         domain.AccountTypeCreditCard,
			Currency:     "USD",
			InitialCents: 0,
		})
		require.NoError(t, err)

		mpA := seeds.SeedMasterPurchase(t, ctx, db.Queries, tenantA.ID, accountA.ID, categoryA.ID, userA.ID, 1000, 2, today)
		_ = seeds.SeedMasterPurchase(t, ctx, db.Queries, tenantB.ID, accountB.ID, categoryB.ID, userB.ID, 2000, 2, today)

		res, err := closer.CloseInvoice(ctx, tenantA.ID, accountA.ID, today)
		require.NoError(t, err)
		require.Equal(t, 1, res.ProcessedCount)

		upA, err := mpRepo.GetByID(ctx, tenantA.ID, mpA.ID)
		require.NoError(t, err)
		require.Equal(t, int32(1), upA.PaidInstallments)
	})

	t.Run("Scenario 5 — Remainder-cent invariant", func(t *testing.T) {
		t.Parallel()
		start := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

		scenarioTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Scenario 5"})
		require.NoError(t, err)
		scenarioUser, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: scenarioTenant.ID,
			Email:    "s5@example.com",
			Name:     "S5 User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)
		scenarioCategory := seeds.CreateCategory(t, ctx, db.Queries, scenarioTenant.ID)
		scenarioAccount, err := accRepo.Create(ctx, scenarioTenant.ID, domain.CreateAccountInput{
			UserID:       scenarioUser.ID,
			Name:         "Prime Card",
			Type:         domain.AccountTypeCreditCard,
			Currency:     "USD",
			InitialCents: 0,
		})
		require.NoError(t, err)

		mp := seeds.SeedMasterPurchase(t, ctx, db.Queries, scenarioTenant.ID, scenarioAccount.ID, scenarioCategory.ID, scenarioUser.ID, 1001, 3, start)

		for i := range 3 {
			closingDate := start.AddDate(0, i, 0)
			res, closeErr := closer.CloseInvoice(ctx, scenarioTenant.ID, scenarioAccount.ID, closingDate)
			require.NoError(t, closeErr)
			require.Equal(t, 1, res.ProcessedCount)
		}

		txs, err := txRepo.List(ctx, scenarioTenant.ID, domain.ListTransactionsParams{AccountID: scenarioAccount.ID})
		require.NoError(t, err)

		var mpTxs []domain.Transaction
		for _, tx := range txs {
			if tx.MasterPurchaseID == mp.ID {
				mpTxs = append(mpTxs, tx)
			}
		}

		sort.Slice(mpTxs, func(i, j int) bool {
			return mpTxs[i].OccurredAt.Before(mpTxs[j].OccurredAt)
		})

		require.Len(t, mpTxs, 3)
		var total int64
		for _, tx := range mpTxs {
			total += tx.AmountCents
		}
		require.Equal(t, int64(1001), total)

		require.Equal(t, int64(333), mpTxs[0].AmountCents)
		require.Equal(t, int64(333), mpTxs[1].AmountCents)
		require.Equal(t, int64(335), mpTxs[2].AmountCents)
	})
}
