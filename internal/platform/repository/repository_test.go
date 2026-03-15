package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

	t.Run("fromTime", func(t *testing.T) {
		t.Parallel()
		now := time.Now()

		valid := pgtype.Timestamptz{Time: now, Valid: true}
		invalid := pgtype.Timestamptz{Valid: false}

		assert.Equal(t, &now, fromTime(valid))
		assert.Nil(t, fromTime(invalid))
	})

	t.Run("toText", func(t *testing.T) {
		t.Parallel()
		str := "hello"

		resValid := toText(&str)
		resInvalid := toText(nil)

		assert.True(t, resValid.Valid)
		assert.Equal(t, str, resValid.String)
		assert.False(t, resInvalid.Valid)
	})

	t.Run("fromText", func(t *testing.T) {
		t.Parallel()
		str := "world"

		valid := pgtype.Text{String: str, Valid: true}
		invalid := pgtype.Text{Valid: false}

		assert.Equal(t, &str, fromText(valid))
		assert.Nil(t, fromText(invalid))
	})

	t.Run("valOrNil", func(t *testing.T) {
		t.Parallel()
		val := 123
		resValid := valOrNil(&val)
		resInvalid := valOrNil(nil)

		assert.True(t, resValid.Valid)
		assert.Equal(t, int32(val), resValid.Int32)
		assert.False(t, resInvalid.Valid)
	})

	t.Run("valOrNil64", func(t *testing.T) {
		t.Parallel()
		val := int64(123456)
		resValid := valOrNil64(&val)
		resInvalid := valOrNil64(nil)

		assert.True(t, resValid.Valid)
		assert.Equal(t, val, resValid.Int64)
		assert.False(t, resInvalid.Valid)
	})

	t.Run("fromPgTimestamptz", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		valid := pgtype.Timestamptz{Time: now, Valid: true}
		invalid := pgtype.Timestamptz{Valid: false}

		assert.Equal(t, &now, fromPgTimestamptz(valid))
		assert.Nil(t, fromPgTimestamptz(invalid))
	})

	t.Run("fromPgInt4", func(t *testing.T) {
		t.Parallel()
		val := 123
		valid := pgtype.Int4{Int32: int32(val), Valid: true}
		invalid := pgtype.Int4{Valid: false}

		assert.Equal(t, &val, fromPgInt4(valid))
		assert.Nil(t, fromPgInt4(invalid))
	})

	t.Run("fromPgInt8", func(t *testing.T) {
		t.Parallel()
		val := int64(123456)
		valid := pgtype.Int8{Int64: val, Valid: true}
		invalid := pgtype.Int8{Valid: false}

		assert.Equal(t, &val, fromPgInt8(valid))
		assert.Nil(t, fromPgInt8(invalid))
	})

	t.Run("toPgTimestamptz", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		fallback := pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}

		resValid := toPgTimestamptz(&now, fallback)
		resFallback := toPgTimestamptz(nil, fallback)

		assert.True(t, resValid.Valid)
		assert.Equal(t, now, resValid.Time)
		assert.Equal(t, fallback, resFallback)
	})

	t.Run("safeTime", func(t *testing.T) {
		t.Parallel()
		now := time.Now()

		assert.Equal(t, now, safeTime(&now))
		assert.Equal(t, time.Time{}, safeTime(nil))
	})

	t.Run("valOrDefault", func(t *testing.T) {
		t.Parallel()
		val := "hello"
		def := "default"

		assert.Equal(t, val, valOrDefault(&val, def))
		assert.Equal(t, def, valOrDefault(nil, def))
	})

	t.Run("toPgInt4", func(t *testing.T) {
		t.Parallel()
		val := 456
		fallback := pgtype.Int4{Int32: 789, Valid: true}

		resValid := toPgInt4(&val, fallback)
		resFallback := toPgInt4(nil, fallback)

		assert.True(t, resValid.Valid)
		assert.Equal(t, int32(val), resValid.Int32)
		assert.Equal(t, fallback, resFallback)
	})

	t.Run("toPgInt8", func(t *testing.T) {
		t.Parallel()
		val := int64(456789)
		fallback := pgtype.Int8{Int64: 987654, Valid: true}

		resValid := toPgInt8(&val, fallback)
		resFallback := toPgInt8(nil, fallback)

		assert.True(t, resValid.Valid)
		assert.Equal(t, val, resValid.Int64)
		assert.Equal(t, fallback, resFallback)
	})
}
