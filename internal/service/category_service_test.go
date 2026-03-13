package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCategoryService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	input := domain.CreateCategoryInput{
		Name: "Groceries",
		Type: domain.CategoryTypeExpense,
		Icon: "🛒",
	}

	t.Run("Success_Root", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		category := &domain.Category{ID: "cat_1", TenantID: tenantID, Name: input.Name, Type: input.Type}
		categoryRepo.On("Create", ctx, tenantID, input).Return(category, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
		require.Equal(t, category, res)
	})

	t.Run("AuditError_Create", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		category := &domain.Category{ID: "cat_1"}
		categoryRepo.On("Create", ctx, tenantID, input).Return(category, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, input)

		require.NoError(t, err)
		require.Equal(t, category, res)
	})

	t.Run("RepoError_Create", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		categoryRepo.On("Create", ctx, tenantID, input).Return(nil, errors.New("db error"))

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Success_Child", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		parentID := "parent_1"
		childInput := input
		childInput.ParentID = parentID

		parent := &domain.Category{ID: parentID, ParentID: ""}
		category := &domain.Category{ID: "cat_2", ParentID: parentID}

		categoryRepo.On("GetByID", ctx, tenantID, parentID).Return(parent, nil)
		categoryRepo.On("Create", ctx, tenantID, childInput).Return(category, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		res, err := svc.Create(ctx, tenantID, childInput)

		require.NoError(t, err)
		require.Equal(t, category, res)
	})

	t.Run("Error_Grandchild", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)

		parentID := "parent_1"
		childInput := input
		childInput.ParentID = parentID

		parentIsAChild := &domain.Category{ID: parentID, ParentID: "grandparent_1"}

		categoryRepo.On("GetByID", ctx, tenantID, parentID).Return(parentIsAChild, nil)

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, childInput)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
		require.Nil(t, res)
	})

	t.Run("Error_ParentLookupFail", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)

		parentID := "parent_1"
		childInput := input
		childInput.ParentID = parentID

		categoryRepo.On("GetByID", ctx, tenantID, parentID).Return(nil, errors.New("parent db error"))

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, childInput)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to fetch parent category")
		require.Nil(t, res)
	})

	t.Run("Error_ParentNotFound", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)

		parentID := "ghost_parent"
		inputWithGhost := input
		inputWithGhost.ParentID = parentID

		categoryRepo.On("GetByID", ctx, tenantID, parentID).Return(nil, nil)

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.Create(ctx, tenantID, inputWithGhost)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Contains(t, err.Error(), "parent category not found")
		require.Nil(t, res)
	})
}

func TestCategoryService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	catID := "cat_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		category := &domain.Category{ID: catID, TenantID: tenantID}

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(category, nil)

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.GetByID(ctx, tenantID, catID)

		require.NoError(t, err)
		require.Equal(t, category, res)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(nil, domain.ErrNotFound)

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.GetByID(ctx, tenantID, catID)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, res)
	})
}

func TestCategoryService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		categories := []domain.Category{{ID: "cat_1"}}
		categoryRepo.On("ListByTenant", ctx, tenantID).Return(categories, nil)

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.ListByTenant(ctx, tenantID)

		require.NoError(t, err)
		require.Equal(t, categories, res)
	})

	t.Run("ListByTenant_Error", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		categoryRepo.On("ListByTenant", ctx, tenantID).Return(([]domain.Category)(nil), errors.New("db error"))

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.ListByTenant(ctx, tenantID)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("ListChildren", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		parentID := "parent_1"
		categories := []domain.Category{{ID: "cat_1", ParentID: parentID}}
		categoryRepo.On("ListChildren", ctx, tenantID, parentID).Return(categories, nil)

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.ListChildren(ctx, tenantID, parentID)

		require.NoError(t, err)
		require.Equal(t, categories, res)
	})

	t.Run("ListChildren_Error", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		parentID := "parent_1"
		categoryRepo.On("ListChildren", ctx, tenantID, parentID).Return(([]domain.Category)(nil), errors.New("db error"))

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.ListChildren(ctx, tenantID, parentID)

		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestCategoryService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	catID := "cat_1"
	newName := "New Name"
	input := domain.UpdateCategoryInput{Name: &newName}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		oldCat := &domain.Category{ID: catID, Name: "Old Name"}
		newCat := &domain.Category{ID: catID, Name: newName}

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(oldCat, nil)
		categoryRepo.On("Update", ctx, tenantID, catID, input).Return(newCat, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		res, err := svc.Update(ctx, tenantID, catID, input)

		require.NoError(t, err)
		require.Equal(t, newCat, res)
	})

	t.Run("Success_Complex", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		oldCat := &domain.Category{ID: catID, Name: "Old Name", Icon: "old-icon", Color: "old-color"}
		newCategoryName := "New Name"
		newIcon := "new-icon"
		newColor := "new-color"
		inputComplex := domain.UpdateCategoryInput{Name: &newCategoryName, Icon: &newIcon, Color: &newColor}
		newCat := &domain.Category{ID: catID, Name: newCategoryName, Icon: newIcon, Color: newColor}

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(oldCat, nil)
		categoryRepo.On("Update", ctx, tenantID, catID, inputComplex).Return(newCat, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		res, err := svc.Update(ctx, tenantID, catID, inputComplex)

		require.NoError(t, err)
		require.Equal(t, newCat, res)
	})

	t.Run("UpdateError", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		oldCat := &domain.Category{ID: catID, Name: "Old Name"}

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(oldCat, nil)
		categoryRepo.On("Update", ctx, tenantID, catID, input).Return((*domain.Category)(nil), errors.New("db error"))

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.Update(ctx, tenantID, catID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Update_FetchError", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(nil, errors.New("db error"))

		svc := service.NewCategoryService(categoryRepo, nil)
		res, err := svc.Update(ctx, tenantID, catID, input)

		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("Update_AuditError", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		oldCat := &domain.Category{ID: catID, Name: "Old Name"}
		newCat := &domain.Category{ID: catID, Name: newName}

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(oldCat, nil)
		categoryRepo.On("Update", ctx, tenantID, catID, input).Return(newCat, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		res, err := svc.Update(ctx, tenantID, catID, input)

		require.NoError(t, err)
		require.Equal(t, newCat, res)
	})
}

func TestCategoryService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	catID := "cat_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(&domain.Category{ID: catID}, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)
		categoryRepo.On("Delete", ctx, tenantID, catID).Return(nil)

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		err := svc.Delete(ctx, tenantID, catID)

		require.NoError(t, err)
	})

	t.Run("AuditError_Delete", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(&domain.Category{ID: catID}, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit fail"))
		categoryRepo.On("Delete", ctx, tenantID, catID).Return(nil)

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		err := svc.Delete(ctx, tenantID, catID)

		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(nil, domain.ErrNotFound)

		svc := service.NewCategoryService(categoryRepo, nil)
		err := svc.Delete(ctx, tenantID, catID)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("DeleteError", func(t *testing.T) {
		t.Parallel()
		categoryRepo := new(mocks.CategoryRepository)
		auditRepo := new(mocks.AuditRepository)

		categoryRepo.On("GetByID", ctx, tenantID, catID).Return(&domain.Category{ID: catID}, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(&domain.AuditLog{}, nil)
		categoryRepo.On("Delete", ctx, tenantID, catID).Return(errors.New("db error"))

		svc := service.NewCategoryService(categoryRepo, auditRepo)
		err := svc.Delete(ctx, tenantID, catID)

		require.Error(t, err)
	})
}
