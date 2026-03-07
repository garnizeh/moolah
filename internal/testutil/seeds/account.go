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

// CreateAccount inserts a test checking account owned by userID within tenantID.
// Defaults: type=checking, currency=BRL, balance_cents=0.
func CreateAccount(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID, userID string) domain.Account {
	t.Helper()

	row, err := q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           ulid.New(),
		TenantID:     tenantID,
		UserID:       userID,
		Name:         "Test Checking Account",
		Type:         sqlc.AccountTypeChecking,
		Currency:     "BRL",
		BalanceCents: 0,
	})
	require.NoError(t, err)

	return domain.Account{
		ID:           row.ID,
		TenantID:     row.TenantID,
		UserID:       row.UserID,
		Name:         row.Name,
		Type:         domain.AccountType(row.Type),
		Currency:     row.Currency,
		BalanceCents: row.BalanceCents,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
