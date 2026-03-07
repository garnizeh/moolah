//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/testutil/containers"
)

func TestAuditRepo_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := containers.NewPostgresDB(t)

	tenantRepo := repository.NewTenantRepository(db.Queries)
	repo := repository.NewAuditRepository(db.Queries)

	// Setup: Create tenant
	tenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Audit Tenant"})
	require.NoError(t, err)

	t.Run("Create and ListByEntity", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateAuditLogInput{
			TenantID:   tenant.ID,
			ActorID:    "actor-1",
			ActorRole:  domain.RoleAdmin,
			Action:     domain.AuditActionUpdate,
			EntityType: "account",
			EntityID:   "acc-1",
			OldValues:  []byte(`{"name":"Old"}`),
			NewValues:  []byte(`{"name":"New"}`),
			IPAddress:  "127.0.0.1",
			UserAgent:  "Mozilla/5.0",
		}

		created, err := repo.Create(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, input.ActorID, created.ActorID)
		require.Equal(t, input.Action, created.Action)

		logs, err := repo.ListByEntity(ctx, tenant.ID, "account", "acc-1")
		require.NoError(t, err)
		require.Len(t, logs, 1)
		require.Equal(t, created.ID, logs[0].ID)
	})

	t.Run("ListByTenant Isolation", func(t *testing.T) {
		t.Parallel()
		isolationTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Isolation T"})
		require.NoError(t, err)

		otherTenant, err := tenantRepo.Create(ctx, domain.CreateTenantInput{Name: "Other T"})
		require.NoError(t, err)

		input := domain.CreateAuditLogInput{
			TenantID:   isolationTenant.ID,
			ActorID:    "actor-1",
			ActorRole:  domain.RoleAdmin,
			Action:     domain.AuditActionCreate,
			EntityType: "category",
			EntityID:   "cat-1",
		}
		_, err = repo.Create(ctx, input)
		require.NoError(t, err)

		// List for correct tenant
		logs, err := repo.ListByTenant(ctx, isolationTenant.ID, domain.ListAuditLogsParams{})
		require.NoError(t, err)
		require.NotEmpty(t, logs)

		// List for other tenant
		logs, err = repo.ListByTenant(ctx, otherTenant.ID, domain.ListAuditLogsParams{})
		require.NoError(t, err)
		require.Empty(t, logs)
	})
}
