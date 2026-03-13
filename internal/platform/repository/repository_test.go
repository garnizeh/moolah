package repository

import (
	"errors"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestTranslateError(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, TranslateError(nil))
	})

	t.Run("pgx.ErrNoRows", func(t *testing.T) {
		t.Parallel()
		err := TranslateError(pgx.ErrNoRows)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("pgx unique_violation", func(t *testing.T) {
		t.Parallel()
		pgErr := &pgconn.PgError{Code: "23505"}
		err := TranslateError(pgErr)
		assert.ErrorIs(t, err, domain.ErrConflict)
	})

	t.Run("pgx foreign_key_violation", func(t *testing.T) {
		t.Parallel()
		pgErr := &pgconn.PgError{Code: "23503"}
		err := TranslateError(pgErr)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("generic error", func(t *testing.T) {
		t.Parallel()
		originalErr := errors.New("some other error")
		err := TranslateError(originalErr)
		assert.Equal(t, originalErr, err)
	})

	t.Run("unhandled pg error", func(t *testing.T) {
		t.Parallel()
		pgErr := &pgconn.PgError{Code: "99999"}
		err := TranslateError(pgErr)
		assert.Equal(t, pgErr, err)
	})
}
