//go:build integration

package bootstrap

import (
	"context"
	"testing"

	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureSysadmin_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pgDB := containers.NewPostgresDB(t)

	cfg := &config.Config{
		SysadminEmail:      "bootstrap-test@moolah.io",
		SysadminTenantName: "Bootstrap Tenant",
	}

	t.Run("it is idempotent and creates exactly one set of records", func(t *testing.T) {
		t.Parallel()

		// First run: Creates
		err := EnsureSysadmin(ctx, pgDB.Queries, cfg)
		require.NoError(t, err)

		// Verify user exists
		user, err := pgDB.Queries.GetUserByEmail(ctx, cfg.SysadminEmail)
		require.NoError(t, err)
		assert.Equal(t, cfg.SysadminEmail, user.Email)
		assert.Equal(t, sqlc.UserRoleSysadmin, user.Role)

		// Verify tenant exists
		tenant, err := pgDB.Queries.GetTenantByID(ctx, user.TenantID)
		require.NoError(t, err)
		assert.Equal(t, cfg.SysadminTenantName, tenant.Name)

		// Second run: No-op
		err = EnsureSysadmin(ctx, pgDB.Queries, cfg)
		require.NoError(t, err)

		// Verify count of users still 1
		users, err := pgDB.Queries.AdminListAllUsers(ctx)
		require.NoError(t, err)

		count := 0
		for _, u := range users {
			if u.Email == cfg.SysadminEmail {
				count++
			}
		}
		assert.Equal(t, 1, count, "should not create duplicate sysadmin")
	})
}
