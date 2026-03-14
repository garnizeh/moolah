package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

// MasterPurchaseRepository provides methods to manage master purchases in the database.
type masterPurchaseRepo struct {
	q sqlc.Querier
}

// NewMasterPurchaseRepository creates a new MasterPurchaseRepository implementation.
func NewMasterPurchaseRepository(q sqlc.Querier) domain.MasterPurchaseRepository {
	return &masterPurchaseRepo{q: q}
}

// Create adds a new master purchase to the database based on the provided input and returns the created master purchase.
func (r *masterPurchaseRepo) Create(ctx context.Context, tenantID string, input domain.CreateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	id := ulid.New()

	arg := sqlc.CreateMasterPurchaseParams{
		ID:               id,
		TenantID:         tenantID,
		AccountID:        input.AccountID,
		CategoryID:       input.CategoryID,
		UserID:           input.UserID,
		Description:      input.Description,
		TotalAmountCents: input.TotalAmountCents,
		// #nosec G115
		InstallmentCount: int16(input.InstallmentCount),
		// #nosec G115
		ClosingDay:           int16(input.ClosingDay),
		FirstInstallmentDate: pgtype.Date{Time: input.FirstInstallmentDate, Valid: true},
	}

	row, err := r.q.CreateMasterPurchase(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create master purchase: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// GetByID retrieves a master purchase by its unique ID and tenant ID.
// If the master purchase does not exist, it returns a domain.ErrMasterPurchaseNotFound error.
func (r *masterPurchaseRepo) GetByID(ctx context.Context, tenantID, id string) (*domain.MasterPurchase, error) {
	row, err := r.q.GetMasterPurchaseByID(ctx, sqlc.GetMasterPurchaseByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get master purchase: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// ListByTenant returns all master purchases for the given tenant.
func (r *masterPurchaseRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.MasterPurchase, error) {
	rows, err := r.q.ListMasterPurchasesByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list master purchases: %w", TranslateError(err))
	}

	purchases := make([]domain.MasterPurchase, 0, len(rows))
	for _, row := range rows {
		purchases = append(purchases, *r.mapToDomain(row))
	}

	return purchases, nil
}

// ListByAccount returns all master purchases for the given tenant and account.
func (r *masterPurchaseRepo) ListByAccount(ctx context.Context, tenantID, accountID string) ([]domain.MasterPurchase, error) {
	rows, err := r.q.ListMasterPurchasesByAccount(ctx, sqlc.ListMasterPurchasesByAccountParams{
		TenantID:  tenantID,
		AccountID: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list master purchases by account: %w", TranslateError(err))
	}

	purchases := make([]domain.MasterPurchase, 0, len(rows))
	for _, row := range rows {
		purchases = append(purchases, *r.mapToDomain(row))
	}

	return purchases, nil
}

// ListPendingClose returns all master purchases that are pending and have a closing day on or before the specified cutoff date.
func (r *masterPurchaseRepo) ListPendingClose(ctx context.Context, tenantID string, cutoffDate time.Time) ([]domain.MasterPurchase, error) {
	rows, err := r.q.ListPendingMasterPurchasesByClosingDay(ctx, sqlc.ListPendingMasterPurchasesByClosingDayParams{
		TenantID: tenantID,
		// #nosec G115
		ClosingDay: int16(cutoffDate.Day()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pending master purchases: %w", TranslateError(err))
	}

	purchases := make([]domain.MasterPurchase, 0, len(rows))
	for _, row := range rows {
		purchases = append(purchases, *r.mapToDomain(row))
	}

	return purchases, nil
}

// Update modifies an existing master purchase with the provided input fields. Only non-nil fields in the input will be updated.
func (r *masterPurchaseRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateMasterPurchaseInput) (*domain.MasterPurchase, error) {
	arg := sqlc.UpdateMasterPurchaseParams{
		TenantID: tenantID,
		ID:       id,
	}

	if input.CategoryID != nil {
		arg.CategoryID = *input.CategoryID
	}
	if input.Description != nil {
		arg.Description = pgtype.Text{String: *input.Description, Valid: true}
	}

	row, err := r.q.UpdateMasterPurchase(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update master purchase: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// IncrementPaidInstallments increments the count of paid installments for a master purchase by 1.
func (r *masterPurchaseRepo) IncrementPaidInstallments(ctx context.Context, tenantID, id string) error {
	_, err := r.q.IncrementPaidInstallments(ctx, sqlc.IncrementPaidInstallmentsParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to increment paid installments: %w", TranslateError(err))
	}

	return nil
}

// Delete performs a soft delete of the master purchase by its ID and tenant ID.
func (r *masterPurchaseRepo) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.DeleteMasterPurchase(ctx, sqlc.DeleteMasterPurchaseParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete master purchase: %w", TranslateError(err))
	}

	return nil
}

// mapToDomain converts a sqlc.MasterPurchase to a domain.MasterPurchase, handling nullable fields appropriately.
func (r *masterPurchaseRepo) mapToDomain(row sqlc.MasterPurchase) *domain.MasterPurchase {
	return &domain.MasterPurchase{
		ID:                   row.ID,
		TenantID:             row.TenantID,
		AccountID:            row.AccountID,
		CategoryID:           row.CategoryID,
		UserID:               row.UserID,
		Description:          row.Description,
		TotalAmountCents:     row.TotalAmountCents,
		InstallmentCount:     int32(row.InstallmentCount),
		PaidInstallments:     int32(row.PaidInstallments),
		ClosingDay:           int32(row.ClosingDay),
		Status:               domain.MasterPurchaseStatus(row.Status),
		FirstInstallmentDate: row.FirstInstallmentDate.Time,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
}
