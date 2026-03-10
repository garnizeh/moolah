package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminHandler_Tenants(t *testing.T) {
	t.Parallel()

	t.Run("ListTenants success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		expected := []domain.Tenant{{ID: "t1"}}

		svc.On("ListAllTenants", mock.Anything, false).Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/tenants", nil)
		rr := httptest.NewRecorder()
		h.ListTenants(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("ListTenants internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("ListAllTenants", mock.Anything, true).Return(([]domain.Tenant)(nil), errors.New("internal error"))

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/tenants?with_deleted=true", nil)
		rr := httptest.NewRecorder()
		h.ListTenants(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("GetTenant success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		expected := &domain.Tenant{ID: "t1"}

		svc.On("GetTenantByID", mock.Anything, "t1").Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/tenants/t1", nil)
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.GetTenant(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GetTenant not found", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("GetTenantByID", mock.Anything, "t1").Return((*domain.Tenant)(nil), domain.ErrNotFound)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/tenants/t1", nil)
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.GetTenant(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("UpdateTenantPlan success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		reqBody := UpdateTenantPlanRequest{Plan: "pro"}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		svc.On("UpdateTenantPlan", mock.Anything, "t1", domain.TenantPlan("pro")).Return(&domain.Tenant{ID: "t1", Plan: "pro"}, nil)

		req := httptest.NewRequest(http.MethodPatch, "/v1/admin/tenants/t1/plan", bytes.NewReader(body))
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.UpdateTenantPlan(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("UpdateTenantPlan internal error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		reqBody := UpdateTenantPlanRequest{Plan: "pro"}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		svc.On("UpdateTenantPlan", mock.Anything, "t1", domain.TenantPlan("pro")).Return((*domain.Tenant)(nil), errors.New("internal error"))

		req := httptest.NewRequest(http.MethodPatch, "/v1/admin/tenants/t1/plan", bytes.NewReader(body))
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.UpdateTenantPlan(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("UpdateTenantPlan invalid body", func(t *testing.T) {
		t.Parallel()
		h := NewAdminHandler(&mocks.AdminService{})
		req := httptest.NewRequest(http.MethodPatch, "/v1/admin/tenants/t1/plan", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()
		h.UpdateTenantPlan(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UpdateTenantPlan validation failed", func(t *testing.T) {
		t.Parallel()
		h := NewAdminHandler(&mocks.AdminService{})
		reqBody := UpdateTenantPlanRequest{Plan: "invalid"}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPatch, "/v1/admin/tenants/t1/plan", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		h.UpdateTenantPlan(rr, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("SuspendTenant success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("SuspendTenant", mock.Anything, "t1").Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/t1/suspend", nil)
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.SuspendTenant(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("SuspendTenant error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("SuspendTenant", mock.Anything, "t1").Return(errors.New("internal error"))

		req := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/t1/suspend", nil)
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.SuspendTenant(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("RestoreTenant success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("RestoreTenant", mock.Anything, "t1").Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/t1/restore", nil)
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.RestoreTenant(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("RestoreTenant error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("RestoreTenant", mock.Anything, "t1").Return(errors.New("internal error"))

		req := httptest.NewRequest(http.MethodPost, "/v1/admin/tenants/t1/restore", nil)
		req.SetPathValue("id", "t1")
		rr := httptest.NewRecorder()
		h.RestoreTenant(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("HardDeleteTenant success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("HardDeleteTenant", mock.Anything, "t1", "t1").Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/tenants/t1", nil)
		req.SetPathValue("id", "t1")
		req.Header.Set("X-Confirm-Token", "t1")
		rr := httptest.NewRecorder()
		h.HardDeleteTenant(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("HardDeleteTenant error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("HardDeleteTenant", mock.Anything, "t1", "t1").Return(errors.New("internal error"))

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/tenants/t1", nil)
		req.SetPathValue("id", "t1")
		req.Header.Set("X-Confirm-Token", "t1")
		rr := httptest.NewRecorder()
		h.HardDeleteTenant(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("HardDeleteTenant token mismatch", func(t *testing.T) {
		t.Parallel()
		h := NewAdminHandler(&mocks.AdminService{})

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/tenants/t1", nil)
		req.SetPathValue("id", "t1")
		req.Header.Set("X-Confirm-Token", "wrong")
		rr := httptest.NewRecorder()
		h.HardDeleteTenant(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAdminHandler_Users(t *testing.T) {
	t.Parallel()

	t.Run("ListUsers success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		expected := []domain.User{{ID: "u1"}}

		svc.On("ListAllUsers", mock.Anything).Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/users", nil)
		rr := httptest.NewRecorder()
		h.ListUsers(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ListUsers error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("ListAllUsers", mock.Anything).Return(([]domain.User)(nil), errors.New("internal error"))

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/users", nil)
		rr := httptest.NewRecorder()
		h.ListUsers(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("GetUser success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		expected := &domain.User{ID: "u1"}

		svc.On("GetUserByID", mock.Anything, "u1").Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/users/u1", nil)
		req.SetPathValue("id", "u1")
		rr := httptest.NewRecorder()
		h.GetUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("GetUser error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("GetUserByID", mock.Anything, "u1").Return((*domain.User)(nil), errors.New("internal error"))

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/users/u1", nil)
		req.SetPathValue("id", "u1")
		rr := httptest.NewRecorder()
		h.GetUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("ForceDeleteUser success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("ForceDeleteUser", mock.Anything, "u1").Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/users/u1", nil)
		req.SetPathValue("id", "u1")
		rr := httptest.NewRecorder()
		h.ForceDeleteUser(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("ForceDeleteUser error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("ForceDeleteUser", mock.Anything, "u1").Return(errors.New("internal error"))

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/users/u1", nil)
		req.SetPathValue("id", "u1")
		rr := httptest.NewRecorder()
		h.ForceDeleteUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestAdminHandler_AuditLogs(t *testing.T) {
	t.Parallel()

	t.Run("ListAuditLogs success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)
		expected := []domain.AuditLog{{ID: "l1"}}

		svc.On("ListAuditLogs", mock.Anything, mock.MatchedBy(func(p domain.ListAuditLogsParams) bool {
			return p.Limit == 10 && p.Offset == 5
		})).Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit-logs?limit=10&offset=5", nil)
		rr := httptest.NewRecorder()
		h.ListAuditLogs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ListAuditLogs error", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.AdminService{}
		h := NewAdminHandler(svc)

		svc.On("ListAuditLogs", mock.Anything, mock.Anything).Return(([]domain.AuditLog)(nil), errors.New("internal error"))

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/audit-logs", nil)
		rr := httptest.NewRecorder()
		h.ListAuditLogs(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
