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

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTenantHandler_GetMe(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		tenant := &domain.Tenant{ID: tenantID, Name: "Household"}

		service.On("GetByID", mock.Anything, tenantID).Return(tenant, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetMe(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp domain.Tenant
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, tenant.ID, resp.ID)
		service.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "non-existent"
		service.On("GetByID", mock.Anything, tenantID).Return(nil, domain.ErrNotFound)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetMe(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/me", nil)
		rr := httptest.NewRecorder()

		h.GetMe(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		service.On("GetByID", mock.Anything, tenantID).Return(nil, errors.New("boom"))

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/me", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetMe(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestTenantHandler_UpdateMe(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		name := "New Name"
		tenant := &domain.Tenant{ID: tenantID, Name: name}

		service.On("Update", mock.Anything, tenantID, domain.UpdateTenantInput{Name: &name}).Return(tenant, nil)

		reqBody, err := json.Marshal(UpdateTenantRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp domain.Tenant
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, name, resp.Name)
	})

	t.Run("invalid_request", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		name := "n" // too short
		reqBody, err := json.Marshal(UpdateTenantRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		name := "New Name"
		service.On("Update", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("boom"))

		reqBody, err := json.Marshal(UpdateTenantRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("invalid_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", bytes.NewReader([]byte("invalid"))).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		name := "New Name"
		service.On("Update", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrNotFound)

		reqBody, err := json.Marshal(UpdateTenantRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		name := "New Name"
		service.On("Update", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrForbidden)

		reqBody, err := json.Marshal(UpdateTenantRequest{Name: &name})
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		req := httptest.NewRequest(http.MethodPatch, "/v1/tenants/me", nil)
		rr := httptest.NewRecorder()

		h.UpdateMe(rr, req)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestTenantHandler_InviteUser(t *testing.T) {
	t.Parallel()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		req := InviteUserRequest{Email: "new@example.com", Role: domain.RoleMember}
		user := &domain.User{Email: req.Email, Role: req.Role}

		service.On("InviteUser", mock.Anything, tenantID, mock.MatchedBy(func(input domain.CreateUserInput) bool {
			return input.Email == req.Email && input.Role == req.Role
		})).Return(user, nil)

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusCreated, rr.Code)
		var resp domain.User
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, req.Email, resp.Email)
	})

	t.Run("conflict", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		req := InviteUserRequest{Email: "exists@example.com", Role: domain.RoleMember}
		service.On("InviteUser", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrConflict)

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		req := InviteUserRequest{Email: "error@example.com", Role: domain.RoleMember}
		service.On("InviteUser", mock.Anything, tenantID, mock.Anything).Return(nil, errors.New("boom"))

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		req := InviteUserRequest{Email: "forbidden@example.com", Role: domain.RoleMember}
		service.On("InviteUser", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrForbidden)

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("tenant_not_found", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		req := InviteUserRequest{Email: "new@example.com", Role: domain.RoleMember}
		service.On("InviteUser", mock.Anything, tenantID, mock.Anything).Return(nil, domain.ErrNotFound)

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid_request", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		req := InviteUserRequest{Email: "invalid-email", Role: "invalid-role"}

		reqBody, err := json.Marshal(req)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader(reqBody)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("invalid_body", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		tenantID := "tenant-id"
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", bytes.NewReader([]byte("invalid"))).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.TenantService)
		h := NewTenantHandler(service)

		httpRequest := httptest.NewRequest(http.MethodPost, "/v1/tenants/me/invite", nil)
		rr := httptest.NewRecorder()

		h.InviteUser(rr, httpRequest)

		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
