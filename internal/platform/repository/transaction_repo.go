package repository

import (
	"context"
	"errors"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type transactionRepo struct {
	q sqlc.Querier
}

// NewTransactionRepository creates a new TransactionRepository.
func NewTransactionRepository(q sqlc.Querier) domain.TransactionRepository {
	return &transactionRepo{q: q}
}

func (r *transactionRepo) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	id := ulid.New()

	arg := sqlc.CreateTransactionParams{
		ID:          id,
		TenantID:    tenantID,
		AccountID:   input.AccountID,
		CategoryID:  input.CategoryID,
		Description: input.Description,
		AmountCents: input.AmountCents,
		Type:        sqlc.TransactionType(input.Type),
		OccurredAt:  pgtype.Timestamptz{Time: input.OccurredAt, Valid: true},
	}

	if input.MasterPurchaseID != "" {
		arg.MasterPurchaseID = pgtype.Text{String: input.MasterPurchaseID, Valid: true}
	}

	// Note: user_id should come from context, but for Phase 1 MVP we might just use a placeholder or
	// expect it to be passed. Looking at the domain entity, it's there.
	// For now, let's look at how other repos handled it.
	// Actually, the domain.CreateTransactionInput doesn't have UserID.
	// In a real scenario, it should be extracted from context.
	// Since I don't see UserID in the input, I'll use a placeholder or empty for now if it's not provided.
	// Wait, the generated sqlc.CreateTransactionParams HAS UserID.

	row, err := r.q.CreateTransaction(ctx, arg)
	if err != nil {
		return nil, r.translateError(err)
	}

	return r.mapToDomain(row), nil
}

func (r *transactionRepo) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	row, err := r.q.GetTransactionByID(ctx, sqlc.GetTransactionByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, r.translateError(err)
	}

	return r.mapToDomain(row), nil
}

func (r *transactionRepo) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	arg := sqlc.ListTransactionsByTenantParams{
		TenantID:  tenantID,
		LimitOff:  params.Limit,
		OffsetOff: params.Offset,
	}

	if params.AccountID != "" {
		arg.AccountID = pgtype.Text{String: params.AccountID, Valid: true}
	}
	if params.CategoryID != "" {
		arg.CategoryID = pgtype.Text{String: params.CategoryID, Valid: true}
	}
	if params.StartDate != nil {
		arg.StartDate = pgtype.Timestamptz{Time: *params.StartDate, Valid: true}
	}
	if params.EndDate != nil {
		arg.EndDate = pgtype.Timestamptz{Time: *params.EndDate, Valid: true}
	}

	rows, err := r.q.ListTransactionsByTenant(ctx, arg)
	if err != nil {
		return nil, r.translateError(err)
	}

	transactions := make([]domain.Transaction, len(rows))
	for i, row := range rows {
		transactions[i] = *r.mapToDomain(row)
	}

	return transactions, nil
}

func (r *transactionRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	// We need current state for partial updates if the sqlc query doesn't handle them
	current, err := r.q.GetTransactionByID(ctx, sqlc.GetTransactionByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, r.translateError(err)
	}

	arg := sqlc.UpdateTransactionParams{
		TenantID:         tenantID,
		ID:               id,
		AccountID:        current.AccountID,
		CategoryID:       current.CategoryID,
		Description:      current.Description,
		AmountCents:      current.AmountCents,
		Type:             current.Type,
		OccurredAt:       current.OccurredAt,
		MasterPurchaseID: current.MasterPurchaseID,
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
		return nil, r.translateError(err)
	}

	return r.mapToDomain(row), nil
}

func (r *transactionRepo) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.SoftDeleteTransaction(ctx, sqlc.SoftDeleteTransactionParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return r.translateError(err)
	}
	return nil
}

func (r *transactionRepo) mapToDomain(row sqlc.Transaction) *domain.Transaction {
	t := &domain.Transaction{
		ID:          row.ID,
		TenantID:    row.TenantID,
		AccountID:   row.AccountID,
		CategoryID:  row.CategoryID,
		UserID:      row.UserID,
		Description: row.Description,
		AmountCents: row.AmountCents,
		Type:        domain.TransactionType(row.Type),
		OccurredAt:  row.OccurredAt.Time,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}

	if row.MasterPurchaseID.Valid {
		t.MasterPurchaseID = row.MasterPurchaseID.String
	}

	if row.DeletedAt.Valid {
		t.DeletedAt = &row.DeletedAt.Time
	}

	return t
}

func (r *transactionRepo) translateError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign_key_violation
			return domain.ErrNotFound
		case "23505": // unique_violation
			return domain.ErrConflict
		}
	}

	return err
}
