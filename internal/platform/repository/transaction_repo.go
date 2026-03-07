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

func (r *transactionRepo) List(ctx context.Context, tenantID string, filter domain.ListTransactionsParams) ([]domain.Transaction, error) {
	rows, err := r.q.ListTransactionsByTenant(ctx, tenantID)
	if err != nil {
		return nil, r.translateError(err)
	}

	transactions := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		transactions = append(transactions, *r.mapToDomain(row))
	}

	return transactions, nil
}

func (r *transactionRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	current, err := r.q.GetTransactionByID(ctx, sqlc.GetTransactionByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, r.translateError(err)
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

func (r *transactionRepo) translateError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return domain.ErrConflict
		}
	}
	return err
}

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
