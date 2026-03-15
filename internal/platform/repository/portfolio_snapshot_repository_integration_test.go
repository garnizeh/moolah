//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/internal/testutil/seeds"
	"github.com/stretchr/testify/require"
)

func TestPortfolioSnapshotRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)
	repo := repository.NewPortfolioSnapshotRepository(db.Queries)

	// Seed tenant
	tenant := seeds.SeedTenant(t, ctx, db.Queries)

	t.Run("Create & Get By Date", func(t *testing.T) {
		t.Parallel()
		date := time.Now().Truncate(24 * time.Hour)
		in := domain.CreatePortfolioSnapshotInput{
			SnapshotDate: date,
			DetailsJSON:  []byte(`{"total": 1000}`),
		}
		snap, err := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, err)
		require.NotEmpty(t, snap.ID)

		fetched, err := repo.GetByDate(ctx, tenant.ID, date)
		require.NoError(t, err)
		require.Equal(t, snap.ID, fetched.ID)
	})

	t.Run("List By Tenant", func(t *testing.T) {
		t.Parallel()
		in := domain.CreatePortfolioSnapshotInput{
			SnapshotDate: time.Now().Add(-48 * time.Hour).Truncate(24 * time.Hour),
			DetailsJSON:  []byte(`{"agg": "history"}`),
		}
		_, err := repo.Create(ctx, tenant.ID, in)
		require.NoError(t, err)

		list, err := repo.ListByTenant(ctx, tenant.ID)
		require.NoError(t, err)
		require.NotEmpty(t, list)
	})
}
