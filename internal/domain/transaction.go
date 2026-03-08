package domain

import (
	"context"
	"time"
)

// TransactionType mirrors the database enum for transaction types.
type TransactionType string

const (
	// TransactionTypeIncome representing income transactions (e.g. salary).
	TransactionTypeIncome TransactionType = "income"
	// TransactionTypeExpense representing expense transactions (e.g. groceries).
	TransactionTypeExpense TransactionType = "expense"
	// TransactionTypeTransfer representing transfers between accounts.
	TransactionTypeTransfer TransactionType = "transfer"
)

// Transaction is a single financial event on an account.
// All amounts are stored in cents (int64) to ensure precision.
type Transaction struct {
	OccurredAt       time.Time       `json:"occurred_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty"`
	ID               string          `json:"id"`
	TenantID         string          `json:"tenant_id"`
	AccountID        string          `json:"account_id"`
	CategoryID       string          `json:"category_id"`
	UserID           string          `json:"user_id"`
	MasterPurchaseID string          `json:"master_purchase_id,omitempty"` // Empty for regular transactions; set in Phase 2
	Description      string          `json:"description"`
	Type             TransactionType `json:"type"`
	AmountCents      int64           `json:"amount_cents"` // Always in cents; positive value; type determines direction
}

// CreateTransactionInput is the value object used for creating a new transaction.
type CreateTransactionInput struct {
	OccurredAt       time.Time       `validate:"required"`
	AccountID        string          `validate:"required"`
	CategoryID       string          `validate:"required"`
	UserID           string          `validate:"required"`
	Description      string          `validate:"required,min=1,max=255"`
	MasterPurchaseID string          `validate:"omitempty"`
	Type             TransactionType `validate:"required,oneof=income expense transfer"`
	AmountCents      int64           `validate:"required,gt=0"`
}

// UpdateTransactionInput is the value object used for updating an existing transaction.
type UpdateTransactionInput struct {
	OccurredAt  *time.Time `validate:"omitempty"`
	CategoryID  *string    `validate:"omitempty"`
	Description *string    `validate:"omitempty,min=1,max=255"`
	AmountCents *int64     `validate:"omitempty,gt=0"`
}

// ListTransactionsParams defines the filters and pagination for listing transactions.
type ListTransactionsParams struct {
	StartDate  *time.Time      `json:"start_date,omitempty"`
	EndDate    *time.Time      `json:"end_date,omitempty"`
	AccountID  string          `json:"account_id,omitempty"`
	CategoryID string          `json:"category_id,omitempty"`
	Type       TransactionType `json:"type,omitempty"`
	Limit      int32           `json:"limit"`
	Offset     int32           `json:"offset"`
}

// TransactionRepository defines persistence operations for transactions.
// It ensures that all data is isolated by tenant.
type TransactionRepository interface {
	// Create persists a new transaction for the specified tenant.
	Create(ctx context.Context, tenantID string, input CreateTransactionInput) (*Transaction, error)

	// GetByID retrieves a specific transaction by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*Transaction, error)

	// List returns a list of transactions matching the specified filters.
	List(ctx context.Context, tenantID string, params ListTransactionsParams) ([]Transaction, error)

	// Update modifies an existing transaction's metadata.
	Update(ctx context.Context, tenantID, id string, input UpdateTransactionInput) (*Transaction, error)

	// Delete performs a soft delete on the specified transaction.
	Delete(ctx context.Context, tenantID, id string) error
}

// TransactionService defines the business-logic contract for transaction management.
type TransactionService interface {
	// Create persists a transaction and updates the account balance.
	Create(ctx context.Context, tenantID string, input CreateTransactionInput) (*Transaction, error)

	// GetByID retrieves a transaction by ID with tenant isolation.
	GetByID(ctx context.Context, tenantID, id string) (*Transaction, error)

	// List returns transactions matching filters for the household.
	List(ctx context.Context, tenantID string, params ListTransactionsParams) ([]Transaction, error)

	// Update modifies a transaction and adjusts balance if AmountCents changed.
	Update(ctx context.Context, tenantID, id string, input UpdateTransactionInput) (*Transaction, error)

	// Delete reverts the balance impact and soft-deletes the transaction.
	Delete(ctx context.Context, tenantID, id string) error
}
