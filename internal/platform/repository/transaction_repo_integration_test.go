//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
)

func TestTransactionRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	tenantRepo := repository.NewTenantRepository(db.Queries)
	userRepo := repository.NewUserRepository(db.Queries)
	accountRepo := repository.NewAccountRepository(db.Queries)
	categoryRepo := repository.NewCategoryRepository(db.Queries)
	repo := repository.NewTransactionRepository(db.Queries)

	// Setup: Create tenant, user, account, category
	tenant, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "TX Tenant"})
	user, _ := userRepo.Create(ctx, domain.CreateUserInput{
		TenantID: tenant.ID,
		Email:    "txuser@example.com",
		Name:     "TX User",
		Role:     domain.RoleMember,
	})
	acc, _ := accountRepo.Create(ctx, tenant.ID, domain.CreateAccountInput{
		UserID:   user.ID,
		Name:     "TX Account",
		Type:     domain.AccountTypeChecking,
		Currency: "USD",
	})

	cat, _ := categoryRepo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
		Name: "TX Category",
		Type: domain.CategoryTypeExpense,
	})

	t.Run("Create and GetByID", func(t *testing.T) {
		t.Parallel()
		now := time.Now().Truncate(time.Microsecond).UTC()
		input := domain.CreateTransactionInput{
			AccountID:   acc.ID,
			CategoryID:  cat.ID,
			UserID:      user.ID,
			Description: "Groceries",
			AmountCents: 5000,
			Type:        domain.TransactionTypeExpense,
			OccurredAt:  now,
		}

		created, err := repo.Create(ctx, tenant.ID, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, input.Description, created.Description)
		require.Equal(t, input.AmountCents, created.AmountCents)
		require.Equal(t, input.Type, created.Type)
		require.Equal(t, now, created.OccurredAt.UTC())

		got, err := repo.GetByID(ctx, tenant.ID, created.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, created.Description, got.Description)
	})

	t.Run("Create with MasterPurchaseID", func(t *testing.T) {
		t.Parallel()
		masterID := "01KK4R6V8NAM0H3AMT18TDAFRX" // Valid ULID format
		input := domain.CreateTransactionInput{
			AccountID:        acc.ID,
			CategoryID:       cat.ID,
			UserID:           user.ID,
			Description:      "Installment 1",
			AmountCents:      1000,
			Type:             domain.TransactionTypeExpense,
			OccurredAt:       time.Now().UTC(),
			MasterPurchaseID: masterID,
		}

		created, err := repo.Create(ctx, tenant.ID, input)
		require.NoError(t, err)
		require.Equal(t, masterID, created.MasterPurchaseID)

		got, _ := repo.GetByID(ctx, tenant.ID, created.ID)
		require.Equal(t, masterID, got.MasterPurchaseID)
	})

	t.Run("List with Filters", func(t *testing.T) {
		t.Parallel()
		t2, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Filter T"})
		u2, _ := userRepo.Create(ctx, domain.CreateUserInput{TenantID: t2.ID, Email: "u2@tx.com", Name: "U2", Role: domain.RoleMember})
		a1, _ := accountRepo.Create(ctx, t2.ID, domain.CreateAccountInput{UserID: u2.ID, Name: "A1", Type: domain.AccountTypeChecking, Currency: "USD"})
		a2, _ := accountRepo.Create(ctx, t2.ID, domain.CreateAccountInput{UserID: u2.ID, Name: "A2", Type: domain.AccountTypeChecking, Currency: "USD"})
		c1, _ := categoryRepo.Create(ctx, t2.ID, domain.CreateCategoryInput{Name: "C1", Type: domain.CategoryTypeExpense})

		// One in A1, yesterday
		yest := time.Now().AddDate(0, 0, -1).UTC()
		_, _ = repo.Create(ctx, t2.ID, domain.CreateTransactionInput{AccountID: a1.ID, CategoryID: c1.ID, UserID: u2.ID, Description: "Yest", AmountCents: 10, Type: domain.TransactionTypeExpense, OccurredAt: yest})
		// One in A2, today
		today := time.Now().UTC()
		_, _ = repo.Create(ctx, t2.ID, domain.CreateTransactionInput{AccountID: a2.ID, CategoryID: c1.ID, UserID: u2.ID, Description: "Today", AmountCents: 20, Type: domain.TransactionTypeExpense, OccurredAt: today})

		// Filter by Account A1
		res, err := repo.List(ctx, t2.ID, domain.ListTransactionsParams{AccountID: a1.ID})
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.Equal(t, "Yest", res[0].Description)

		// Filter by Date (only today)
		startToday := today.Add(-1 * time.Hour)
		res, err = repo.List(ctx, t2.ID, domain.ListTransactionsParams{StartDate: &startToday})
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.Equal(t, "Today", res[0].Description)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		tx, _ := repo.Create(ctx, tenant.ID, domain.CreateTransactionInput{
			AccountID:   acc.ID,
			CategoryID:  cat.ID,
			UserID:      user.ID,
			Description: "Before Update",
			AmountCents: 100,
			Type:        domain.TransactionTypeExpense,
			OccurredAt:  time.Now().UTC(),
		})

		newDesc := "After Update"
		newAmount := int64(200)
		updated, err := repo.Update(ctx, tenant.ID, tx.ID, domain.UpdateTransactionInput{
			Description: &newDesc,
			AmountCents: &newAmount,
		})
		require.NoError(t, err)
		require.Equal(t, newDesc, updated.Description)
		require.Equal(t, newAmount, updated.AmountCents)
		require.Equal(t, acc.ID, updated.AccountID) // preserved
	})

	t.Run("SoftDelete", func(t *testing.T) {
		t.Parallel()
		tx, _ := repo.Create(ctx, tenant.ID, domain.CreateTransactionInput{
			AccountID:   acc.ID,
			CategoryID:  cat.ID,
			UserID:      user.ID,
			Description: "To Delete",
			AmountCents: 10,
			Type:        domain.TransactionTypeExpense,
			OccurredAt:  time.Now().UTC(),
		})

		err := repo.Delete(ctx, tenant.ID, tx.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, tenant.ID, tx.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)

		list, _ := repo.List(ctx, tenant.ID, domain.ListTransactionsParams{})
		for _, row := range list {
			require.NotEqual(t, tx.ID, row.ID)
		}
	})
}
