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

func TestUserRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	tenantRepo := repository.NewTenantRepository(db.Queries)
	userRepo := repository.NewUserRepository(db.Queries)

	// Setup: Create a tenant first as users belong to tenants
	tenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "User Test Tenant"})
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    "test@example.com",
			Name:     "Test User",
			Role:     domain.RoleAdmin,
		}

		created, err := userRepo.Create(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, input.Email, created.Email)
		require.Equal(t, input.Name, created.Name)
		require.Equal(t, input.Role, created.Role)

		got, err := userRepo.GetByID(ctx, tenant.ID, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, created.Email, got.Email)
	})

	t.Run("GetByID Cross-Tenant or Not Found", func(t *testing.T) {
		t.Parallel()
		// Another tenant
		otherTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Other Tenant"})
		require.NoError(t, err)
		user, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    "cross@example.com",
			Name:     "Cross User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		// Try to fetch from wrong tenant
		got, err := userRepo.GetByID(ctx, otherTenant.ID, user.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, got)

		// Try non-existent ID
		got, err = userRepo.GetByID(ctx, tenant.ID, "non-existent")
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, got)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		t.Parallel()
		email := "byemail@example.com"
		user, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    email,
			Name:     "Email User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		got, err := userRepo.GetByEmail(ctx, email)
		require.NoError(t, err)
		require.Equal(t, user.ID, got.ID)
		require.Equal(t, email, got.Email)

		// Not found
		_, err = userRepo.GetByEmail(ctx, "unknown@example.com")
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("Create Duplicate Email Conflict", func(t *testing.T) {
		t.Parallel()
		email := "duplicate@example.com"
		_, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    email,
			Name:     "User 1",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		// Try to create another user IN THE SAME tenant with same email
		_, err = userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    email,
			Name:     "User 2",
			Role:     domain.RoleMember,
		})
		require.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		t3, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "List Tenant"})
		require.NoError(t, err)
		u1, err := userRepo.Create(ctx, domain.CreateUserInput{TenantID: t3.ID, Email: "u1@t3.com", Name: "U1", Role: domain.RoleMember})
		require.NoError(t, err)
		u2, err := userRepo.Create(ctx, domain.CreateUserInput{TenantID: t3.ID, Email: "u2@t3.com", Name: "U2", Role: domain.RoleMember})
		require.NoError(t, err)

		list, err := userRepo.ListByTenant(ctx, t3.ID)
		require.NoError(t, err)
		require.Len(t, list, 2)

		ids := make([]string, len(list))
		for i, u := range list {
			ids[i] = u.ID
		}
		require.Contains(t, ids, u1.ID)
		require.Contains(t, ids, u2.ID)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		user, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    "update@example.com",
			Name:     "Original Name",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		newName := "Updated Name"
		newRole := domain.RoleAdmin
		updated, err := userRepo.Update(ctx, tenant.ID, user.ID, domain.UpdateUserInput{
			Name: &newName,
			Role: &newRole,
		})
		require.NoError(t, err)
		require.Equal(t, newName, updated.Name)
		require.Equal(t, newRole, updated.Role)

		got, err := userRepo.GetByID(ctx, tenant.ID, user.ID)
		require.NoError(t, err)
		require.Equal(t, newName, got.Name)
	})

	t.Run("UpdateLastLogin", func(t *testing.T) {
		t.Parallel()
		user, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    "login@example.com",
			Name:     "Login User",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		require.Nil(t, user.LastLoginAt)

		err = userRepo.UpdateLastLogin(ctx, user.ID)
		require.NoError(t, err)

		got, err := userRepo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, got.LastLoginAt)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		t.Parallel()
		user, err := userRepo.Create(ctx, domain.CreateUserInput{
			TenantID: tenant.ID,
			Email:    "delete@example.com",
			Name:     "Going Away",
			Role:     domain.RoleMember,
		})
		require.NoError(t, err)

		err = userRepo.Delete(ctx, tenant.ID, user.ID)
		require.NoError(t, err)

		_, err = userRepo.GetByID(ctx, tenant.ID, user.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)

		list, err := userRepo.ListByTenant(ctx, tenant.ID)
		require.NoError(t, err)
		for _, u := range list {
			require.NotEqual(t, user.ID, u.ID)
		}
	})
}
