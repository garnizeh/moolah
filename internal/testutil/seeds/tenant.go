//go:build integration

package seeds

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// SeedTenant inserts a test tenant and returns the mapped domain.Tenant.
// Defaults: plan=free, name="Test Household".
func SeedTenant(t *testing.T, ctx context.Context, q sqlc.Querier) domain.Tenant {
	t.Helper()

	row, err := q.CreateTenant(ctx, sqlc.CreateTenantParams{
		ID:   ulid.New(),
		Name: "Test Household",
		Plan: sqlc.TenantPlanFree,
	})
	require.NoError(t, err)

	return domain.Tenant{
		ID:        row.ID,
		Name:      row.Name,
		Plan:      domain.TenantPlan(row.Plan),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
