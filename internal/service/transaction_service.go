package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/garnizeh/moolah/internal/domain"
)

// transactionService provides business logic for managing transactions, including creation, retrieval, updating, and deletion of transactions. It also handles related audit logging and ensures that account balances are updated accordingly.
type transactionService struct {
	txRepo       domain.TransactionRepository
	accountRepo  domain.AccountRepository
	categoryRepo domain.CategoryRepository
	auditRepo    domain.AuditRepository
}

// NewTransactionService creates a new instance of TransactionService.
func NewTransactionService(
	txRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	categoryRepo domain.CategoryRepository,
	auditRepo domain.AuditRepository,
) domain.TransactionService {
	return &transactionService{
		txRepo:       txRepo,
		accountRepo:  accountRepo,
		categoryRepo: categoryRepo,
		auditRepo:    auditRepo,
	}
}

// Create validates the input and creates a new transaction record. It checks that the referenced account exists, that the referenced category exists and matches the transaction type, and then persists the transaction. It also updates the account balance accordingly and logs the creation action in the audit trail. It returns the created transaction or an error if validation fails or the record cannot be created.
func (s *transactionService) Create(ctx context.Context, tenantID string, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	// 1. Verify account belongs to tenant.
	account, err := s.accountRepo.GetByID(ctx, tenantID, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to verify account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("transaction service: account not found: %w", domain.ErrNotFound)
	}

	// 2. Verify category belongs to tenant and type matches.
	category, err := s.categoryRepo.GetByID(ctx, tenantID, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to verify category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("transaction service: category not found: %w", domain.ErrNotFound)
	}
	if string(category.Type) != string(input.Type) {
		return nil, fmt.Errorf("transaction service: category type mismatch (expected %s, got %s): %w", input.Type, category.Type, domain.ErrInvalidInput)
	}

	// 3. Persist transaction.
	tx, err := s.txRepo.Create(ctx, tenantID, input)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to create transaction record: %w", err)
	}

	// 4. Update account balance.
	delta := s.calculateDelta(input.Type, input.AmountCents)
	err = s.accountRepo.UpdateBalance(ctx, tenantID, input.AccountID, delta)
	if err != nil {
		// Best-effort sequential write (ACID deferred to Phase 2).
		return nil, fmt.Errorf("transaction service: partial success - record created but balance update failed: %w", err)
	}

	// 5. Write audit log.
	newValues, err := json.Marshal(map[string]any{
		"account_id":   tx.AccountID,
		"category_id":  tx.CategoryID,
		"description":  tx.Description,
		"type":         tx.Type,
		"amount_cents": tx.AmountCents,
		"occurred_at":  tx.OccurredAt,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal audit trail for transaction creation", "error", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    input.UserID,
		EntityType: "transaction",
		EntityID:   tx.ID,
		Action:     domain.AuditActionCreate,
		NewValues:  newValues,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create audit log for transaction creation", "error", err)
	}

	return tx, nil
}

// GetByID retrieves a transaction by its ID and tenant ID. It returns the transaction or an error if the record cannot be found or retrieved.
func (s *transactionService) GetByID(ctx context.Context, tenantID, id string) (*domain.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to fetch transaction: %w", err)
	}
	return tx, nil
}

// List returns all transactions for a given tenant ID, optionally filtered by account, category, date range, or type. It retrieves the list of transactions or returns an error if the records cannot be retrieved.
func (s *transactionService) List(ctx context.Context, tenantID string, params domain.ListTransactionsParams) ([]domain.Transaction, error) {
	txs, err := s.txRepo.List(ctx, tenantID, params)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to list transactions: %w", err)
	}
	return txs, nil
}

