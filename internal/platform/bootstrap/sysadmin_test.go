package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/garnizeh/moolah/internal/config"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEnsureSysadmin(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		SysadminEmail:      "admin@test.com",
		SysadminTenantName: "System",
	}

	t.Run("skip if sysadmin already exists", func(t *testing.T) {
		t.Parallel()

		m := new(mocks.Querier)
		m.On("GetUserByEmail", mock.Anything, cfg.SysadminEmail).Return(sqlc.User{
			Role: sqlc.UserRoleSysadmin,
		}, nil)

		err := EnsureSysadmin(context.Background(), m, cfg)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("error if user exists with different role", func(t *testing.T) {
		t.Parallel()

		m := new(mocks.Querier)
		m.On("GetUserByEmail", mock.Anything, cfg.SysadminEmail).Return(sqlc.User{
			Role: sqlc.UserRoleMember,
		}, nil)

		err := EnsureSysadmin(context.Background(), m, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exists but is not a sysadmin")
		m.AssertExpectations(t)
	})

	t.Run("bootstrap on first run", func(t *testing.T) {
		t.Parallel()

		m := new(mocks.Querier)
		m.On("GetUserByEmail", mock.Anything, cfg.SysadminEmail).Return(sqlc.User{}, pgx.ErrNoRows)

		// Expect tenant creation
		m.On("CreateTenant", mock.Anything, mock.MatchedBy(func(p sqlc.CreateTenantParams) bool {
			return p.Name == cfg.SysadminTenantName
		})).Return(sqlc.Tenant{}, nil)

		// Expect user creation
		m.On("CreateUser", mock.Anything, mock.MatchedBy(func(p sqlc.CreateUserParams) bool {
			return p.Email == cfg.SysadminEmail && p.Role == sqlc.UserRoleSysadmin
		})).Return(sqlc.User{}, nil)

		err := EnsureSysadmin(context.Background(), m, cfg)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("error on lookup failure", func(t *testing.T) {
		t.Parallel()

		m := new(mocks.Querier)
		m.On("GetUserByEmail", mock.Anything, cfg.SysadminEmail).Return(sqlc.User{}, errors.New("db error"))

		err := EnsureSysadmin(context.Background(), m, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check existing sysadmin")
		m.AssertExpectations(t)
	})

	t.Run("error on tenant creation failure", func(t *testing.T) {
		t.Parallel()

		m := new(mocks.Querier)
		m.On("GetUserByEmail", mock.Anything, cfg.SysadminEmail).Return(sqlc.User{}, pgx.ErrNoRows)
		m.On("CreateTenant", mock.Anything, mock.Anything).Return(sqlc.Tenant{}, errors.New("tenant creation failed"))

		err := EnsureSysadmin(context.Background(), m, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create system tenant")
		m.AssertExpectations(t)
	})

	t.Run("error on user creation failure", func(t *testing.T) {
		t.Parallel()

		m := new(mocks.Querier)
		m.On("GetUserByEmail", mock.Anything, cfg.SysadminEmail).Return(sqlc.User{}, pgx.ErrNoRows)
		m.On("CreateTenant", mock.Anything, mock.Anything).Return(sqlc.Tenant{}, nil)
		m.On("CreateUser", mock.Anything, mock.Anything).Return(sqlc.User{}, errors.New("user creation failed"))

		err := EnsureSysadmin(context.Background(), m, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create sysadmin user")
		m.AssertExpectations(t)
	})
}
