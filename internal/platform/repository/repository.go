// Package repository provides implementations of repositories for the domain layer.
package repository

import (
	"errors"
	"math"
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

// valOrNil returns a pgtype.Int4 with Valid=false if the pointer is nil, otherwise returns the value.
func valOrNil(ptr *int) pgtype.Int4 {
	if ptr == nil {
		return pgtype.Int4{Valid: false}
	}

	val := *ptr
	if val > math.MaxInt32 || val < math.MinInt32 {
		return pgtype.Int4{Valid: false}
	}

	return pgtype.Int4{Int32: int32(val), Valid: true}
}

// valOrNil64 returns a pgtype.Int8 with Valid=false if the pointer is nil, otherwise returns the value.
func valOrNil64(ptr *int64) pgtype.Int8 {
	if ptr == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *ptr, Valid: true}
}

// toPgInt4 converts a *int to a pgtype.Int4, using the fallback if the pointer is nil.
func toPgInt4(ptr *int, fallback pgtype.Int4) pgtype.Int4 {
	if ptr == nil {
		return fallback
	}

	val := *ptr
	if val > math.MaxInt32 || val < math.MinInt32 {
		return fallback
	}

	return pgtype.Int4{Int32: int32(val), Valid: true}
}

// toPgInt8 converts a *int64 to a pgtype.Int8, using the fallback if the pointer is nil.
func toPgInt8(ptr *int64, fallback pgtype.Int8) pgtype.Int8 {
	if ptr == nil {
		return fallback
	}
	return pgtype.Int8{Int64: *ptr, Valid: true}
}

// fromPgTimestamptz converts a pgtype.Timestamptz to a *time.Time, returning nil if the value is not valid.
func fromPgTimestamptz(p pgtype.Timestamptz) *time.Time {
	if !p.Valid {
		return nil
	}
	t := p.Time
	return &t
}

// fromPgInt4 converts a pgtype.Int4 to a *int, returning nil if the value is not valid.
func fromPgInt4(p pgtype.Int4) *int {
	if !p.Valid {
		return nil
	}
	v := int(p.Int32)
	return &v
}

// fromPgInt8 converts a pgtype.Int8 to a *int64, returning nil if the value is not valid.
func fromPgInt8(p pgtype.Int8) *int64 {
	if !p.Valid {
		return nil
	}
	v := p.Int64
	return &v
}

// toPgTimestamptz converts a *time.Time to a pgtype.Timestamptz, marking it as invalid if the pointer is nil.
func toPgTimestamptz(t *time.Time, fallback pgtype.Timestamptz) pgtype.Timestamptz {
	if t == nil {
		return fallback
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// safeTime returns the value of the time pointer or zero time if the pointer is nil.
func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// valOrDefault returns the value pointed to by ptr, or def if ptr is nil.
func valOrDefault[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}
