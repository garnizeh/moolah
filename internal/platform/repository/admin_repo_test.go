package repository

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminTenantRepo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	id := "01H7XFRP9K1A1A1A1A1A1A1A1A"

	t.Run("ListAll", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminListAllTenants", ctx, true).Return([]sqlc.Tenant{
			{ID: id, Name: "Tenant 1", Plan: sqlc.TenantPlanFree, CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}, DeletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
		}, nil)

		tenants, err := repo.ListAll(ctx, true)
		require.NoError(t, err)
		assert.Len(t, tenants, 1)
		assert.Equal(t, id, tenants[0].ID)
		assert.NotNil(t, tenants[0].DeletedAt)
	})

	t.Run("ListAll error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminListAllTenants", ctx, mock.Anything).Return(nil, errors.New("db error"))

		tenants, err := repo.ListAll(ctx, true)
		require.Error(t, err)
		assert.Nil(t, tenants)
	})

	t.Run("GetByID success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminGetTenantByID", ctx, id).Return(sqlc.Tenant{
			ID: id, Name: "Tenant 1", Plan: sqlc.TenantPlanFree, CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		tenant, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, tenant.ID)
	})

	t.Run("GetByID error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminGetTenantByID", ctx, id).Return(sqlc.Tenant{}, errors.New("db error"))

		tenant, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, tenant)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminGetTenantByID", ctx, id).Return(sqlc.Tenant{}, pgx.ErrNoRows)

		tenant, err := repo.GetByID(ctx, id)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, tenant)
	})

	t.Run("UpdatePlan", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminUpdateTenantPlan", ctx, sqlc.AdminUpdateTenantPlanParams{
			ID:   id,
			Plan: sqlc.TenantPlanPremium,
		}).Return(sqlc.Tenant{
			ID: id, Plan: sqlc.TenantPlanPremium, CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		tenant, err := repo.UpdatePlan(ctx, id, domain.TenantPlanPremium)
		require.NoError(t, err)
		assert.Equal(t, domain.TenantPlanPremium, tenant.Plan)
	})

	t.Run("UpdatePlan not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminUpdateTenantPlan", ctx, mock.Anything).Return(sqlc.Tenant{}, pgx.ErrNoRows)

		tenant, err := repo.UpdatePlan(ctx, id, domain.TenantPlanPremium)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, tenant)
	})

	t.Run("UpdatePlan error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminUpdateTenantPlan", ctx, mock.Anything).Return(sqlc.Tenant{}, errors.New("db error"))

		tenant, err := repo.UpdatePlan(ctx, id, domain.TenantPlanPremium)
		require.Error(t, err)
		assert.Nil(t, tenant)
	})

	t.Run("Suspend success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminSuspendTenant", ctx, id).Return(nil)

		err := repo.Suspend(ctx, id)
		require.NoError(t, err)
	})

	t.Run("Suspend error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminSuspendTenant", ctx, id).Return(errors.New("db error"))

		err := repo.Suspend(ctx, id)
		require.Error(t, err)
	})

	t.Run("Restore success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminRestoreTenant", ctx, id).Return(nil)

		err := repo.Restore(ctx, id)
		require.NoError(t, err)
	})

	t.Run("Restore error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminRestoreTenant", ctx, id).Return(errors.New("db error"))

		err := repo.Restore(ctx, id)
		require.Error(t, err)
	})

	t.Run("HardDelete success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminHardDeleteTenant", ctx, id).Return(nil)

		err := repo.HardDelete(ctx, id)
		require.NoError(t, err)
	})

	t.Run("HardDelete error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminTenantRepository(mockQuerier)

		mockQuerier.On("AdminHardDeleteTenant", ctx, id).Return(errors.New("db error"))

		err := repo.HardDelete(ctx, id)
		require.Error(t, err)
	})
}

