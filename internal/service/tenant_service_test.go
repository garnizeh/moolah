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

func TestTenantService_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := domain.CreateTenantInput{Name: "Test Tenant"}
	tenant := &domain.Tenant{ID: "tenant_1", Name: "Test Tenant", Plan: domain.TenantPlanFree}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		auditRepo := new(mocks.AuditRepository)

		tenantRepo.On("Create", ctx, input).Return(tenant, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.TenantID == tenant.ID && in.Action == domain.AuditActionCreate && in.EntityType == "tenant"
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewTenantService(tenantRepo, nil, auditRepo)
		result, err := svc.Create(ctx, input)

		require.NoError(t, err)
		require.Equal(t, tenant, result)
		tenantRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("RepoError", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		tenantRepo.On("Create", ctx, input).Return(nil, errors.New("db error"))

		svc := service.NewTenantService(tenantRepo, nil, nil)
		_, err := svc.Create(ctx, input)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create tenant")
	})

	t.Run("AuditError", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		auditRepo := new(mocks.AuditRepository)

		tenantRepo.On("Create", ctx, input).Return(tenant, nil)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit error"))

		svc := service.NewTenantService(tenantRepo, nil, auditRepo)
		_, err := svc.Create(ctx, input)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create audit log")
	})
}

func TestTenantService_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "tenant_1"
	tenant := &domain.Tenant{ID: id, Name: "Test Tenant"}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		tenantRepo.On("GetByID", ctx, id).Return(tenant, nil)

		svc := service.NewTenantService(tenantRepo, nil, nil)
		result, err := svc.GetByID(ctx, id)

		require.NoError(t, err)
		require.Equal(t, tenant, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		tenantRepo.On("GetByID", ctx, id).Return(nil, domain.ErrNotFound)

		svc := service.NewTenantService(tenantRepo, nil, nil)
		_, err := svc.GetByID(ctx, id)

		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestTenantService_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenants := []domain.Tenant{{ID: "1", Name: "T1"}, {ID: "2", Name: "T2"}}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		tenantRepo.On("List", ctx).Return(tenants, nil)

		svc := service.NewTenantService(tenantRepo, nil, nil)
		result, err := svc.List(ctx)

		require.NoError(t, err)
		require.Equal(t, tenants, result)
	})
}

func TestTenantService_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "tenant_1"
	newName := "Updated Tenant"
	input := domain.UpdateTenantInput{Name: &newName}
	oldTenant := &domain.Tenant{ID: id, Name: "Old Tenant", Plan: domain.TenantPlanFree}
	newTenant := &domain.Tenant{ID: id, Name: newName, Plan: domain.TenantPlanFree}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		auditRepo := new(mocks.AuditRepository)

		tenantRepo.On("GetByID", ctx, id).Return(oldTenant, nil)
		tenantRepo.On("Update", ctx, id, input).Return(newTenant, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionUpdate && in.EntityID == id
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewTenantService(tenantRepo, nil, auditRepo)
		result, err := svc.Update(ctx, id, input)

		require.NoError(t, err)
		require.Equal(t, newTenant, result)
		tenantRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		tenantRepo.On("GetByID", ctx, id).Return(nil, domain.ErrNotFound)

		svc := service.NewTenantService(tenantRepo, nil, nil)
		_, err := svc.Update(ctx, id, input)

		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestTenantService_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "tenant_1"

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tenantRepo := new(mocks.TenantRepository)
		auditRepo := new(mocks.AuditRepository)

		auditRepo.On("Create", ctx, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionSoftDelete && in.EntityID == id
		})).Return(&domain.AuditLog{}, nil)
		tenantRepo.On("Delete", ctx, id).Return(nil)

		svc := service.NewTenantService(tenantRepo, nil, auditRepo)
		err := svc.Delete(ctx, id)

		require.NoError(t, err)
		tenantRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("AuditError", func(t *testing.T) {
		t.Parallel()

		auditRepo := new(mocks.AuditRepository)
		auditRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("audit error"))

		svc := service.NewTenantService(nil, nil, auditRepo)
		err := svc.Delete(ctx, id)

		require.Error(t, err)
	})
}

func TestTenantService_InviteUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	input := domain.CreateUserInput{
		Email: "new@example.com",
		Name:  "New User",
		Role:  domain.RoleMember,
	}
	user := &domain.User{ID: "user_1", Email: input.Email, TenantID: tenantID, Role: input.Role}

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mocks.UserRepository)
		auditRepo := new(mocks.AuditRepository)

		userRepo.On("Create", ctx, mock.MatchedBy(func(in domain.CreateUserInput) bool {
			return in.TenantID == tenantID && in.Email == input.Email
		})).Return(user, nil)
		auditRepo.On("Create", ctx, mock.MatchedBy(func(in domain.CreateAuditLogInput) bool {
			return in.Action == domain.AuditActionCreate && in.EntityType == "user" && in.EntityID == user.ID
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewTenantService(nil, userRepo, auditRepo)
		result, err := svc.InviteUser(ctx, tenantID, input)

		require.NoError(t, err)
		require.Equal(t, user, result)
		userRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("RepoError", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mocks.UserRepository)
		userRepo.On("Create", ctx, mock.Anything).Return(nil, errors.New("db error"))

		svc := service.NewTenantService(nil, userRepo, nil)
		_, err := svc.InviteUser(ctx, tenantID, input)

		require.Error(t, err)
	})
}
