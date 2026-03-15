//go:build integration

package seeds

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// SeedUser inserts a test user for the given tenantID.
// Defaults: role=member, name="Test User", unique email.
func SeedUser(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID string) domain.User {
	t.Helper()

	id := ulid.New()
	row, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       id,
		TenantID: tenantID,
		Email:    fmt.Sprintf("user-%s@example.com", id),
		Name:     "Test User",
		Role:     sqlc.UserRoleMember,
	})
	require.NoError(t, err)

	return domain.User{
		ID:          row.ID,
		TenantID:    row.TenantID,
		Email:       row.Email,
		Name:        row.Name,
		Role:        domain.Role(row.Role),
		LastLoginAt: &row.LastLoginAt.Time,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
