package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	input := domain.CreateCategoryInput{
		Name:     "Food",
		Icon:     "utensils",
		Color:    "#FF0000",
		Type:     domain.CategoryTypeExpense,
		ParentID: "parent_id",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("CreateCategory", ctx, mock.MatchedBy(func(p sqlc.CreateCategoryParams) bool {
			return p.TenantID == tenantID &&
				p.Name == input.Name &&
				p.Icon.String == input.Icon &&
				p.Color.String == input.Color &&
				p.Type == sqlc.CategoryType(input.Type) &&
				p.ParentID == input.ParentID
		})).Return(sqlc.Category{
			ID:       "cat_id",
			TenantID: tenantID,
			ParentID: input.ParentID,
			Name:     input.Name,
			Icon:     pgtype.Text{String: input.Icon, Valid: true},
			Color:    pgtype.Text{String: input.Color, Valid: true},
			Type:     sqlc.CategoryType(input.Type),
		}, nil)

		got, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		assert.Equal(t, "cat_id", got.ID)
		assert.Equal(t, input.Name, got.Name)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("CreateCategory", ctx, mock.Anything).Return(sqlc.Category{}, &pgconn.PgError{Code: "23505"})

		got, err := repo.Create(ctx, tenantID, input)
		require.ErrorIs(t, err, domain.ErrConflict)
		assert.Nil(t, got)
	})
}

func TestCategoryRepository_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "cat_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, sqlc.GetCategoryByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(sqlc.Category{
			ID:       id,
			TenantID: tenantID,
			Name:     "Food",
		}, nil)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.NoError(t, err)
		assert.Equal(t, id, got.ID)
		assert.Equal(t, "Food", got.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{}, pgx.ErrNoRows)

		got, err := repo.GetByID(ctx, tenantID, id)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}

func TestCategoryRepository_ListByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("ListCategoriesByTenant", ctx, tenantID).Return([]sqlc.Category{
			{ID: "1", Name: "Food"},
			{ID: "2", Name: "Rent"},
		}, nil)

		got, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})
}

func TestCategoryRepository_ListChildren(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	parentID := "parent_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("ListChildCategories", ctx, sqlc.ListChildCategoriesParams{
			TenantID: tenantID,
			ParentID: parentID,
		}).Return([]sqlc.Category{
			{ID: "1", Name: "Sub1", ParentID: parentID},
		}, nil)

		got, err := repo.ListChildren(ctx, tenantID, parentID)
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, parentID, got[0].ParentID)
	})
}

func TestCategoryRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "cat_id"
	newName := "New Food"
	input := domain.UpdateCategoryInput{
		Name: &newName,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		current := sqlc.Category{
			ID:       id,
			TenantID: tenantID,
			Name:     "Old Food",
			Icon:     pgtype.Text{String: "old-icon", Valid: true},
			Color:    pgtype.Text{String: "#000", Valid: true},
			Type:     sqlc.CategoryTypeExpense,
		}

		mockQuerier.On("GetCategoryByID", ctx, sqlc.GetCategoryByIDParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(current, nil)

		mockQuerier.On("UpdateCategory", ctx, mock.MatchedBy(func(p sqlc.UpdateCategoryParams) bool {
			return p.ID == id && p.Name == newName && p.Icon.String == "old-icon"
		})).Return(sqlc.Category{
			ID:       id,
			TenantID: tenantID,
			Name:     newName,
		}, nil)

		got, err := repo.Update(ctx, tenantID, id, input)
		require.NoError(t, err)
		assert.Equal(t, newName, got.Name)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{ID: id}, nil)
		mockQuerier.On("UpdateCategory", ctx, mock.Anything).Return(sqlc.Category{}, &pgconn.PgError{Code: "23505"})

		got, err := repo.Update(ctx, tenantID, id, input)
		require.ErrorIs(t, err, domain.ErrConflict)
		assert.Nil(t, got)
	})
}

func TestCategoryRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "cat_id"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("SoftDeleteCategory", ctx, sqlc.SoftDeleteCategoryParams{
			TenantID: tenantID,
			ID:       id,
		}).Return(nil)

		err := repo.Delete(ctx, tenantID, id)
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("SoftDeleteCategory", ctx, mock.Anything).Return(errors.New("db error"))

		err := repo.Delete(ctx, tenantID, id)
		require.Error(t, err)
	})
}

