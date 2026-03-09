package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCategoryHandler_List(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categories := []domain.Category{
			{ID: "c1", Name: "Food", Type: domain.CategoryTypeExpense},
			{ID: "c2", Name: "Salary", Type: domain.CategoryTypeIncome},
		}

		service.On("ListByTenant", mock.Anything, tenantID).Return(categories, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.Category
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("filter_by_type", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categories := []domain.Category{
			{ID: "c1", Name: "Food", Type: domain.CategoryTypeExpense},
			{ID: "c2", Name: "Salary", Type: domain.CategoryTypeIncome},
		}

		service.On("ListByTenant", mock.Anything, tenantID).Return(categories, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories?type=income", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.Category
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, domain.CategoryTypeIncome, resp[0].Type)
	})

	t.Run("list_children", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		parentID := "p1"
		children := []domain.Category{
			{ID: "c1", Name: "Child", ParentID: parentID},
		}

		service.On("ListChildren", mock.Anything, tenantID, parentID).Return(children, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories?parent_id="+parentID, nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.Category
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp, 1)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		service.On("ListByTenant", mock.Anything, tenantID).Return(nil, errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestCategoryHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		reqBody := CreateCategoryRequest{
			Name: "Food",
			Type: domain.CategoryTypeExpense,
		}

		service.On("Create", mock.Anything, tenantID, mock.MatchedBy(func(in domain.CreateCategoryInput) bool {
			return in.Name == reqBody.Name
		})).Return(&domain.Category{ID: "c1", Name: reqBody.Name}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("hierarchy_violation", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		parentID := "p1"
		reqBody := CreateCategoryRequest{
			Name:     "Child",
			Type:     domain.CategoryTypeExpense,
			ParentID: &parentID,
		}

		service.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrInvalidInput)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		reqBody := CreateCategoryRequest{
			Name: "Food",
			Type: domain.CategoryTypeExpense,
		}

		service.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrConflict)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		reqBody := CreateCategoryRequest{Name: ""} // name is required
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		reqBody := CreateCategoryRequest{
			Name: "Food",
			Type: domain.CategoryTypeExpense,
		}

		service.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("boom"))

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		reqBody := CreateCategoryRequest{Name: "Food", Type: domain.CategoryTypeExpense}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewReader([]byte("invalid"))).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestCategoryHandler_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		service.On("GetByID", mock.Anything, tenantID, categoryID).Return(&domain.Category{ID: categoryID}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories/"+categoryID, nil).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		service.On("GetByID", mock.Anything, tenantID, categoryID).Return(nil, domain.ErrNotFound)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories/"+categoryID, nil).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		service.On("GetByID", mock.Anything, tenantID, categoryID).Return(nil, errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories/"+categoryID, nil).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/v1/categories/c1", nil)
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing_id", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodGet, "/v1/categories/", nil).WithContext(ctx)
		// No PathValue set
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestCategoryHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		newName := "New Food"
		reqBody := UpdateCategoryRequest{Name: &newName}

		service.On("Update", mock.Anything, tenantID, categoryID, mock.MatchedBy(func(in domain.UpdateCategoryInput) bool {
			return *in.Name == newName
		})).Return(&domain.Category{ID: categoryID, Name: newName}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/"+categoryID, bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		name := "New Name"
		reqBody := UpdateCategoryRequest{Name: &name}

		service.On("Update", mock.Anything, tenantID, categoryID, mock.Anything).Return(nil, domain.ErrNotFound)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/"+categoryID, bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		name := "Duplicate"
		reqBody := UpdateCategoryRequest{Name: &name}

		service.On("Update", mock.Anything, tenantID, categoryID, mock.Anything).Return(nil, domain.ErrConflict)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/"+categoryID, bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		name := "Name"
		reqBody := UpdateCategoryRequest{Name: &name}

		service.On("Update", mock.Anything, tenantID, categoryID, mock.Anything).Return(nil, errors.New("boom"))

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/"+categoryID, bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		reqBody := UpdateCategoryRequest{}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/c1", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing_id", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		reqBody := UpdateCategoryRequest{}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPatch, "/v1/categories/c1", bytes.NewReader([]byte("invalid"))).WithContext(ctx)
		req.SetPathValue("id", "c1")
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestCategoryHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		service.On("Delete", mock.Anything, tenantID, categoryID).Return(nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodDelete, "/v1/categories/"+categoryID, nil).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		service.On("Delete", mock.Anything, tenantID, categoryID).Return(domain.ErrNotFound)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodDelete, "/v1/categories/"+categoryID, nil).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		tenantID := "t1"
		categoryID := "c1"
		service.On("Delete", mock.Anything, tenantID, categoryID).Return(errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodDelete, "/v1/categories/"+categoryID, nil).WithContext(ctx)
		req.SetPathValue("id", categoryID)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		req := httptest.NewRequest(http.MethodDelete, "/v1/categories/c1", nil)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing_id", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.CategoryService)
		h := NewCategoryHandler(service)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodDelete, "/v1/categories/", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
