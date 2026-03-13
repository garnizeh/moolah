package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountHandler_List(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "tenant-1"
		accounts := []domain.Account{{ID: "acc-1", Name: "Checking"}}
		service.On("ListByTenant", mock.Anything, tenantID).Return(accounts, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.Account
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp, 1)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		req := httptest.NewRequest(http.MethodGet, "/v1/accounts", nil)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "tenant-1"
		service.On("ListByTenant", mock.Anything, tenantID).Return(nil, errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		req := httptest.NewRequest(http.MethodDelete, "/v1/accounts/a1", nil)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing_id", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodDelete, "/v1/accounts/", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAccountHandler_Create(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "tenant-1"
		userID := "user-1"
		reqBody := CreateAccountRequest{
			Name:         "Savings",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 1000,
		}

		service.On("Create", mock.Anything, tenantID, mock.MatchedBy(func(in domain.CreateAccountInput) bool {
			return in.Name == reqBody.Name && in.UserID == userID
		})).Return(&domain.Account{Name: reqBody.Name}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		service.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrConflict)

		body, err := json.Marshal(CreateAccountRequest{
			Name:         "Savings Account",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 1000,
		})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, "u1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		service.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrForbidden)

		body, err := json.Marshal(CreateAccountRequest{
			Name:         "Savings Account",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 1000,
		})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, "u1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		service.On("Create", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("boom"))

		body, err := json.Marshal(CreateAccountRequest{
			Name:         "Savings Account",
			Type:         domain.AccountTypeSavings,
			Currency:     "USD",
			InitialCents: 1000,
		})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, "u1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("invalid_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		ctx = context.WithValue(ctx, middleware.UserIDKey, "u1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", bytes.NewReader([]byte("invalid"))).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		reqBody := CreateAccountRequest{Name: ""} // Invalid
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		ctx = context.WithValue(ctx, middleware.UserIDKey, "u1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestAccountHandler_GetByID(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		service.On("GetByID", mock.Anything, tenantID, accID).Return(&domain.Account{ID: accID}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		service.On("GetByID", mock.Anything, "t1", "a1").Return(nil, domain.ErrNotFound)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})
	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		service.On("GetByID", mock.Anything, "t1", "a1").Return(nil, errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/a1", nil)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing_id", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/", nil).WithContext(ctx)
		// No PathValue set
		rr := httptest.NewRecorder()

		h.GetByID(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAccountHandler_Update(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		name := "New Name"
		service.On("Update", mock.Anything, tenantID, accID, mock.MatchedBy(func(in domain.UpdateAccountInput) bool {
			return *in.Name == name
		})).Return(&domain.Account{Name: name}, nil)

		body, err := json.Marshal(UpdateAccountRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		service.On("Update", mock.Anything, tenantID, accID, mock.Anything).Return(nil, domain.ErrNotFound)

		name := "New Name"
		body, err := json.Marshal(UpdateAccountRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		service.On("Update", mock.Anything, tenantID, accID, mock.Anything).Return(nil, domain.ErrForbidden)

		name := "New Name"
		body, err := json.Marshal(UpdateAccountRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		service.On("Update", mock.Anything, tenantID, accID, mock.Anything).Return(nil, domain.ErrConflict)

		name := "New Name"
		body, err := json.Marshal(UpdateAccountRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		service.On("Update", mock.Anything, tenantID, accID, mock.Anything).Return(nil, errors.New("boom"))

		name := "New Name"
		body, err := json.Marshal(UpdateAccountRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		req := httptest.NewRequest(http.MethodPost, "/v1/accounts", nil)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestAccountHandler_Delete(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "t1"
		accID := "a1"
		service.On("Delete", mock.Anything, tenantID, accID).Return(nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodDelete, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", accID)
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		service.On("Delete", mock.Anything, "t1", "a1").Return(domain.ErrNotFound)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodDelete, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		service.On("Delete", mock.Anything, "t1", "a1").Return(domain.ErrForbidden)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodDelete, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		service.On("Delete", mock.Anything, "t1", "a1").Return(errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodDelete, "/v1/accounts/a1", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.Delete(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", nil)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("missing_id", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader([]byte("invalid"))).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		name := ""
		body, err := json.Marshal(UpdateAccountRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPatch, "/v1/accounts/a1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestAccountHandler_CloseInvoice(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		tenantID := "tenant-1"
		accountID := "acc-1"
		closingDateStr := "2026-03-12"

		closer.On("CloseInvoice", mock.Anything, tenantID, accountID, mock.MatchedBy(func(t time.Time) bool {
			return t.Format("2006-01-02") == closingDateStr
		})).Return(domain.CloseInvoiceResult{ProcessedCount: 3}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		body := bytes.NewReader([]byte(`{"closing_date": "2026-03-12"}`))
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/acc-1/close-invoice", body).WithContext(ctx)
		req.SetPathValue("id", accountID)
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp CloseInvoiceResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, accountID, resp.AccountID)
		require.Equal(t, 3, resp.ProcessedCount)
	})

	t.Run("success_default_date", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		closer.On("CloseInvoice", mock.Anything, "t1", "a1", mock.Anything).Return(domain.CloseInvoiceResult{ProcessedCount: 0}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("account_not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		closer.On("CloseInvoice", mock.Anything, "t1", "a1", mock.Anything).Return(domain.CloseInvoiceResult{}, domain.ErrNotFound)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid_account_type", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		closer.On("CloseInvoice", mock.Anything, "t1", "a1", mock.Anything).Return(domain.CloseInvoiceResult{}, domain.ErrInvalidInput)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("partial_failures", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.AccountService)
		closer := new(mocks.InvoiceCloser)
		h := NewAccountHandler(service, closer)

		closer.On("CloseInvoice", mock.Anything, "t1", "a1", mock.Anything).Return(domain.CloseInvoiceResult{
			ProcessedCount: 1,
			Errors:         []error{errors.New("MP2 failed")},
		}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", nil).WithContext(ctx)
		req.SetPathValue("id", "a1")
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp CloseInvoiceResponse
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp.Errors, 1)
		require.Equal(t, "MP2 failed", resp.Errors[0])
	})

	t.Run("unauthorized_missing_tenant_id", func(t *testing.T) {
		t.Parallel()
		h := NewAccountHandler(new(mocks.AccountService), new(mocks.InvoiceCloser))

		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", nil)
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
		require.Contains(t, rr.Body.String(), "unauthorized")
	})

	t.Run("missing_account_id_in_path", func(t *testing.T) {
		t.Parallel()
		h := NewAccountHandler(new(mocks.AccountService), new(mocks.InvoiceCloser))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts//close-invoice", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "missing account id")
	})

	t.Run("validation_failure_invalid_date", func(t *testing.T) {
		t.Parallel()
		h := NewAccountHandler(new(mocks.AccountService), new(mocks.InvoiceCloser))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		body := bytes.NewReader([]byte(`{"closing_date": "not-a-date"}`))
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", body).WithContext(ctx)
		req.SetPathValue("id", "a1")
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("invalid_closing_date_parse_failure", func(t *testing.T) {
		t.Parallel()
		h := NewAccountHandler(new(mocks.AccountService), new(mocks.InvoiceCloser))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "t1")
		// "2026-02-31" passes the datetime=2006-01-02 validator in some versions/cases
		// but time.Parse("2006-01-02", ...) correctly fails on non-existent days.
		date := "2026-02-31"
		body := bytes.NewReader([]byte(`{"closing_date": "` + date + `"}`))
		req := httptest.NewRequest(http.MethodPost, "/v1/accounts/a1/close-invoice", body).WithContext(ctx)
		req.SetPathValue("id", "a1")
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.CloseInvoice(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		require.Contains(t, rr.Body.String(), "ClosingDate")
	})
}
