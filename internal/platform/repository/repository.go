// Package repository provides implementations of repositories for the domain layer.
package repository

import (
	"errors"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

// fromTime converts a pgtype.Timestamptz to a *time.Time, returning nil if the value is not valid.
func fromTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// toTime converts a *time.Time to a pgtype.Timestamptz, marking it as invalid if the pointer is nil.
func toText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// fromText converts a pgtype.Text to a *string, returning nil if the value is not valid.
func fromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}
