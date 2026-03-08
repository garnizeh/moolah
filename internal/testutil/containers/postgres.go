//go:build integration

package containers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/garnizeh/moolah/internal/platform/db/migrations"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
)

// TestPostgresDB holds an active pgxpool and the sqlc Queries layer bound to it.
type TestPostgresDB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

var gooseMigrationMu sync.Mutex

// NewPostgresDB starts an ephemeral PostgreSQL container, applies migrations,
// and returns a connected TestPostgresDB. Container and pool are cleaned up when t
// completes.
func NewPostgresDB(t *testing.T) *TestPostgresDB {
	t.Helper()
	ctx := context.Background()

	// Start the PostgreSQL container
	pgc, err := tcpostgres.Run(ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("moolah_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	// Get the connection string from the container
	dsn, err := pgc.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	// Get the configuration from your DSN
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err, "failed to parse config")

	// Create the pgxpool as usual
	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err, "failed to connect to postgres")

	// Open the stdlib bridge for goose migrations
	db := stdlib.OpenDBFromPool(pool)

	// Goose keeps migration configuration in package-level state, so running
	// setup concurrently across integration suites triggers race detector failures.
	gooseMigrationMu.Lock()
	goose.SetBaseFS(migrations.FS)
	err = goose.SetDialect("postgres")
	require.NoError(t, err, "failed to set goose dialect")
	err = goose.Up(db, ".")
	require.NoError(t, err, "failed to run migrations")
	gooseMigrationMu.Unlock()

	// Close the stdlib bridge
	err = db.Close()
	require.NoError(t, err, "failed to close sqlDB")

	// Ensure the pool is closed and container terminated when the test finishes
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
