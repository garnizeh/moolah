package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPositionSnapshotRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z1"
	positionID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z2"
	now := time.Now().UTC()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionSnapshotRepository(mQuerier)
		qty := decimal.NewFromInt(10)

		input := domain.CreatePositionSnapshotInput{
			PositionID:     positionID,
			SnapshotDate:   now,
			Quantity:       qty,
			LastPriceCents: 1000,
			Currency:       "BRL",
		}

		mQuerier.On("CreatePositionSnapshot", ctx, mock.MatchedBy(func(arg sqlc.CreatePositionSnapshotParams) bool {
			return arg.TenantID == tenantID &&
				arg.PositionID == positionID &&
				arg.SnapshotDate.Time.Equal(input.SnapshotDate) &&
				arg.Quantity.Equal(input.Quantity) &&
				arg.LastPriceCents == input.LastPriceCents
		})).Return(sqlc.PositionSnapshot{
			ID:             "SNAP_1",
			TenantID:       tenantID,
			PositionID:     positionID,
			SnapshotDate:   pgtype.Date{Time: input.SnapshotDate, Valid: true},
			Quantity:       qty,
			LastPriceCents: 1000,
			Currency:       "BRL",
			CreatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		snap, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		require.NotNil(t, snap)
		assert.Equal(t, "SNAP_1", snap.ID)
		assert.Equal(t, tenantID, snap.TenantID)
		assert.True(t, snap.Quantity.Equal(qty))
	})

	t.Run("Create Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionSnapshotRepository(mQuerier)

		mQuerier.On("CreatePositionSnapshot", ctx, mock.Anything).
			Return(sqlc.PositionSnapshot{}, fmt.Errorf("db error"))

		snap, err := repo.Create(ctx, tenantID, domain.CreatePositionSnapshotInput{})
		require.Error(t, err)
		assert.Nil(t, snap)
		assert.Contains(t, err.Error(), "failed to create position snapshot")
	})

	t.Run("ListByPosition", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionSnapshotRepository(mQuerier)

		mQuerier.On("ListPositionSnapshotsByPosition", ctx, sqlc.ListPositionSnapshotsByPositionParams{
			PositionID: positionID,
			TenantID:   tenantID,
		}).Return([]sqlc.PositionSnapshot{
			{ID: "S1", TenantID: tenantID, PositionID: positionID},
			{ID: "S2", TenantID: tenantID, PositionID: positionID},
		}, nil)

		list, err := repo.ListByPosition(ctx, tenantID, positionID)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("ListByPosition Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionSnapshotRepository(mQuerier)

		mQuerier.On("ListPositionSnapshotsByPosition", ctx, mock.Anything).
			Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListByPosition(ctx, tenantID, "any")
		require.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to list position snapshots")
	})

	t.Run("ListByTenantSince", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionSnapshotRepository(mQuerier)

		since := now.AddDate(0, 0, -7)
		mQuerier.On("ListPositionSnapshotsByTenantSince", ctx, sqlc.ListPositionSnapshotsByTenantSinceParams{
			TenantID:     tenantID,
			SnapshotDate: pgtype.Date{Time: since, Valid: true},
		}).Return([]sqlc.PositionSnapshot{
			{ID: "S1", TenantID: tenantID},
		}, nil)

		list, err := repo.ListByTenantSince(ctx, tenantID, since)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("ListByTenantSince Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPositionSnapshotRepository(mQuerier)

		mQuerier.On("ListPositionSnapshotsByTenantSince", ctx, mock.Anything).
			Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListByTenantSince(ctx, tenantID, now)
		require.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to list tenant snapshots")
	})
}
