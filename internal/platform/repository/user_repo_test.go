package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	input := domain.CreateUserInput{
		TenantID: "01H7XFRP9K1A1A1A1A1A1A1A1A",
		Email:    "test@example.com",
		Name:     "Test User",
		Role:     domain.RoleMember,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("CreateUser", ctx, mock.MatchedBy(func(arg sqlc.CreateUserParams) bool {
			return arg.TenantID == input.TenantID && arg.Email == input.Email && arg.ID != ""
		})).Return(sqlc.User{
			ID:        "01H7XFRP9K1A1A1A1A1A1A1A1B",
			TenantID:  input.TenantID,
			Email:     input.Email,
			Name:      input.Name,
			Role:      sqlc.UserRoleMember,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

		user, err := repo.Create(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, input.Email, user.Email)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("CreateUser", ctx, mock.Anything).Return(sqlc.User{}, &pgconn.PgError{Code: "23505"})

		user, err := repo.Create(ctx, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrConflict)
		assert.Nil(t, user)
	})

	t.Run("generic db error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("CreateUser", ctx, mock.Anything).Return(sqlc.User{}, errors.New("db error"))

		user, err := repo.Create(ctx, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
		assert.Nil(t, user)
	})
}

func TestUserRepo_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "T1"
	id := "U1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, sqlc.GetUserByIDParams{ID: id, TenantID: tenantID}).Return(sqlc.User{
			ID:       id,
			TenantID: tenantID,
			Email:    "test@example.com",
		}, nil)

		user, err := repo.GetByID(ctx, tenantID, id)

		require.NoError(t, err)
		assert.Equal(t, id, user.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, mock.Anything).Return(sqlc.User{}, pgx.ErrNoRows)

		user, err := repo.GetByID(ctx, tenantID, id)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, user)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, mock.Anything).Return(sqlc.User{}, errors.New("db error"))

		user, err := repo.GetByID(ctx, tenantID, id)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user by id")
		assert.Nil(t, user)
	})
}

func TestUserRepo_GetByEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "test@example.com"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByEmail", ctx, sqlc.GetUserByEmailParams{Email: email}).Return(sqlc.User{
			ID:    "U1",
			Email: email,
		}, nil)

		user, err := repo.GetByEmail(ctx, email)

		require.NoError(t, err)
		assert.Equal(t, email, user.Email)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByEmail", ctx, mock.Anything).Return(sqlc.User{}, pgx.ErrNoRows)

		user, err := repo.GetByEmail(ctx, email)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, user)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByEmail", ctx, mock.Anything).Return(sqlc.User{}, errors.New("db error"))

		user, err := repo.GetByEmail(ctx, email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user by email")
		assert.Nil(t, user)
	})
}

func TestUserRepo_ListByTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "T1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("ListUsersByTenant", ctx, tenantID).Return([]sqlc.User{
			{ID: "U1", TenantID: tenantID},
			{ID: "U2", TenantID: tenantID},
		}, nil)

		list, err := repo.ListByTenant(ctx, tenantID)

		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("ListUsersByTenant", ctx, tenantID).Return([]sqlc.User(nil), errors.New("db error"))

		list, err := repo.ListByTenant(ctx, tenantID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list users")
		assert.Nil(t, list)
	})
}

