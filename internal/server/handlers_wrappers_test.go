package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/handler"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func makeServerWithTenantHandler(tsvc *mocks.TenantService) *Server {
	tokenParser := func(token string) (*paseto.Claims, error) {
		return &paseto.Claims{TenantID: "t1", UserID: "u1", Role: "owner"}, nil
	}
	rateLimiterStore := middleware.NewRateLimiterStore()
	idemp := new(mocks.IdempotencyStore)

	return &Server{
		tenantHandler:    handler.NewTenantHandler(tsvc),
		authHandler:      handler.NewAuthHandler(nil),
		tokenParser:      tokenParser,
		rateLimiterStore: rateLimiterStore,
		idempotencyStore: idemp,
	}
}

func TestServer_handleInviteUser_DelegatesToTenantHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.TenantService)
	s := makeServerWithTenantHandler(svc)

	reqBody := handler.InviteUserRequest{Email: "foo@example.com", Role: "member"}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	svc.On("InviteUser", mock.Anything, "t1", mock.Anything).Return(&domain.User{Email: reqBody.Email, Role: reqBody.Role}, nil)

	req := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(body))
	// add auth header so middleware would normally run (we call handler directly)
	req = req.WithContext(context.WithValue(req.Context(), middleware.TenantIDKey, "t1"))
	rr := httptest.NewRecorder()

	s.handleInviteUser(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	svc.AssertExpectations(t)
}