func TestAdminUserRepo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := "01H7XFRP9K1A1A1A1A1A1A1A1U"

	t.Run("ListAll", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminListAllUsers", ctx).Return([]sqlc.User{
			{
				ID: userID, Email: "user@example.com", Role: sqlc.UserRoleMember, CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				DeletedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
				LastLoginAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
		}, nil)

		users, err := repo.ListAll(ctx)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, userID, users[0].ID)
		assert.NotNil(t, users[0].DeletedAt)
		assert.NotNil(t, users[0].LastLoginAt)
	})

	t.Run("ListAll error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminListAllUsers", ctx).Return(nil, errors.New("db error"))

		users, err := repo.ListAll(ctx)
		require.Error(t, err)
		assert.Nil(t, users)
	})

	t.Run("GetByID success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminGetUserByID", ctx, userID).Return(sqlc.User{
			ID: userID, Email: "user@example.com", Role: sqlc.UserRoleMember, CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		user, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
	})

	t.Run("GetByID error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminGetUserByID", ctx, userID).Return(sqlc.User{}, errors.New("db error"))

		user, err := repo.GetByID(ctx, userID)
		require.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminGetUserByID", ctx, userID).Return(sqlc.User{}, pgx.ErrNoRows)

		user, err := repo.GetByID(ctx, userID)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, user)
	})

	t.Run("ForceDelete success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminForceDeleteUser", ctx, userID).Return(nil)

		err := repo.ForceDelete(ctx, userID)
		require.NoError(t, err)
	})

	t.Run("ForceDelete error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminUserRepository(mockQuerier)

		mockQuerier.On("AdminForceDeleteUser", ctx, userID).Return(errors.New("db error"))

		err := repo.ForceDelete(ctx, userID)
		require.Error(t, err)
	})
}

func TestAdminAuditRepo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("ListAll", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminAuditRepository(mockQuerier)

		now := time.Now()
		params := domain.ListAuditLogsParams{
			Limit:      10,
			Offset:     5,
			EntityType: "account",
			EntityID:   "01H7XFRP9K1A1A1A1A1A1A1A1E",
			ActorID:    "01H7XFRP9K1A1A1A1A1A1A1A1A",
			Action:     domain.AuditActionCreate,
			StartDate:  &now,
			EndDate:    &now,
		}

		mockQuerier.On("AdminListAllAuditLogs", ctx, mock.MatchedBy(func(arg sqlc.AdminListAllAuditLogsParams) bool {
			return arg.LimitOff == 10 &&
				arg.OffsetOff == 5 &&
				arg.EntityType.String == "account" &&
				arg.EntityID.String == "01H7XFRP9K1A1A1A1A1A1A1A1E" &&
				arg.ActorID == "01H7XFRP9K1A1A1A1A1A1A1A1A" &&
				arg.Action.AuditAction == sqlc.AuditActionCreate &&
				arg.StartDate.Time.Equal(now) &&
				arg.EndDate.Time.Equal(now)
		})).Return([]sqlc.AuditLog{
			{
				ID:        "01H7XFRP9K1A1A1A1A1A1A1A1L",
				Action:    sqlc.AuditActionCreate,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				IpAddress: &netip.Addr{},
				EntityID:  pgtype.Text{String: "01H7XFRP9K1A1A1A1A1A1A1A1E", Valid: true},
				UserAgent: pgtype.Text{String: "Mozilla", Valid: true},
			},
		}, nil)

		logs, err := repo.ListAll(ctx, params)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, "01H7XFRP9K1A1A1A1A1A1A1A1E", logs[0].EntityID)
		assert.Equal(t, "Mozilla", logs[0].UserAgent)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("ListAll error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAdminAuditRepository(mockQuerier)

		mockQuerier.On("AdminListAllAuditLogs", ctx, mock.Anything).Return(nil, errors.New("db error"))

		logs, err := repo.ListAll(ctx, domain.ListAuditLogsParams{})
		require.Error(t, err)
		assert.Nil(t, logs)
	})
}
