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
	"github.com/garnizeh/moolah/pkg/ulid"
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
	tenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "TX Tenant"})
	require.NoError(t, err)

	user, err := userRepo.Create(ctx, domain.CreateUserInput{
		TenantID: tenant.ID,
		Email:    "txuser@example.com",
		Name:     "TX User",
		Role:     domain.RoleMember,
	})
	require.NoError(t, err)

	acc, err := accountRepo.Create(ctx, tenant.ID, domain.CreateAccountInput{
		UserID:   user.ID,
		Name:     "TX Account",
		Type:     domain.AccountTypeChecking,
		Currency: "USD",
	})
	require.NoError(t, err)

	cat, err := categoryRepo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
		Name: "TX Category",
		Type: domain.CategoryTypeExpense,
	})
	require.NoError(t, err)

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
		masterID := ulid.New()
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO master_purchases (
				id, tenant_id, account_id, category_id, user_id,
				description, total_amount_cents, installment_count, installment_cents,
				first_due_date, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, NOW(), NOW()
			)
		`, masterID, tenant.ID, acc.ID, cat.ID, user.ID, "Master Purchase", int64(12000), int16(12), int64(1000), time.Now().UTC())
		require.NoError(t, err)

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

		got, err := repo.GetByID(ctx, tenant.ID, created.ID)
		require.NoError(t, err)
		require.Equal(t, masterID, got.MasterPurchaseID)
	})

	t.Run("List with Filters", func(t *testing.T) {
		t.Parallel()
		t2, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Filter T"})
		require.NoError(t, err)
		u2, err := userRepo.Create(ctx, domain.CreateUserInput{TenantID: t2.ID, Email: "u2@tx.com", Name: "U2", Role: domain.RoleMember})
		require.NoError(t, err)
		a1, err := accountRepo.Create(ctx, t2.ID, domain.CreateAccountInput{UserID: u2.ID, Name: "A1", Type: domain.AccountTypeChecking, Currency: "USD"})
		require.NoError(t, err)
		a2, err := accountRepo.Create(ctx, t2.ID, domain.CreateAccountInput{UserID: u2.ID, Name: "A2", Type: domain.AccountTypeChecking, Currency: "USD"})
		require.NoError(t, err)
		c1, err := categoryRepo.Create(ctx, t2.ID, domain.CreateCategoryInput{Name: "C1", Type: domain.CategoryTypeExpense})
		require.NoError(t, err)

		// One in A1, yesterday
		yest := time.Now().AddDate(0, 0, -1).UTC()
		_, err = repo.Create(ctx, t2.ID, domain.CreateTransactionInput{AccountID: a1.ID, CategoryID: c1.ID, UserID: u2.ID, Description: "Yest", AmountCents: 10, Type: domain.TransactionTypeExpense, OccurredAt: yest})
		require.NoError(t, err)
		// One in A2, today
		today := time.Now().UTC()
		_, err = repo.Create(ctx, t2.ID, domain.CreateTransactionInput{AccountID: a2.ID, CategoryID: c1.ID, UserID: u2.ID, Description: "Today", AmountCents: 20, Type: domain.TransactionTypeExpense, OccurredAt: today})
		require.NoError(t, err)

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
		tx, err := repo.Create(ctx, tenant.ID, domain.CreateTransactionInput{
			AccountID:   acc.ID,
			CategoryID:  cat.ID,
			UserID:      user.ID,
			Description: "Before Update",
			AmountCents: 100,
			Type:        domain.TransactionTypeExpense,
			OccurredAt:  time.Now().UTC(),
		})
		require.NoError(t, err)

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
		tx, err := repo.Create(ctx, tenant.ID, domain.CreateTransactionInput{
			AccountID:   acc.ID,
			CategoryID:  cat.ID,
			UserID:      user.ID,
			Description: "To Delete",
			AmountCents: 10,
			Type:        domain.TransactionTypeExpense,
			OccurredAt:  time.Now().UTC(),
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, tenant.ID, tx.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, tenant.ID, tx.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)

		list, err := repo.List(ctx, tenant.ID, domain.ListTransactionsParams{})
		require.NoError(t, err)
		for _, row := range list {
			require.NotEqual(t, tx.ID, row.ID)
		}
	})
}
