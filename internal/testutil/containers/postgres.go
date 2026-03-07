//go:build integration

package containers

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
)

// TestPostgresDB holds an active pgxpool and the sqlc Queries layer bound to it.
type TestPostgresDB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// NewPostgresDB starts an ephemeral PostgreSQL container, applies docs/schema.sql,
// and returns a connected TestPostgresDB. Container and pool are cleaned up when t
// completes.
func NewPostgresDB(t *testing.T) *TestPostgresDB {
	t.Helper()
	ctx := context.Background()

	// Get absolute path to schema.sql relative to this file
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	schemaPath := filepath.Join(basepath, "..", "..", "..", "docs", "schema.sql")

	pgc, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("moolah_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithInitScripts(schemaPath),
	)
	require.NoError(t, err, "failed to start postgres container")

	dsn, err := pgc.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "failed to connect to postgres")

	t.Cleanup(func() {
		pool.Close()
		err := pgc.Terminate(ctx)
		require.NoError(t, err, "failed to terminate postgres container")
	})

	return &TestPostgresDB{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}
