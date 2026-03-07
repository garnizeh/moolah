//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
)

func TestAccountRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	tenantRepo := repository.NewTenantRepository(db.Queries)
	userRepo := repository.NewUserRepository(db.Queries)
	repo := repository.NewAccountRepository(db.Queries)

	// Setup: Create tenant and user
	tenant, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Account Tenant"})
	user, _ := userRepo.Create(ctx, domain.CreateUserInput{
		TenantID: tenant.ID,
		Email:    "accuser@example.com",
		Name:     "Account User",
		Role:     domain.RoleMember,
	})

	t.Run("Create and GetByID", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateAccountInput{
			UserID:       user.ID,
			Name:         "Savings Account",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 10000,
		}

		created, err := repo.Create(ctx, tenant.ID, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, input.Name, created.Name)
		require.Equal(t, input.InitialCents, created.BalanceCents)

		got, err := repo.GetByID(ctx, tenant.ID, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, created.Name, got.Name)
	})

	t.Run("Cross-Tenant GetByID", func(t *testing.T) {
		t.Parallel()
		otherTenant, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Other T"})
		acc, _ := repo.Create(ctx, tenant.ID, domain.CreateAccountInput{
			UserID:       user.ID,
			Name:         "Tenant 1 Account",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 0,
		})

		_, err := repo.GetByID(ctx, otherTenant.ID, acc.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("List Methods", func(t *testing.T) {
		t.Parallel()
		t2, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "List T"})
		u21, _ := userRepo.Create(ctx, domain.CreateUserInput{TenantID: t2.ID, Email: "u21@t2.com", Name: "U1", Role: domain.RoleMember})
		u2_2, _ := userRepo.Create(ctx, domain.CreateUserInput{TenantID: t2.ID, Email: "u22@t2.com", Name: "U2", Role: domain.RoleMember})

		a1, _ := repo.Create(ctx, t2.ID, domain.CreateAccountInput{UserID: u21.ID, Name: "A1", Type: domain.AccountTypeSavings, Currency: "USD"})
		_, _ = repo.Create(ctx, t2.ID, domain.CreateAccountInput{UserID: u2_2.ID, Name: "A2", Type: domain.AccountTypeSavings, Currency: "USD"})

		// ListByTenant
		all, err := repo.ListByTenant(ctx, t2.ID)
		require.NoError(t, err)
		require.Len(t, all, 2)

		// ListByUser
		u1Accs, err := repo.ListByUser(ctx, t2.ID, u21.ID)
		require.NoError(t, err)
		require.Len(t, u1Accs, 1)
		require.Equal(t, a1.ID, u1Accs[0].ID)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		acc, _ := repo.Create(ctx, tenant.ID, domain.CreateAccountInput{
			UserID:       user.ID,
			Name:         "Before Update",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 500,
		})

		newName := "Updated Name"
		newCurrency := "EUR"
		updated, err := repo.Update(ctx, tenant.ID, acc.ID, domain.UpdateAccountInput{
			Name:     &newName,
			Currency: &newCurrency,
		})
		require.NoError(t, err)
		require.Equal(t, newName, updated.Name)
		require.Equal(t, newCurrency, updated.Currency)
		require.Equal(t, int64(500), updated.BalanceCents) // should preserve
	})

	t.Run("UpdateBalance", func(t *testing.T) {
		t.Parallel()
		acc, _ := repo.Create(ctx, tenant.ID, domain.CreateAccountInput{
			UserID:       user.ID,
			Name:         "Balance Test",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 100,
		})

		err := repo.UpdateBalance(ctx, tenant.ID, acc.ID, 500)
		require.NoError(t, err)

		got, _ := repo.GetByID(ctx, tenant.ID, acc.ID)
		require.Equal(t, int64(500), got.BalanceCents)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		t.Parallel()
		acc, _ := repo.Create(ctx, tenant.ID, domain.CreateAccountInput{
			UserID:   user.ID,
			Name:     "For Deletion",
			Type:     domain.AccountTypeSavings,
			Currency: "USD",
		})

		err := repo.Delete(ctx, tenant.ID, acc.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, tenant.ID, acc.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)

		list, _ := repo.ListByTenant(ctx, tenant.ID)
		for _, a := range list {
			require.NotEqual(t, acc.ID, a.ID)
		}
	})
}