func TestUserRepo_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "T1"
	id := "U1"
	newName := "New Name"
	input := domain.UpdateUserInput{Name: &newName}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, sqlc.GetUserByIDParams{ID: id, TenantID: tenantID}).Return(sqlc.User{
			ID:       id,
			TenantID: tenantID,
			Name:     "Old Name",
			Role:     sqlc.UserRoleMember,
		}, nil)

		mockQuerier.On("UpdateUser", ctx, mock.MatchedBy(func(arg sqlc.UpdateUserParams) bool {
			return arg.ID == id && arg.Name == newName
		})).Return(sqlc.User{
			ID:   id,
			Name: newName,
			Role: sqlc.UserRoleMember,
		}, nil)

		user, err := repo.Update(ctx, tenantID, id, input)

		require.NoError(t, err)
		assert.Equal(t, newName, user.Name)
	})

	t.Run("success with role update", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)
		newRole := domain.RoleAdmin
		roleInput := domain.UpdateUserInput{Role: &newRole}

		mockQuerier.On("GetUserByID", ctx, sqlc.GetUserByIDParams{ID: id, TenantID: tenantID}).Return(sqlc.User{
			ID:       id,
			TenantID: tenantID,
			Name:     "Old Name",
			Role:     sqlc.UserRoleMember,
		}, nil)

		mockQuerier.On("UpdateUser", ctx, mock.MatchedBy(func(arg sqlc.UpdateUserParams) bool {
			return arg.ID == id && string(arg.Role) == string(newRole)
		})).Return(sqlc.User{
			ID:   id,
			Name: "Old Name",
			Role: sqlc.UserRoleAdmin,
		}, nil)

		user, err := repo.Update(ctx, tenantID, id, roleInput)

		require.NoError(t, err)
		assert.Equal(t, domain.RoleAdmin, user.Role)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, mock.Anything).Return(sqlc.User{}, pgx.ErrNoRows)

		user, err := repo.Update(ctx, tenantID, id, input)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, user)
	})

	t.Run("db error on get", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, mock.Anything).Return(sqlc.User{}, errors.New("db error"))

		user, err := repo.Update(ctx, tenantID, id, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user for update")
		assert.Nil(t, user)
	})

	t.Run("db error on update", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("GetUserByID", ctx, sqlc.GetUserByIDParams{ID: id, TenantID: tenantID}).Return(sqlc.User{
			ID:       id,
			TenantID: tenantID,
			Name:     "Old Name",
			Role:     sqlc.UserRoleMember,
		}, nil)

		mockQuerier.On("UpdateUser", ctx, mock.Anything).Return(sqlc.User{}, errors.New("db error"))

		user, err := repo.Update(ctx, tenantID, id, input)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user")
		assert.Nil(t, user)
	})
}

func TestUserRepo_UpdateLastLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := "U1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("UpdateUserLastLogin", ctx, sqlc.UpdateUserLastLoginParams{ID: id}).Return(nil)

		err := repo.UpdateLastLogin(ctx, id)

		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("UpdateUserLastLogin", ctx, mock.Anything).Return(errors.New("db error"))

		err := repo.UpdateLastLogin(ctx, id)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user last login")
	})
}

func TestUserRepo_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "T1"
	id := "U1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("SoftDeleteUser", ctx, sqlc.SoftDeleteUserParams{ID: id, TenantID: tenantID}).Return(nil)

		err := repo.Delete(ctx, tenantID, id)

		require.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		mockQuerier := new(sqlc.MockQuerier)
		repo := NewUserRepository(mockQuerier)

		mockQuerier.On("SoftDeleteUser", ctx, mock.Anything).Return(errors.New("db error"))

		err := repo.Delete(ctx, tenantID, id)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to soft delete user")
	})
}

func TestUserRepo_MapUser(t *testing.T) {
	t.Parallel()

	repo := &userRepo{}

	t.Run("all fields valid", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		sqlcUser := sqlc.User{
			DeletedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			LastLoginAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
		}

		user := repo.mapUser(sqlcUser)

		require.NotNil(t, user.DeletedAt)
		require.NotNil(t, user.LastLoginAt)
		assert.True(t, user.DeletedAt.Equal(now))
		assert.True(t, user.LastLoginAt.Equal(now.Add(time.Hour)))
	})

	t.Run("optional fields invalid", func(t *testing.T) {
		t.Parallel()
		sqlcUser := sqlc.User{
			DeletedAt:   pgtype.Timestamptz{Valid: false},
			LastLoginAt: pgtype.Timestamptz{Valid: false},
		}

		user := repo.mapUser(sqlcUser)

		assert.Nil(t, user.DeletedAt)
		assert.Nil(t, user.LastLoginAt)
	})
}
