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

// SeedMasterPurchase inserts a master purchase with given parameters.
func SeedMasterPurchase(
	t *testing.T,
	ctx context.Context,
	q sqlc.Querier,
	tenantID, accountID, categoryID, userID string,
	totalCents int64,
	installments int32,
	firstDue time.Time,
) domain.MasterPurchase {
	t.Helper()

	id := ulid.New()
	// #nosec G115
	row, err := q.CreateMasterPurchase(ctx, sqlc.CreateMasterPurchaseParams{
		ID:                   id,
		TenantID:             tenantID,
		AccountID:            accountID,
		CategoryID:           categoryID,
		UserID:               userID,
		Description:          "Test Master Purchase",
		TotalAmountCents:     totalCents,
		InstallmentCount:     int16(installments),
		ClosingDay:           int16(firstDue.Day()),
		FirstInstallmentDate: pgtype.Date{Time: firstDue, Valid: true},
	})
	if err != nil {
		t.Fatalf(
			"failed to create master purchase (tenant_id=%q, account_id=%q, category_id=%q, user_id=%q): %v",
			tenantID,
			accountID,
			categoryID,
			userID,
			err,
		)
	}

	return domain.MasterPurchase{
		ID:                   row.ID,
		TenantID:             row.TenantID,
		AccountID:            row.AccountID,
		CategoryID:           row.CategoryID,
		UserID:               row.UserID,
		Description:          row.Description,
		TotalAmountCents:     row.TotalAmountCents,
		InstallmentCount:     int32(row.InstallmentCount),
		PaidInstallments:     int32(row.PaidInstallments),
		Status:               domain.MasterPurchaseStatus(row.Status),
		FirstInstallmentDate: row.FirstInstallmentDate.Time,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		ClosingDay:           int32(row.ClosingDay),
	}
}

// UpdateMasterPurchasePaidInstallments forcefully updates the paid_installments count for a record.
func UpdateMasterPurchasePaidInstallments(
	t *testing.T,
	ctx context.Context,
	q sqlc.Querier,
	tenantID string,
	id string,
	paid int32,
	status domain.MasterPurchaseStatus,
) {
	t.Helper()

	current, err := q.GetMasterPurchaseByID(ctx, sqlc.GetMasterPurchaseByIDParams{
		ID:       id,
		TenantID: tenantID,
	})
	require.NoError(t, err)

	// #nosec G115
	_, err = q.UpdateMasterPurchase(ctx, sqlc.UpdateMasterPurchaseParams{
		ID:         id,
		TenantID:   tenantID,
		CategoryID: current.CategoryID,
		PaidInstallments: pgtype.Int2{
			Int16: int16(paid),
			Valid: true,
		},
		Status: sqlc.NullMasterPurchaseStatus{
			MasterPurchaseStatus: sqlc.MasterPurchaseStatus(status),
			Valid:                true,
		},
	})
	require.NoError(t, err)
}
