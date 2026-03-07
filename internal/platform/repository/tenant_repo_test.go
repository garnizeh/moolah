package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTenantRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := domain.CreateTenantInput{Name: "Acme Corp"}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("CreateTenant", ctx, mock.MatchedBy(func(arg sqlc.CreateTenantParams) bool {
			return arg.Name == input.Name && arg.ID != ""
		})).Return(sqlc.Tenant{
			ID:        "01H7XFRP9K1A1A1A1A1A1A1A1A",
			Name:      "Acme Corp",
			Plan:      sqlc.TenantPlanFree,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		tenant, err := repo.Create(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, tenant)
		assert.Equal(t, "Acme Corp", tenant.Name)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("duplicate name", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("CreateTenant", ctx, mock.Anything).Return(sqlc.Tenant{}, &pgconn.PgError{Code: "23505"})

		tenant, err := repo.Create(ctx, input)

		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})
}

func TestTenantRepo_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "01H7XFRP9K1A1A1A1A1A1A1A1A"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{
			ID:        id,
			Name:      "Acme Corp",
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		tenant, err := repo.GetByID(ctx, id)

		require.NoError(t, err)
		assert.Equal(t, id, tenant.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{}, pgx.ErrNoRows)

		tenant, err := repo.GetByID(ctx, id)

		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestTenantRepo_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("ListTenants", ctx).Return([]sqlc.Tenant{
			{ID: "1", Name: "T1", CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
			{ID: "2", Name: "T2", CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
		}, nil)

		tenants, err := repo.List(ctx)

		require.NoError(t, err)
		assert.Len(t, tenants, 2)
		assert.Equal(t, "T1", tenants[0].Name)
		assert.Equal(t, "T2", tenants[1].Name)
	})
}

func TestTenantRepo_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "01H7XFRP9K1A1A1A1A1A1A1A1A"
	newName := "New Name"
	input := domain.UpdateTenantInput{Name: &newName}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{
			ID:   id,
			Name: "Old Name",
			Plan: sqlc.TenantPlanFree,
		}, nil)

		mockQuerier.On("UpdateTenant", ctx, mock.MatchedBy(func(arg sqlc.UpdateTenantParams) bool {
			return arg.ID == id && arg.Name == newName
		})).Return(sqlc.Tenant{
			ID:        id,
			Name:      newName,
			Plan:      sqlc.TenantPlanFree,
			UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		tenant, err := repo.Update(ctx, id, input)

		require.NoError(t, err)
		assert.Equal(t, newName, tenant.Name)
	})
}

func TestTenantRepo_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "01H7XFRP9K1A1A1A1A1A1A1A1A"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("SoftDeleteTenant", ctx, id).Return(nil)

		err := repo.Delete(ctx, id)

		require.NoError(t, err)
	})
}

func TestTenantRepo_Update_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "1"
	newName := "New"
	input := domain.UpdateTenantInput{Name: &newName}

	t.Run("get current fails", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{}, errors.New("db error"))

		tenant, err := repo.Update(ctx, id, input)

		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.Contains(t, err.Error(), "failed to get tenant for update")
	})

	t.Run("update fails with conflict", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{ID: id, Name: "Old"}, nil)
		mockQuerier.On("UpdateTenant", ctx, mock.Anything).Return(sqlc.Tenant{}, &pgconn.PgError{Code: "23505"})

		tenant, err := repo.Update(ctx, id, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrConflict)
		assert.Nil(t, tenant)
	})
}

func TestTenantRepo_Delete_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "1"

	t.Run("soft delete fails", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("SoftDeleteTenant", ctx, id).Return(errors.New("db error"))

		err := repo.Delete(ctx, id)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to soft delete tenant")
	})
}

func TestTenantRepo_MapTenant(t *testing.T) {
	t.Parallel()

	repo := &tenantRepo{}

	t.Run("with deleted_at", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		sqlcTenant := sqlc.Tenant{
			DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}
		tenant := repo.mapTenant(sqlcTenant)
		assert.NotNil(t, tenant.DeletedAt)
		assert.True(t, tenant.DeletedAt.Equal(now))
	})

	t.Run("without deleted_at", func(t *testing.T) {
		t.Parallel()
		sqlcTenant := sqlc.Tenant{
			DeletedAt: pgtype.Timestamptz{Valid: false},
		}
		tenant := repo.mapTenant(sqlcTenant)
		assert.Nil(t, tenant.DeletedAt)
	})
}

func TestTenantRepo_Errors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "1"

	t.Run("create database error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("CreateTenant", ctx, mock.Anything).Return(sqlc.Tenant{}, errors.New("db error"))

		tenant, err := repo.Create(ctx, domain.CreateTenantInput{Name: "Err"})
		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.Contains(t, err.Error(), "failed to create tenant")
	})

	t.Run("get by id query error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{}, errors.New("db error"))

		tenant, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.Contains(t, err.Error(), "failed to get tenant by id")
	})

	t.Run("list database error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("ListTenants", ctx).Return([]sqlc.Tenant(nil), errors.New("db error"))

		tenants, err := repo.List(ctx)
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Contains(t, err.Error(), "failed to list tenants")
	})

	t.Run("update generic database error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{ID: id}, nil)
		mockQuerier.On("UpdateTenant", ctx, mock.Anything).Return(sqlc.Tenant{}, errors.New("db error"))

		tenant, err := repo.Update(ctx, id, domain.UpdateTenantInput{})
		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.Contains(t, err.Error(), "failed to update tenant")
	})

	t.Run("update not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{}, pgx.ErrNoRows)

		tenant, err := repo.Update(ctx, id, domain.UpdateTenantInput{})
		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestTenantRepo_Update_Partial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "1"
	plan := domain.TenantPlanPremium
	input := domain.UpdateTenantInput{Plan: &plan}

	t.Run("update plan only", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(mocks.Querier)
		repo := NewTenantRepository(mockQuerier)

		mockQuerier.On("GetTenantByID", ctx, id).Return(sqlc.Tenant{
			ID:   id,
			Name: "Stay the same",
			Plan: sqlc.TenantPlanFree,
		}, nil)

		mockQuerier.On("UpdateTenant", ctx, mock.MatchedBy(func(arg sqlc.UpdateTenantParams) bool {
			return arg.ID == id && arg.Plan == sqlc.TenantPlanPremium && arg.Name == "Stay the same"
		})).Return(sqlc.Tenant{
			ID:   id,
			Name: "Stay the same",
			Plan: sqlc.TenantPlanPremium,
		}, nil)

		tenant, err := repo.Update(ctx, id, input)

		require.NoError(t, err)
		assert.Equal(t, domain.TenantPlanPremium, tenant.Plan)
	})
}
