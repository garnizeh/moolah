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
)

func TestTransactionHandler_Create(t *testing.T) {
	t.Parallel()

	tenantID := "tenant_123"
	userID := "user_456"
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)

		reqBody := CreateTransactionRequest{
			OccurredAt:  now,
			AccountID:   "acc_1",
			CategoryID:  "cat_1",
			Description: "coffee",
			Type:        domain.TransactionTypeExpense,
			AmountCents: 500,
		}

		expectedTx := &domain.Transaction{ID: "tx_1"}
		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(expectedTx, nil)

		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("unauthorized - missing tenant", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", nil)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("unauthorized - missing user", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader([]byte("invalid")))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validation failed", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		reqBody := CreateTransactionRequest{AmountCents: -1}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrForbidden)
		body, err := json.Marshal(CreateTransactionRequest{OccurredAt: now, AccountID: "a", CategoryID: "c", Description: "d", Type: "income", AmountCents: 1})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrNotFound)
		body, err := json.Marshal(CreateTransactionRequest{OccurredAt: now, AccountID: "a", CategoryID: "c", Description: "d", Type: "income", AmountCents: 1})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrInvalidInput)
		body, err := json.Marshal(CreateTransactionRequest{OccurredAt: now, AccountID: "a", CategoryID: "c", Description: "d", Type: "income", AmountCents: 1})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("boom"))
		body, err := json.Marshal(CreateTransactionRequest{OccurredAt: now, AccountID: "a", CategoryID: "c", Description: "d", Type: "income", AmountCents: 1})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Create(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTransactionHandler_GetByID(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("GetByID", mock.Anything, tenantID, "tx_1").Return(&domain.Transaction{ID: "tx_1"}, nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/tx_1", nil)
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/tx_1", nil)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("GetByID", mock.Anything, tenantID, "tx_1").Return(nil, domain.ErrNotFound)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/tx_1", nil)
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("GetByID", mock.Anything, tenantID, "tx_1").Return(nil, errors.New("boom"))
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions/tx_1", nil)
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.GetByID(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTransactionHandler_Update(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Update", mock.Anything, tenantID, "tx_1", mock.Anything).Return(&domain.Transaction{ID: "tx_1"}, nil)
		body, err := json.Marshal(UpdateTransactionRequest{})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader(body))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", nil)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader([]byte("invalid")))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validation failed", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		desc := ""
		body, err := json.Marshal(UpdateTransactionRequest{Description: &desc})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader(body))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Update", mock.Anything, tenantID, "tx_1", mock.Anything).Return(nil, domain.ErrNotFound)
		body, err := json.Marshal(UpdateTransactionRequest{})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader(body))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Update", mock.Anything, tenantID, "tx_1", mock.Anything).Return(nil, domain.ErrInvalidInput)
		body, err := json.Marshal(UpdateTransactionRequest{})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader(body))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Update", mock.Anything, tenantID, "tx_1", mock.Anything).Return(nil, domain.ErrForbidden)
		body, err := json.Marshal(UpdateTransactionRequest{})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader(body))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Update", mock.Anything, tenantID, "tx_1", mock.Anything).Return(nil, errors.New("boom"))
		body, err := json.Marshal(UpdateTransactionRequest{})
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPatch, "/v1/transactions/tx_1", bytes.NewReader(body))
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Update(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTransactionHandler_Delete(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Delete", mock.Anything, tenantID, "tx_1").Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/tx_1", nil)
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/tx_1", nil)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Delete", mock.Anything, tenantID, "tx_1").Return(domain.ErrNotFound)
		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/tx_1", nil)
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("Delete", mock.Anything, tenantID, "tx_1").Return(errors.New("boom"))
		req := httptest.NewRequest(http.MethodDelete, "/v1/transactions/tx_1", nil)
		req.SetPathValue("id", "tx_1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.Delete(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTransactionHandler_List(t *testing.T) {
	t.Parallel()
	tenantID := "tenant_123"

	t.Run("success with all params", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("List", mock.Anything, tenantID, mock.Anything).Return([]domain.Transaction{}, nil)
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions?account_id=a&category_id=c&type=income&start_date=2024-01-01T00:00:00Z&end_date=2024-01-02T00:00:00Z&limit=10&offset=5", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		h := NewTransactionHandler(&mocks.TransactionService{})
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.TransactionService{}
		h := NewTransactionHandler(svc)
		svc.On("List", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("boom"))
		req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.List(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