func TestCategoryRepository_GenericErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "cat_id"

	t.Run("create error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("CreateCategory", ctx, mock.Anything).Return(sqlc.Category{}, errors.New("db error"))

		got, err := repo.Create(ctx, tenantID, domain.CreateCategoryInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create category")
		assert.Nil(t, got)
	})

	t.Run("get by id generic error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, mock.MatchedBy(func(p sqlc.GetCategoryByIDParams) bool {
			return p.ID == id
		})).Return(sqlc.Category{}, errors.New("db error"))

		got, err := repo.GetByID(ctx, tenantID, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get category")
		assert.Nil(t, got)
	})

	t.Run("list by tenant error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("ListCategoriesByTenant", ctx, mock.Anything).Return([]sqlc.Category(nil), errors.New("db error"))

		got, err := repo.ListByTenant(ctx, tenantID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list categories")
		assert.Nil(t, got)
	})

	t.Run("list children error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("ListChildCategories", ctx, mock.Anything).Return([]sqlc.Category(nil), errors.New("db error"))

		got, err := repo.ListChildren(ctx, tenantID, "parent_id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list child categories")
		assert.Nil(t, got)
	})

	t.Run("update get current generic error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{}, errors.New("db error"))

		got, err := repo.Update(ctx, tenantID, id, domain.UpdateCategoryInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get category for update")
		assert.Nil(t, got)
	})

	t.Run("update generic database error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{ID: id}, nil)
		mockQuerier.On("UpdateCategory", ctx, mock.Anything).Return(sqlc.Category{}, errors.New("db error"))

		got, err := repo.Update(ctx, tenantID, id, domain.UpdateCategoryInput{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update category")
		assert.Nil(t, got)
	})
}

func TestCategoryRepository_UpdatePartial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_id"
	id := "cat_id"

	t.Run("update icon and color", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		icon := "new-icon"
		color := "#NEW"
		input := domain.UpdateCategoryInput{
			Icon:  &icon,
			Color: &color,
		}

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{
			ID:   id,
			Name: "Stay",
			Icon: pgtype.Text{String: "old", Valid: true},
		}, nil)

		mockQuerier.On("UpdateCategory", ctx, mock.MatchedBy(func(p sqlc.UpdateCategoryParams) bool {
			return p.Icon.String == icon && p.Color.String == color && p.Name == "Stay"
		})).Return(sqlc.Category{ID: id}, nil)

		_, err := repo.Update(ctx, tenantID, id, input)
		require.NoError(t, err)
	})

	t.Run("update name only", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		name := "New Name"
		input := domain.UpdateCategoryInput{
			Name: &name,
		}

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{
			ID:   id,
			Name: "Old",
			Icon: pgtype.Text{String: "icon", Valid: true},
		}, nil)

		mockQuerier.On("UpdateCategory", ctx, mock.MatchedBy(func(p sqlc.UpdateCategoryParams) bool {
			return p.Name == name && p.Icon.String == "icon"
		})).Return(sqlc.Category{ID: id}, nil)

		_, err := repo.Update(ctx, tenantID, id, input)
		require.NoError(t, err)
	})

	t.Run("update icon and name", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		icon := "new-icon"
		name := "New Name"
		input := domain.UpdateCategoryInput{
			Icon: &icon,
			Name: &name,
		}

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{
			ID:    id,
			Name:  "Old",
			Icon:  pgtype.Text{String: "old", Valid: true},
			Color: pgtype.Text{String: "color", Valid: true},
		}, nil)

		mockQuerier.On("UpdateCategory", ctx, mock.MatchedBy(func(p sqlc.UpdateCategoryParams) bool {
			return p.Icon.String == icon && p.Name == name && p.Color.String == "color"
		})).Return(sqlc.Category{ID: id}, nil)

		_, err := repo.Update(ctx, tenantID, id, input)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewCategoryRepository(mockQuerier)

		mockQuerier.On("GetCategoryByID", ctx, mock.Anything).Return(sqlc.Category{}, pgx.ErrNoRows)

		got, err := repo.Update(ctx, tenantID, id, domain.UpdateCategoryInput{})
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, got)
	})
}
