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

	svc.AssertExpectations(t)
}
