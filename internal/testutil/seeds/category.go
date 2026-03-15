//go:build integration

package seeds

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// SeedCategory inserts a root expense category within tenantID.
// Defaults: type=expense, no parent, name="Test Category".
func SeedCategory(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID string) domain.Category {
	t.Helper()

	row, err := q.CreateCategory(ctx, sqlc.CreateCategoryParams{
		ID:       ulid.New(),
		TenantID: tenantID,
		ParentID: pgtype.Text{}, // Root category
		Name:     "Test Category",
		Icon:     pgtype.Text{String: "🍔", Valid: true},
		Color:    pgtype.Text{String: "#FF0000", Valid: true},
		Type:     sqlc.CategoryTypeExpense,
	})
	require.NoError(t, err)

	return domain.Category{
		ID:        row.ID,
		TenantID:  row.TenantID,
		ParentID:  row.ParentID.String,
		Name:      row.Name,
		Icon:      row.Icon.String,
		Color:     row.Color.String,
		Type:      domain.CategoryType(row.Type),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
