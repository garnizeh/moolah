package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMasterPurchaseHandler_Create(t *testing.T) {
	t.Parallel()

	tenantID := "tenant_123"
	userID := "user_456"
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)

		reqBody := CreateMasterPurchaseRequest{
			AccountID:            "acc_1",
			CategoryID:           "cat_1",
			Description:          "iPhone",
			TotalAmountCents:     120000,
			InstallmentCount:     12,
			ClosingDay:           10,
			FirstInstallmentDate: now,
		}

		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.MasterPurchase{ID: "mp_1"}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("validation failed", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)

		reqBody := CreateMasterPurchaseRequest{
			Description: "", // Invalid
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("unauthorized - missing tenant", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", nil)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", bytes.NewReader([]byte("invalid")))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		reqBody := CreateMasterPurchaseRequest{
			AccountID:            "acc_1",
			CategoryID:           "cat_1",
			Description:          "iPhone",
			TotalAmountCents:     120000,
			InstallmentCount:     12,
			ClosingDay:           10,
			FirstInstallmentDate: now,
		}
		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("db error"))
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestMasterPurchaseHandler_GetByID(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"
	mpID := "mp_1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("GetByID", mock.Anything, tenantID, mpID).Return(&domain.MasterPurchase{ID: mpID}, nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/"+mpID, nil)
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/"+mpID, nil)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/", nil)
		// No SetPathValue
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("GetByID", mock.Anything, tenantID, mpID).Return(nil, domain.ErrNotFound)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/"+mpID, nil)
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestMasterPurchaseHandler_ListByTenant(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"

	t.Run("success list all", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("ListByTenant", mock.Anything, tenantID).Return([]domain.MasterPurchase{{ID: "1"}}, nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.ListByTenant(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("success filter by account", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("ListByAccount", mock.Anything, tenantID, "acc_1").Return([]domain.MasterPurchase{{ID: "1"}}, nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases?account_id=acc_1", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.ListByTenant(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases", nil)
		rr := httptest.NewRecorder()
		h.ListByTenant(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("ListByTenant", mock.Anything, tenantID).Return(nil, errors.New("db error"))
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.ListByTenant(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestMasterPurchaseHandler_Project(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"
	mpID := "mp_1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		mp := &domain.MasterPurchase{ID: mpID}
		svc.On("GetByID", mock.Anything, tenantID, mpID).Return(mp, nil)
		svc.On("ProjectInstallments", mp).Return([]domain.ProjectedInstallment{{AmountCents: 100}})
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/"+mpID+"/projection", nil)
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Project(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("service error get", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("GetByID", mock.Anything, tenantID, mpID).Return(nil, domain.ErrNotFound)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/"+mpID+"/projection", nil)
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Project(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/"+mpID+"/projection", nil)
		rr := httptest.NewRecorder()
		h.Project(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestMasterPurchaseHandler_Update(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"
	mpID := "mp_1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		desc := "New Desc"
		svc.On("Update", mock.Anything, tenantID, mpID, mock.Anything).Return(&domain.MasterPurchase{ID: mpID}, nil)
		body, _ := json.Marshal(UpdateMasterPurchaseRequest{Description: &desc})
		req := httptest.NewRequest(http.MethodPatch, "/v1/master-purchases/"+mpID, bytes.NewReader(body))
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodPatch, "/v1/master-purchases/"+mpID, nil)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodPatch, "/v1/master-purchases/"+mpID, bytes.NewReader([]byte("invalid")))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validation failed", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		invalidDesc := ""
		body, _ := json.Marshal(UpdateMasterPurchaseRequest{Description: &invalidDesc})
		req := httptest.NewRequest(http.MethodPatch, "/v1/master-purchases/"+mpID, bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("Update", mock.Anything, tenantID, mpID, mock.Anything).Return(nil, errors.New("db error"))
		body, _ := json.Marshal(UpdateMasterPurchaseRequest{})
		req := httptest.NewRequest(http.MethodPatch, "/v1/master-purchases/"+mpID, bytes.NewReader(body))
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestMasterPurchaseHandler_Delete(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"
	mpID := "mp_1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("Delete", mock.Anything, tenantID, mpID).Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/v1/master-purchases/"+mpID, nil)
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)
		req := httptest.NewRequest(http.MethodDelete, "/v1/master-purchases/"+mpID, nil)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)
		svc.On("Delete", mock.Anything, tenantID, mpID).Return(domain.ErrNotFound)
		req := httptest.NewRequest(http.MethodDelete, "/v1/master-purchases/"+mpID, nil)
		req.SetPathValue("id", mpID)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
