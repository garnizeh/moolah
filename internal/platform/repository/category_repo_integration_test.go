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

func TestCategoryRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	tenantRepo := repository.NewTenantRepository(db.Queries)
	repo := repository.NewCategoryRepository(db.Queries)

	// Setup: Create a tenant first
	tenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Category Tenant"})
	require.NoError(t, err)

	t.Run("Create and GetByID", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateCategoryInput{
			Name:  "Food",
			Type:  domain.CategoryTypeExpense,
			Icon:  "utensils",
			Color: "#FF0000",
		}

		created, err := repo.Create(ctx, tenant.ID, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, input.Name, created.Name)
		require.Equal(t, input.Icon, created.Icon)
		require.Equal(t, input.Color, created.Color)

		got, err := repo.GetByID(ctx, tenant.ID, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
	})

	t.Run("Create with Parent and ListChildren", func(t *testing.T) {
		t.Parallel()
		parent, err := repo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
			Name: "Transport",
			Type: domain.CategoryTypeExpense,
		})
		require.NoError(t, err)

		child, err := repo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
			ParentID: parent.ID,
			Name:     "Fuel",
			Type:     domain.CategoryTypeExpense,
		})
		require.NoError(t, err)

		children, err := repo.ListChildren(ctx, tenant.ID, parent.ID)
		require.NoError(t, err)
		require.Len(t, children, 1)
		require.Equal(t, child.ID, children[0].ID)

		// ListByTenant should include both
		all, err := repo.ListByTenant(ctx, tenant.ID)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(all), 2)
	})

	t.Run("Cross-Tenant Isolation", func(t *testing.T) {
		t.Parallel()
		otherTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Other T"})
		require.NoError(t, err)
		cat, err := repo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
			Name: "Secret Category",
			Type: domain.CategoryTypeIncome,
		})
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, otherTenant.ID, cat.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()
		cat, err := repo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
			Name: "Original",
			Type: domain.CategoryTypeExpense,
		})
		require.NoError(t, err)

		newName := "Modified"
		newIcon := "star"
		updated, err := repo.Update(ctx, tenant.ID, cat.ID, domain.UpdateCategoryInput{
			Name: &newName,
			Icon: &newIcon,
		})
		require.NoError(t, err)
		require.Equal(t, newName, updated.Name)
		require.Equal(t, newIcon, updated.Icon)
	})

	t.Run("SoftDelete", func(t *testing.T) {
		t.Parallel()
		cat, err := repo.Create(ctx, tenant.ID, domain.CreateCategoryInput{
			Name: "To Delete",
			Type: domain.CategoryTypeExpense,
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, tenant.ID, cat.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, tenant.ID, cat.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}
