// Package repository provides implementations of repositories for the domain layer.
package repository

import (
	"errors"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TranslateError maps database-specific errors (pgx, postgres) to domain-specific errors.
func TranslateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505": // unique_violation
			return domain.ErrConflict
		case "23503": // foreign_key_violation
			return domain.ErrNotFound
		}
	}

	return err
}
