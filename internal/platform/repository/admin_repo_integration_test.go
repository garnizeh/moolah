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

func TestAdminRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	// Repositories
	tenantRepo := repository.NewTenantRepository(db.Queries)
	adminTenantRepo := repository.NewAdminTenantRepository(db.Queries)

	userRepo := repository.NewUserRepository(db.Queries)
	adminUserRepo := repository.NewAdminUserRepository(db.Queries)

	t.Run("AdminTenantRepo", func(t *testing.T) {
		t.Parallel()
		// Setup: Create some tenants
		t1, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "T1"})
		t2, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "T2"})

		// 1. ListAll
		all, err := adminTenantRepo.ListAll(ctx, false)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(all), 2)

		// 2. Suspend (Soft Delete)
		err = adminTenantRepo.Suspend(ctx, t1.ID)
		require.NoError(t, err)

		// GetByID (Internal tenant repo) should see it as not found
		_, err = tenantRepo.GetByID(ctx, t1.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)

		// AdminGetByID should still find it
		got, err := adminTenantRepo.GetByID(ctx, t1.ID)
		require.NoError(t, err)
		require.Equal(t, t1.ID, got.ID)

		// 3. Restore
		err = adminTenantRepo.Restore(ctx, t1.ID)
		require.NoError(t, err)
		_, err = tenantRepo.GetByID(ctx, t1.ID)
		require.NoError(t, err)

		// 4. UpdatePlan
		updated, err := adminTenantRepo.UpdatePlan(ctx, t2.ID, domain.TenantPlanPremium)
		require.NoError(t, err)
		require.Equal(t, domain.TenantPlanPremium, updated.Plan)

		// 5. HardDelete
		err = adminTenantRepo.HardDelete(ctx, t2.ID)
		require.NoError(t, err)

		_, err = adminTenantRepo.GetByID(ctx, t2.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("AdminUserRepo", func(t *testing.T) {
		t.Parallel()
		tenant, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Admin User T"})
		u1, _ := userRepo.Create(ctx, domain.CreateUserInput{TenantID: tenant.ID, Email: "adminu1@test.com", Name: "U1", Role: domain.RoleMember})
		u2Tenant, _ := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Admin User T2"})
		u2, _ := userRepo.Create(ctx, domain.CreateUserInput{TenantID: u2Tenant.ID, Email: "adminu2@test.com", Name: "U2", Role: domain.RoleMember})

		// 1. ListAll (Should see users from both tenants)
		all, err := adminUserRepo.ListAll(ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(all), 2)

		// Check both are present
		foundU1 := false
		foundU2 := false
		for _, u := range all {
			if u.ID == u1.ID {
				foundU1 = true
			}
			if u.ID == u2.ID {
				foundU2 = true
			}
		}
		require.True(t, foundU1)
		require.True(t, foundU2)

		allAfter, _ := adminUserRepo.ListAll(ctx)
		for _, u := range allAfter {
			require.NotEqual(t, u2.ID, u.ID)
		}
	})
}
