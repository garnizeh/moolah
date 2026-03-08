package db

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/platform/db/migrations"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Querier is a helper function that connects to the database, runs migrations, and returns a sqlc.Querier instance.
func Querier(ctx context.Context, databaseURL string) (sqlc.Querier, error) {
	dbPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbPool.Close()

	if err := runMigrations(dbPool); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	querier := sqlc.New(dbPool)

	return querier, nil
}

func runMigrations(dbPool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(dbPool)
	defer db.Close()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	return nil
}
