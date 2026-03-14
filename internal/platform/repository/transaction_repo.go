package repository

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

// transactionRepo implements the domain.TransactionRepository interface using a sqlc.Querier for database interactions.
type transactionRepo struct {
	q sqlc.Querier
}

// NewTransactionRepository creates a new TransactionRepository.
func NewTransactionRepository(q sqlc.Querier) domain.TransactionRepository {
	return &transactionRepo{q: q}
}

// Create persists a new transaction for the specified tenant.
func (r *transactionRepo) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	id := ulid.New()

	arg := sqlc.CreateTransactionParams{
		ID:          id,
		TenantID:    tenantID,
		AccountID:   input.AccountID,
		CategoryID:  input.CategoryID,
		UserID:      input.UserID,
		Description: input.Description,
		AmountCents: input.AmountCents,
		Type:        sqlc.TransactionType(input.Type),
		OccurredAt:  pgtype.Timestamptz{Time: input.OccurredAt, Valid: true},
	}

	if input.MasterPurchaseID != "" {
		arg.MasterPurchaseID = pgtype.Text{String: input.MasterPurchaseID, Valid: true}
	}

	row, err := r.q.CreateTransaction(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// GetByID retrieves a specific transaction by its ID and tenant ID.
func (r *transactionRepo) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	row, err := r.q.GetTransactionByID(ctx, sqlc.GetTransactionByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// List returns transactions for a specific tenant with optional filters
// for account, category, type, date range, and pagination.
func (r *transactionRepo) List(ctx context.Context, tenantID string, filter domain.ListTransactionsParams) ([]domain.Transaction, error) {
	arg := sqlc.ListTransactionsByTenantParams{
		TenantID:    tenantID,
		OffsetCount: filter.Offset,
	}

	if filter.AccountID != "" {
		arg.AccountID = pgtype.Text{String: filter.AccountID, Valid: true}
	}
	if filter.CategoryID != "" {
		arg.CategoryID = pgtype.Text{String: filter.CategoryID, Valid: true}
	}
	if filter.Type != "" {
		arg.Type = sqlc.NullTransactionType{TransactionType: sqlc.TransactionType(filter.Type), Valid: true}
	}
	if filter.StartDate != nil {
		arg.StartDate = pgtype.Timestamptz{Time: *filter.StartDate, Valid: true}
	}
	if filter.EndDate != nil {
		arg.EndDate = pgtype.Timestamptz{Time: *filter.EndDate, Valid: true}
	}
	if filter.Limit > 0 {
		arg.LimitCount = filter.Limit
	}

	rows, err := r.q.ListTransactionsByTenant(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", TranslateError(err))
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, *r.mapToDomain(row))
	}

	return transactions, nil
}

// ListByAccount returns transactions for a specific tenant and account
// with optional filters for category, type, date range, and pagination.
func (r *transactionRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	current, err := r.q.GetTransactionByID(ctx, sqlc.GetTransactionByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction for update: %w", TranslateError(err))
	}

	arg := sqlc.UpdateTransactionParams{
		TenantID:    tenantID,
		ID:          id,
		AccountID:   current.AccountID,
		CategoryID:  current.CategoryID,
		Description: current.Description,
		AmountCents: current.AmountCents,
		Type:        current.Type,
		OccurredAt:  current.OccurredAt,
	}

	if input.CategoryID != nil {
		arg.CategoryID = *input.CategoryID
	}
	if input.Description != nil {
		arg.Description = *input.Description
	}
	if input.AmountCents != nil {
		arg.AmountCents = *input.AmountCents
	}
	if input.OccurredAt != nil {
		arg.OccurredAt = pgtype.Timestamptz{Time: *input.OccurredAt, Valid: true}
	}

	row, err := r.q.UpdateTransaction(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", TranslateError(err))
	}

	return r.mapToDomain(row), nil
}

// Delete performs a soft delete of the transaction by its ID and tenant ID.
func (r *transactionRepo) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.SoftDeleteTransaction(ctx, sqlc.SoftDeleteTransactionParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", TranslateError(err))
	}
	return nil
}

// mapToDomain converts a sqlc.Transaction to a domain.Transaction, handling nullable fields appropriately.
func (r *transactionRepo) mapToDomain(row sqlc.Transaction) *domain.Transaction {
	var masterPurchaseID string
	if row.MasterPurchaseID.Valid {
		masterPurchaseID = row.MasterPurchaseID.String
	}

	return &domain.Transaction{
		ID:               row.ID,
		TenantID:         row.TenantID,
		AccountID:        row.AccountID,
		CategoryID:       row.CategoryID,
		UserID:           row.UserID,
		MasterPurchaseID: masterPurchaseID,
		Description:      row.Description,
		AmountCents:      row.AmountCents,
		Type:             domain.TransactionType(row.Type),
		OccurredAt:       row.OccurredAt.Time,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		DeletedAt:        nil,
	}
}
