//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/platform/db"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestQuerier_Integration(t *testing.T) {
	t.Parallel()

	// We don't use t.Parallel() here because goose uses package-level state
	// and we want to test the full connection + migration flow in isolation.
	ctx := context.Background()

	// Start a fresh Postgres container without pre-applied migrations
	pgc, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("moolah_db_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		terr := pgc.Terminate(ctx)
		require.NoError(t, terr)
	})

	dsn, err := pgc.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	t.Run("Connect and Run Migrations", func(t *testing.T) {
		t.Parallel()

		pool, querier, err := db.Querier(ctx, dsn)
		require.NoError(t, err)
		require.NotNil(t, pool)
		require.NotNil(t, querier)

		defer pool.Close()

		// Verify we can ping
		err = pool.Ping(ctx)
		require.NoError(t, err)

		// Verify migrations ran by checking if a table exists (e.g. users or tenants)
		// We can just try a simple query through the pool
		var exists bool
		err = pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
		require.NoError(t, err)
		require.True(t, exists, "migrations should have created the users table")
	})

	t.Run("Invalid Connection String", func(t *testing.T) {
		t.Parallel()

		pool, querier, err := db.Querier(ctx, "invalid-connection-string")
		require.Error(t, err)
		require.Nil(t, pool)
		require.Nil(t, querier)
	})
}