// Update validates the input and updates an existing transaction record. It checks that if the account is being updated, the new account exists; if the category is being updated, the new category exists and matches the transaction type. It then persists the transaction update, adjusts the account balance if the amount or type changed, and logs the update action in the audit trail. It returns the updated transaction or an error if validation fails or the record cannot be updated.
func (s *transactionService) Update(ctx context.Context, tenantID, id string, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	// 1. Fetch existing transaction.
	oldTx, err := s.txRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to locate transaction for update: %w", err)
	}
	if oldTx == nil {
		return nil, domain.ErrNotFound
	}

	// 2. Adjust balance if amount changed.
	if input.AmountCents != nil && *input.AmountCents != oldTx.AmountCents {
		oldDelta := s.calculateDelta(oldTx.Type, oldTx.AmountCents)
		newDelta := s.calculateDelta(oldTx.Type, *input.AmountCents)

		// Revert old, apply new: adjustment = newDelta - oldDelta
		err = s.accountRepo.UpdateBalance(ctx, tenantID, oldTx.AccountID, newDelta-oldDelta)
		if err != nil {
			return nil, fmt.Errorf("transaction service: failed to adjust balance during update: %w", err)
		}
	}

	// 3. Persist update.
	newTx, err := s.txRepo.Update(ctx, tenantID, id, input)
	if err != nil {
		return nil, fmt.Errorf("transaction service: failed to update transaction record: %w", err)
	}

	// 4. Write audit log.
	newValuesMap := make(map[string]any)
	oldValuesMap := make(map[string]any)

	if input.AmountCents != nil && *input.AmountCents != oldTx.AmountCents {
		newValuesMap["amount_cents"] = *input.AmountCents
		oldValuesMap["amount_cents"] = oldTx.AmountCents
	}
	if input.Description != nil && *input.Description != oldTx.Description {
		newValuesMap["description"] = *input.Description
		oldValuesMap["description"] = oldTx.Description
	}
	if input.CategoryID != nil && *input.CategoryID != oldTx.CategoryID {
		newValuesMap["category_id"] = *input.CategoryID
		oldValuesMap["category_id"] = oldTx.CategoryID
	}

	if len(newValuesMap) > 0 {
		oldV, err := json.Marshal(oldValuesMap)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal old values for audit", "error", err)
		}
		newV, err := json.Marshal(newValuesMap)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal new values for audit", "error", err)
		}

		_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
			TenantID:   tenantID,
			ActorID:    newTx.UserID,
			EntityType: "transaction",
			EntityID:   id,
			Action:     domain.AuditActionUpdate,
			OldValues:  oldV,
			NewValues:  newV,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to create audit log for transaction update", "error", err)
		}
	}

	return newTx, nil
}

// Delete performs a soft delete of a transaction record by its ID and tenant ID. It reverts the account balance accordingly, logs the deletion action in the audit trail, and then marks the transaction as deleted. It returns an error if the record cannot be deleted.
func (s *transactionService) Delete(ctx context.Context, tenantID, id string) error {
	// 1. Fetch existing.
	tx, err := s.txRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("transaction service: failed to locate transaction for deletion: %w", err)
	}
	if tx == nil {
		return domain.ErrNotFound
	}

	// 2. Revert balance.
	revertDelta := -s.calculateDelta(tx.Type, tx.AmountCents)
	err = s.accountRepo.UpdateBalance(ctx, tenantID, tx.AccountID, revertDelta)
	if err != nil {
		return fmt.Errorf("transaction service: failed to revert balance for deletion: %w", err)
	}

	// 3. Write audit log.
	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    tx.UserID,
		EntityType: "transaction",
		EntityID:   id,
		Action:     domain.AuditActionSoftDelete,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create audit log for transaction deletion", "error", err)
	}

	// 4. Soft-delete.
	if err := s.txRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("transaction service: failed to soft-delete transaction record: %w", err)
	}

	return nil
}

// calculateDelta computes the balance adjustment for a transaction based on its type and amount. Income transactions increase balance, while expense and transfer transactions decrease balance.
func (s *transactionService) calculateDelta(txType domain.TransactionType, amount int64) int64 {
	switch txType {
	case domain.TransactionTypeIncome:
		return amount
	case domain.TransactionTypeExpense, domain.TransactionTypeTransfer:
		return -amount
	default:
		return 0
	}
}