func TestServer_adminWrappers_DelegateToAdminHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.AdminService)
	s := &Server{adminHandler: handler.NewAdminHandler(svc)}

	// UpdateTenantPlan
	planReq := handler.UpdateTenantPlanRequest{Plan: "pro"}
	body, err := json.Marshal(planReq)
	require.NoError(t, err)
	svc.On("UpdateTenantPlan", mock.Anything, "t1", domain.TenantPlan("pro")).Return(&domain.Tenant{ID: "t1", Plan: "pro"}, nil)
	req := httptest.NewRequest(http.MethodPatch, "/v1/admin/tenants/t1/plan", bytes.NewReader(body))
	req.SetPathValue("id", "t1")
	rr := httptest.NewRecorder()
	s.handleAdminUpdatePlan(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// SuspendTenant
	svc.On("SuspendTenant", mock.Anything, "t1").Return(nil)
	req = httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/t1/suspend", nil)
	req.SetPathValue("id", "t1")
	rr = httptest.NewRecorder()
	s.handleAdminSuspendTenant(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	// RestoreTenant
	svc.On("RestoreTenant", mock.Anything, "t1").Return(nil)
	req = httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/t1/restore", nil)
	req.SetPathValue("id", "t1")
	rr = httptest.NewRecorder()
	s.handleAdminRestoreTenant(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	// HardDeleteTenant
	svc.On("HardDeleteTenant", mock.Anything, "t1", "t1").Return(nil)
	req = httptest.NewRequest(http.MethodDelete, "/v1/admin/tenants/t1", nil)
	req.SetPathValue("id", "t1")
	req.Header.Set("X-Confirm-Token", "t1")
	rr = httptest.NewRecorder()
	s.handleAdminHardDeleteTenant(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	// ForceDeleteUser
	svc.On("ForceDeleteUser", mock.Anything, "u1").Return(nil)
	req = httptest.NewRequest(http.MethodDelete, "/v1/admin/users/u1", nil)
	req.SetPathValue("id", "u1")
	rr = httptest.NewRecorder()
	s.handleAdminForceDeleteUser(rr, req)
	require.Equal(t, http.StatusNoContent, rr.Code)

	// Direct calls for remaining coverage (no mock setup needed as they will return 401/error without ctx)
	// Admin List Tenants
	t.Run("handleAdminListTenants", func(t *testing.T) {
		t.Parallel()
		req = httptest.NewRequest(http.MethodGet, "/v1/admin/tenants", nil)
		rr = httptest.NewRecorder()
		s.handleAdminListTenants(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	// Admin Get Tenant
	t.Run("handleAdminGetTenant", func(t *testing.T) {
		t.Parallel()
		req = httptest.NewRequest(http.MethodGet, "/v1/admin/tenants/t1", nil)
		req.SetPathValue("id", "t1")
		rr = httptest.NewRecorder()
		s.handleAdminGetTenant(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	// Admin List Users
	t.Run("handleAdminListUsers", func(t *testing.T) {
		t.Parallel()
		req = httptest.NewRequest(http.MethodGet, "/v1/admin/users", nil)
		rr = httptest.NewRecorder()
		s.handleAdminListUsers(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	// Admin List Audit Logs
	t.Run("handleAdminListAuditLogs", func(t *testing.T) {
		t.Parallel()
		req = httptest.NewRequest(http.MethodGet, "/v1/admin/audit-logs", nil)
		rr = httptest.NewRecorder()
		s.handleAdminListAuditLogs(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	svc.AssertExpectations(t)
}

func TestServer_TenantWrappers_DelegateToHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.TenantService)
	h := handler.NewTenantHandler(svc)
	s := &Server{tenantHandler: h}

	t.Run("handleGetTenantMe", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/me", nil)
		rr := httptest.NewRecorder()
		s.handleGetTenantMe(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handleUpdateTenantMe", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", nil)
		rr := httptest.NewRecorder()
		s.handleUpdateTenantMe(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestServer_AccountWrappers_DelegateToHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.AccountService)
	h := handler.NewAccountHandler(svc)
	s := &Server{accountHandler: h}

	wrappers := []struct {
		fn     func(http.ResponseWriter, *http.Request)
		name   string
		method string
		path   string
	}{
		{fn: s.handleListAccounts, name: "handleListAccounts", method: http.MethodGet, path: "/v1/accounts"},
		{fn: s.handleCreateAccount, name: "handleCreateAccount", method: http.MethodPost, path: "/v1/accounts"},
		{fn: s.handleGetAccountByID, name: "handleGetAccountByID", method: http.MethodGet, path: "/v1/accounts/1"},
		{fn: s.handleUpdateAccount, name: "handleUpdateAccount", method: http.MethodPatch, path: "/v1/accounts/1"},
		{fn: s.handleDeleteAccount, name: "handleDeleteAccount", method: http.MethodDelete, path: "/v1/accounts/1"},
	}

	for _, w := range wrappers {
		t.Run(w.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(w.method, w.path, nil)
			rr := httptest.NewRecorder()
			w.fn(rr, req)
			require.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestServer_CategoryWrappers_DelegateToHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.CategoryService)
	h := handler.NewCategoryHandler(svc)
	s := &Server{categoryHandler: h}

	wrappers := []struct {
		fn     func(http.ResponseWriter, *http.Request)
		name   string
		method string
		path   string
	}{
		{fn: s.handleListCategories, name: "handleListCategories", method: http.MethodGet, path: "/v1/categories"},
		{fn: s.handleCreateCategory, name: "handleCreateCategory", method: http.MethodPost, path: "/v1/categories"},
		{fn: s.handleGetCategoryByID, name: "handleGetCategoryByID", method: http.MethodGet, path: "/v1/categories/1"},
		{fn: s.handleUpdateCategory, name: "handleUpdateCategory", method: http.MethodPatch, path: "/v1/categories/1"},
		{fn: s.handleDeleteCategory, name: "handleDeleteCategory", method: http.MethodDelete, path: "/v1/categories/1"},
	}

	for _, w := range wrappers {
		t.Run(w.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(w.method, w.path, nil)
			rr := httptest.NewRecorder()
			w.fn(rr, req)
			require.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestServer_TransactionWrappers_DelegateToHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.TransactionService)
	h := handler.NewTransactionHandler(svc)
	s := &Server{transactionHandler: h}

	wrappers := []struct {
		fn     func(http.ResponseWriter, *http.Request)
		name   string
		method string
		path   string
	}{
		{fn: s.handleListTransactions, name: "handleListTransactions", method: http.MethodGet, path: "/v1/transactions"},
		{fn: s.handleCreateTransaction, name: "handleCreateTransaction", method: http.MethodPost, path: "/v1/transactions"},
		{fn: s.handleGetTransactionByID, name: "handleGetTransactionByID", method: http.MethodGet, path: "/v1/transactions/1"},
		{fn: s.handleUpdateTransaction, name: "handleUpdateTransaction", method: http.MethodPatch, path: "/v1/transactions/1"},
		{fn: s.handleDeleteTransaction, name: "handleDeleteTransaction", method: http.MethodDelete, path: "/v1/transactions/1"},
	}

	for _, w := range wrappers {
		t.Run(w.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(w.method, w.path, nil)
			rr := httptest.NewRecorder()
			w.fn(rr, req)
			require.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestServer_MasterPurchaseWrappers_DelegateToHandler(t *testing.T) {
	t.Parallel()
	svc := new(mocks.MasterPurchaseService)
	h := handler.NewMasterPurchaseHandler(svc)
	s := &Server{masterPurchaseHandler: h}

	t.Run("handleListMasterPurchases", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases", nil)
		rr := httptest.NewRecorder()
		s.handleListMasterPurchases(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handleCreateMasterPurchase", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", nil)
		rr := httptest.NewRecorder()
		s.handleCreateMasterPurchase(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handleGetMasterPurchaseByID", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/1", nil)
		rr := httptest.NewRecorder()
		s.handleGetMasterPurchaseByID(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handleProjectMasterPurchase", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/master-purchases/1/projection", nil)
		rr := httptest.NewRecorder()
		s.handleProjectMasterPurchase(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handleUpdateMasterPurchase", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPatch, "/v1/master-purchases/1", nil)
		rr := httptest.NewRecorder()
		s.handleUpdateMasterPurchase(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("handleDeleteMasterPurchase", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodDelete, "/v1/master-purchases/1", nil)
		rr := httptest.NewRecorder()
		s.handleDeleteMasterPurchase(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
