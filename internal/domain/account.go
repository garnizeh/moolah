package domain

import (
	"context"
	"time"
)

// AccountType mirrors the database enum for financial account types.
type AccountType string

const (
	// AccountTypeChecking representing a standard checking account.
	AccountTypeChecking AccountType = "checking"
	// AccountTypeSavings representing a savings or high-yield account.
	AccountTypeSavings AccountType = "savings"
	// AccountTypeCreditCard representing a credit card account.
	AccountTypeCreditCard AccountType = "credit_card"
	// AccountTypeInvestment representing an investment or brokerage account.
	AccountTypeInvestment AccountType = "investment"
)

// Account represents a financial account owned by a user within a household (tenant).
// All balances are stored in cents (int64) to ensure precision.
type Account struct {
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt    *time.Time  `json:"deleted_at,omitempty"`
	ID           string      `json:"id"`
	TenantID     string      `json:"tenant_id"`
	UserID       string      `json:"user_id"`
	Name         string      `json:"name"`
	Type         AccountType `json:"type"`
	Currency     string      `json:"currency"`      // ISO 4217 code (e.g. "USD")
	BalanceCents int64       `json:"balance_cents"` // Always in cents; never float
}

// CreateAccountInput is the value object used for creating a new account.
type CreateAccountInput struct {
	UserID       string      `validate:"required"`
	Name         string      `validate:"required,min=1,max=100"`
	Type         AccountType `validate:"required,oneof=checking savings credit_card investment"`
	Currency     string      `validate:"required,len=3"`
	InitialCents int64       `validate:"required"`
}

// UpdateAccountInput is the value object used for updating an existing account.
type UpdateAccountInput struct {
	Name     *string `validate:"omitempty,min=1,max=100"`
	Currency *string `validate:"omitempty,len=3"`
}

// AccountRepository defines the persistence operations for financial accounts.
// It ensures that all data is isolated by tenant.
type AccountRepository interface {
	// Create persists a new account for the specified tenant.
	Create(ctx context.Context, tenantID string, input CreateAccountInput) (*Account, error)

	// GetByID retrieves a specific account by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*Account, error)

	// ListByTenant returns all active accounts for the given tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]Account, error)

	// ListByUser returns all accounts associated with a specific user within a tenant.
	ListByUser(ctx context.Context, tenantID, userID string) ([]Account, error)

	// Update modifies an existing account's metadata.
	Update(ctx context.Context, tenantID, id string, input UpdateAccountInput) (*Account, error)

	// UpdateBalance updates the balance of an account. This should be used strictly
	// through service logic to ensure consistency with transactions.
	UpdateBalance(ctx context.Context, tenantID, id string, newBalanceCents int64) error

	// Delete performs a soft delete on the specified account.
	Delete(ctx context.Context, tenantID, id string) error
}
