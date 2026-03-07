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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuditRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	actorID := "actor_1"
	entityID := "entity_1"
	ip := "127.0.0.1"
	ua := "Go-Test"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		input := domain.CreateAuditLogInput{
			TenantID:   tenantID,
			ActorID:    actorID,
			ActorRole:  domain.RoleAdmin,
			Action:     domain.AuditActionCreate,
			EntityType: "account",
			EntityID:   entityID,
			IPAddress:  ip,
			UserAgent:  ua,
			OldValues:  nil,
			NewValues:  []byte(`{"name":"test"}`),
		}

		addr, _ := netip.ParseAddr(ip)

		mockQuerier.On("CreateAuditLog", ctx, mock.MatchedBy(func(arg sqlc.CreateAuditLogParams) bool {
			return arg.TenantID == tenantID &&
				arg.ActorID == actorID &&
				arg.ActorRole == sqlc.UserRoleAdmin &&
				arg.Action == sqlc.AuditActionCreate &&
				arg.EntityType == "account" &&
				arg.EntityID == pgtype.Text{String: entityID, Valid: true} &&
				arg.IpAddress.String() == addr.String() &&
				arg.UserAgent == pgtype.Text{String: ua, Valid: true} &&
				string(arg.NewValues) == `{"name":"test"}`
		})).Return(sqlc.AuditLog{
			ID:         "ulid_1",
			TenantID:   tenantID,
			ActorID:    actorID,
			ActorRole:  sqlc.UserRoleAdmin,
			Action:     sqlc.AuditActionCreate,
			EntityType: "account",
			EntityID:   pgtype.Text{String: entityID, Valid: true},
			NewValues:  []byte(`{"name":"test"}`),
			IpAddress:  &addr,
			UserAgent:  pgtype.Text{String: ua, Valid: true},
			CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		log, err := repo.Create(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, log)
		require.Equal(t, "ulid_1", log.ID)
		require.Equal(t, ip, log.IPAddress)
	})

	t.Run("db_error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		input := domain.CreateAuditLogInput{
			TenantID:   tenantID,
			ActorID:    actorID,
			ActorRole:  domain.RoleAdmin,
			Action:     domain.AuditActionSoftDelete,
			EntityType: "category",
		}

		mockQuerier.On("CreateAuditLog", ctx, mock.Anything).Return(sqlc.AuditLog{}, errors.New("db error"))

		log, err := repo.Create(ctx, input)
		require.Error(t, err)
		require.Nil(t, log)
		require.Contains(t, err.Error(), "db error")
	})
}

func TestAuditRepo_ListByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		params := domain.ListAuditLogsParams{
			Limit:  10,
			Offset: 0,
		}

		mockQuerier.On("ListAuditLogsByTenant", ctx, sqlc.ListAuditLogsByTenantParams{
			TenantID:  tenantID,
			LimitOff:  10,
			OffsetOff: 0,
		}).Return([]sqlc.AuditLog{
			{ID: "log_1", TenantID: tenantID, Action: sqlc.AuditActionCreate},
			{ID: "log_2", TenantID: tenantID, Action: sqlc.AuditActionUpdate},
		}, nil)

		logs, err := repo.ListByTenant(ctx, tenantID, params)
		require.NoError(t, err)
		require.Len(t, logs, 2)
		require.Equal(t, "log_1", logs[0].ID)
		require.Equal(t, domain.AuditActionCreate, logs[0].Action)
	})

	t.Run("db_error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		mockQuerier.On("ListAuditLogsByTenant", ctx, mock.Anything).Return(([]sqlc.AuditLog)(nil), errors.New("db error"))

		logs, err := repo.ListByTenant(ctx, tenantID, domain.ListAuditLogsParams{})
		require.Error(t, err)
		require.Nil(t, logs)
	})
}

func TestAuditRepo_ListByEntity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "tenant_1"
	entityID := "acc_123"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		mockQuerier.On("ListAuditLogsByEntity", ctx, sqlc.ListAuditLogsByEntityParams{
			TenantID:   tenantID,
			EntityType: "account",
			EntityID:   pgtype.Text{String: entityID, Valid: true},
		}).Return([]sqlc.AuditLog{
			{ID: "log_1", TenantID: tenantID, EntityType: "account", EntityID: pgtype.Text{String: entityID, Valid: true}},
		}, nil)

		logs, err := repo.ListByEntity(ctx, tenantID, "account", entityID)
		require.NoError(t, err)
		require.Len(t, logs, 1)
		require.Equal(t, entityID, logs[0].EntityID)
	})

	t.Run("db_error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		mockQuerier.On("ListAuditLogsByEntity", ctx, mock.Anything).Return(([]sqlc.AuditLog)(nil), errors.New("db error"))

		logs, err := repo.ListByEntity(ctx, tenantID, "account", entityID)
		require.Error(t, err)
		require.Nil(t, logs)
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewAuditRepository(mockQuerier)

		mockQuerier.On("ListAuditLogsByEntity", ctx, mock.Anything).Return(([]sqlc.AuditLog)(nil), pgx.ErrNoRows)

		logs, err := repo.ListByEntity(ctx, tenantID, "account", entityID)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, logs)
	})
}
