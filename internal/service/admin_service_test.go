package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminService_TenantOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("ListAllTenants_Success", func(t *testing.T) {
		t.Parallel()
		tenantRepo := new(mocks.AdminTenantRepository)
		expected := []domain.Tenant{{ID: "t1"}, {ID: "t2"}}
		tenantRepo.On("ListAll", mock.Anything, true).Return(expected, nil)

		svc := service.NewAdminService(tenantRepo, nil, nil, nil)
		res, err := svc.ListAllTenants(ctx, true)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("GetTenantByID_NotFound", func(t *testing.T) {
		t.Parallel()
		tenantRepo := new(mocks.AdminTenantRepository)
		tenantRepo.On("GetByID", mock.Anything, "t1").Return((*domain.Tenant)(nil), nil)

		svc := service.NewAdminService(tenantRepo, nil, nil, nil)
		res, err := svc.GetTenantByID(ctx, "t1")

		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, res)
	})

	t.Run("UpdateTenantPlan_Success", func(t *testing.T) {
		t.Parallel()
		tenantRepo := new(mocks.AdminTenantRepository)
		auditRepo := new(mocks.AuditRepository)
		updated := &domain.Tenant{ID: "t1", Plan: domain.TenantPlanPremium}

		tenantRepo.On("UpdatePlan", mock.Anything, "t1", domain.TenantPlanPremium).Return(updated, nil)
		auditRepo.On("Create", mock.Anything, mock.MatchedBy(func(input domain.CreateAuditLogInput) bool {
			return input.EntityID == "t1" && input.Action == domain.AuditActionUpdate
		})).Return(&domain.AuditLog{}, nil)

		svc := service.NewAdminService(tenantRepo, nil, nil, auditRepo)
		res, err := svc.UpdateTenantPlan(ctx, "t1", domain.TenantPlanPremium)

		require.NoError(t, err)
		require.Equal(t, updated, res)
	})

	t.Run("SuspendTenant_Success", func(t *testing.T) {
		t.Parallel()
		tenantRepo := new(mocks.AdminTenantRepository)
		auditRepo := new(mocks.AuditRepository)

		tenantRepo.On("Suspend", mock.Anything, "t1").Return(nil)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)

		svc := service.NewAdminService(tenantRepo, nil, nil, auditRepo)
		err := svc.SuspendTenant(ctx, "t1")

		require.NoError(t, err)
		tenantRepo.AssertExpectations(t)
	})

	t.Run("HardDeleteTenant_InvalidToken", func(t *testing.T) {
		t.Parallel()
		svc := service.NewAdminService(nil, nil, nil, nil)
		err := svc.HardDeleteTenant(ctx, "t1", "wrong-token")
		require.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("HardDeleteTenant_Success", func(t *testing.T) {
		t.Parallel()
		tenantRepo := new(mocks.AdminTenantRepository)
		auditRepo := new(mocks.AuditRepository)

		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)
		tenantRepo.On("HardDelete", mock.Anything, "t1").Return(nil)

		svc := service.NewAdminService(tenantRepo, nil, nil, auditRepo)
		err := svc.HardDeleteTenant(ctx, "t1", "t1")

		require.NoError(t, err)
		tenantRepo.AssertExpectations(t)
	})
}

func TestAdminService_UserOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("ListAllUsers_Success", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.AdminUserRepository)
		expected := []domain.User{{ID: "u1"}}
		userRepo.On("ListAll", mock.Anything).Return(expected, nil)

		svc := service.NewAdminService(nil, userRepo, nil, nil)
		res, err := svc.ListAllUsers(ctx)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("ForceDeleteUser_Success", func(t *testing.T) {
		t.Parallel()
		userRepo := new(mocks.AdminUserRepository)
		auditRepo := new(mocks.AuditRepository)

		auditRepo.On("Create", mock.Anything, mock.Anything).Return(&domain.AuditLog{}, nil)
		userRepo.On("ForceDelete", mock.Anything, "u1").Return(nil)

		svc := service.NewAdminService(nil, userRepo, nil, auditRepo)
		err := svc.ForceDeleteUser(ctx, "u1")

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
	})
}

func TestAdminService_AuditOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("ListAuditLogs_Success", func(t *testing.T) {
		t.Parallel()
		adminAudit := new(mocks.AdminAuditRepository)
		params := domain.ListAuditLogsParams{Limit: 10}
		expected := []domain.AuditLog{{ID: "a1"}}
		adminAudit.On("ListAll", mock.Anything, params).Return(expected, nil)

		svc := service.NewAdminService(nil, nil, adminAudit, nil)
		res, err := svc.ListAuditLogs(ctx, params)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})
}
