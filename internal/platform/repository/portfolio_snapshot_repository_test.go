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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPortfolioSnapshotRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tenantID := "01H2PJZ6X7R8K9M4VQB5A7Y2Z1"
	now := time.Now().UTC()

	t.Run("Create", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPortfolioSnapshotRepository(mQuerier)

		input := domain.CreatePortfolioSnapshotInput{
			SnapshotDate: now,
			DetailsJSON:  []byte(`{"total_assets": 1000}`),
		}

		mQuerier.On("CreatePortfolioSnapshot", ctx, mock.MatchedBy(func(arg sqlc.CreatePortfolioSnapshotParams) bool {
			return arg.TenantID == tenantID &&
				arg.SnapshotDate.Time.Equal(input.SnapshotDate) &&
				string(arg.Details) == string(input.DetailsJSON)
		})).Return(sqlc.PortfolioSnapshot{
			ID:           "SNAP_1",
			TenantID:     tenantID,
			SnapshotDate: pgtype.Date{Time: input.SnapshotDate, Valid: true},
			Details:      input.DetailsJSON,
			CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		snap, err := repo.Create(ctx, tenantID, input)
		require.NoError(t, err)
		require.NotNil(t, snap)
		assert.Equal(t, "SNAP_1", snap.ID)
		assert.Equal(t, tenantID, snap.TenantID)
		assert.True(t, snap.SnapshotDate.Equal(input.SnapshotDate))
		assert.Equal(t, input.DetailsJSON, snap.DetailsJSON)
	})

	t.Run("Create Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPortfolioSnapshotRepository(mQuerier)

		mQuerier.On("CreatePortfolioSnapshot", ctx, mock.Anything).
			Return(sqlc.PortfolioSnapshot{}, fmt.Errorf("db error"))

		snap, err := repo.Create(ctx, tenantID, domain.CreatePortfolioSnapshotInput{})
		require.Error(t, err)
		assert.Nil(t, snap)
		assert.Contains(t, err.Error(), "failed to create portfolio snapshot")
	})

	t.Run("GetByDate", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPortfolioSnapshotRepository(mQuerier)

		date := now.Truncate(24 * time.Hour)

		mQuerier.On("GetPortfolioSnapshotByDate", ctx, sqlc.GetPortfolioSnapshotByDateParams{
			SnapshotDate: pgtype.Date{Time: date, Valid: true},
			TenantID:     tenantID,
		}).Return(sqlc.PortfolioSnapshot{
			ID:           "SNAP_1",
			TenantID:     tenantID,
			SnapshotDate: pgtype.Date{Time: date, Valid: true},
			Details:      []byte(`{}`),
			CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

		snap, err := repo.GetByDate(ctx, tenantID, date)
		require.NoError(t, err)
		require.NotNil(t, snap)
		assert.Equal(t, "SNAP_1", snap.ID)
	})

	t.Run("GetByDate Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPortfolioSnapshotRepository(mQuerier)

		mQuerier.On("GetPortfolioSnapshotByDate", ctx, mock.Anything).
			Return(sqlc.PortfolioSnapshot{}, fmt.Errorf("db error"))

		snap, err := repo.GetByDate(ctx, tenantID, now)
		require.Error(t, err)
		assert.Nil(t, snap)
		assert.Contains(t, err.Error(), "failed to get portfolio snapshot")
	})

	t.Run("ListByTenant", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPortfolioSnapshotRepository(mQuerier)

		mQuerier.On("ListPortfolioSnapshots", ctx, tenantID).Return([]sqlc.PortfolioSnapshot{
			{
				ID:           "SNAP_1",
				TenantID:     tenantID,
				SnapshotDate: pgtype.Date{Time: now, Valid: true},
				CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			},
			{
				ID:           "SNAP_2",
				TenantID:     tenantID,
				SnapshotDate: pgtype.Date{Time: now.AddDate(0, 0, -1), Valid: true},
				CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
			},
		}, nil)

		list, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, "SNAP_1", list[0].ID)
		assert.Equal(t, "SNAP_2", list[1].ID)
	})

	t.Run("ListByTenant Error", func(t *testing.T) {
		t.Parallel()
		mQuerier := new(mocks.Querier)
		repo := NewPortfolioSnapshotRepository(mQuerier)

		mQuerier.On("ListPortfolioSnapshots", ctx, tenantID).
			Return(nil, fmt.Errorf("db error"))

		list, err := repo.ListByTenant(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "failed to list portfolio snapshots")
	})
}
