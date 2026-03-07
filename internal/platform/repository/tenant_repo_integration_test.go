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

func TestTenantRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := containers.NewPostgresDB(t)

	repo := repository.NewTenantRepository(db.Queries)

	t.Run("Create and GetByID", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateTenantInput{
			Name: "Integration Tenant",
		}

		created, err := repo.Create(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, input.Name, created.Name)
		require.Equal(t, domain.TenantPlanFree, created.Plan)

		got, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, created.Name, got.Name)
	})

	t.Run("GetByID Not Found", func(t *testing.T) {
		t.Parallel()
		got, err := repo.GetByID(ctx, "non-existent-id")
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, got)
	})

	t.Run("Create Conflict", func(t *testing.T) {
		t.Parallel()
		name := "Duplicate Tenant"
		_, err := repo.Create(ctx, domain.CreateTenantInput{Name: name})
		require.NoError(t, err)

		_, err = repo.Create(ctx, domain.CreateTenantInput{Name: name})
		require.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		tenant, err := repo.Create(ctx, domain.CreateTenantInput{Name: "To Update"})
		require.NoError(t, err)

		newName := "Updated Name"
		newPlan := domain.TenantPlanPremium

		updated, err := repo.Update(ctx, tenant.ID, domain.UpdateTenantInput{
			Name: &newName,
			Plan: &newPlan,
		})
		require.NoError(t, err)
		require.Equal(t, newName, updated.Name)
		require.Equal(t, newPlan, updated.Plan)

		got, err := repo.GetByID(ctx, tenant.ID)
		require.NoError(t, err)
		require.Equal(t, newName, got.Name)
		require.Equal(t, newPlan, got.Plan)
	})

	t.Run("SoftDelete and List", func(t *testing.T) {
		t.Parallel()
		t1, err := repo.Create(ctx, domain.CreateTenantInput{Name: "Tenant 1"})
		require.NoError(t, err)
		t2, err := repo.Create(ctx, domain.CreateTenantInput{Name: "Tenant 2"})
		require.NoError(t, err)

		list, err := repo.List(ctx)
		require.NoError(t, err)
		require.Contains(t, ids(list), t1.ID)
		require.Contains(t, ids(list), t2.ID)

		err = repo.Delete(ctx, t1.ID)
		require.NoError(t, err)

		list, err = repo.List(ctx)
		require.NoError(t, err)
		require.NotContains(t, ids(list), t1.ID)
		require.Contains(t, ids(list), t2.ID)

		// GetByID should also fail for soft-deleted
		_, err = repo.GetByID(ctx, t1.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func ids(tenants []domain.Tenant) []string {
	res := make([]string, len(tenants))
	for i, t := range tenants {
		res[i] = t.ID
	}
	return res
}
