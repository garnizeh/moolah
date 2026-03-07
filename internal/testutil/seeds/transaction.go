//go:build integration

package seeds

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// CreateTransaction inserts a minimal expense transaction.
// Defaults: type=expense, amount_cents=100, occurred_at=time.Now().
func CreateTransaction(
	t *testing.T,
	ctx context.Context,
	q sqlc.Querier,
	tenantID, accountID, categoryID, userID string,
) domain.Transaction {
	t.Helper()

	now := time.Now()
	row, err := q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:         ulid.New(),
		TenantID:   tenantID,
		AccountID:  accountID,
		CategoryID: categoryID,
		UserID:     userID,
		MasterPurchaseID: pgtype.Text{
			String: "",
			Valid:  false,
		},
		Description: "Test Transaction",
		AmountCents: 100,
		Type:        sqlc.TransactionTypeExpense,
		OccurredAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	})
	require.NoError(t, err)

	return domain.Transaction{
		ID:               row.ID,
		TenantID:         row.TenantID,
		AccountID:        row.AccountID,
		CategoryID:       row.CategoryID,
		UserID:           row.UserID,
		MasterPurchaseID: row.MasterPurchaseID.String,
		Description:      row.Description,
		AmountCents:      row.AmountCents,
		Type:             domain.TransactionType(row.Type),
		OccurredAt:       row.OccurredAt.Time,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
