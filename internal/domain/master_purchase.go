package domain

import (
	"context"
	"time"
)

// MasterPurchaseStatus represents the lifecycle state of a master purchase.
type MasterPurchaseStatus string

const (
	// MasterPurchaseStatusOpen indicates installments are still pending.
	MasterPurchaseStatusOpen MasterPurchaseStatus = "open"
	// MasterPurchaseStatusClosed indicates all installments have been materialized.
	MasterPurchaseStatusClosed MasterPurchaseStatus = "closed"
)

// MasterPurchase is a credit-card purchase that generates installment transactions
// one at a time at each invoice-close cycle, rather than all at once.
// All monetary values are stored in cents (int64).
type MasterPurchase struct {
	FirstInstallmentDate time.Time            `json:"first_installment_date"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	DeletedAt            *time.Time           `json:"deleted_at,omitempty"`
	ID                   string               `json:"id"`
	TenantID             string               `json:"tenant_id"`
	AccountID            string               `json:"account_id"` // must be AccountTypeCreditCard
	CategoryID           string               `json:"category_id"`
	UserID               string               `json:"user_id"`
	Description          string               `json:"description"`
	Status               MasterPurchaseStatus `json:"status"`
	TotalAmountCents     int64                `json:"total_amount_cents"`
	InstallmentCount     int32                `json:"installment_count"`
	PaidInstallments     int32                `json:"paid_installments"` // incremented each close cycle
	ClosingDay           int32                `json:"closing_day"`       // day-of-month invoice closes (1–28)
}

// CreateMasterPurchaseInput is the value object for creating a new master purchase.
type CreateMasterPurchaseInput struct {
	FirstInstallmentDate time.Time `validate:"required"`
	AccountID            string    `validate:"required"`
	CategoryID           string    `validate:"required"`
	UserID               string    `validate:"required"`
	Description          string    `validate:"required,min=1,max=255"`
	TotalAmountCents     int64     `validate:"required,gt=0"`
	InstallmentCount     int32     `validate:"required,min=2,max=48"`
	ClosingDay           int32     `validate:"required,min=1,max=28"`
}

// UpdateMasterPurchaseInput allows patching description or category only.
type UpdateMasterPurchaseInput struct {
	CategoryID  *string `validate:"omitempty"`
	Description *string `validate:"omitempty,min=1,max=255"`
}

// ProjectedInstallment is a runtime-computed installment — never stored in the DB directly.
type ProjectedInstallment struct {
	DueDate           time.Time `json:"due_date"`
	AmountCents       int64     `json:"amount_cents"` // last installment absorbs remainder
	InstallmentNumber int32     `json:"installment_number"`
}

// MasterPurchaseRepository defines persistence operations for master purchases.
// All queries enforce tenant isolation via tenant_id.
type MasterPurchaseRepository interface {
	// Create persists a new master purchase for the specified tenant.
	Create(ctx context.Context, tenantID string, input CreateMasterPurchaseInput) (*MasterPurchase, error)
	// GetByID retrieves a specific master purchase by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*MasterPurchase, error)
	// ListByTenant returns all active master purchases for the given tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]MasterPurchase, error)
	// ListByAccount returns all master purchases associated with a specific account.
	ListByAccount(ctx context.Context, tenantID, accountID string) ([]MasterPurchase, error)
	// ListPendingClose returns open master purchases whose next installment falls on or before cutoffDate.
	ListPendingClose(ctx context.Context, tenantID string, cutoffDate time.Time) ([]MasterPurchase, error)
	// Update modifies an existing master purchase's description or category.
	Update(ctx context.Context, tenantID, id string, input UpdateMasterPurchaseInput) (*MasterPurchase, error)
	// IncrementPaidInstallments atomically advances PaidInstallments and, if all paid, sets status=closed.
	IncrementPaidInstallments(ctx context.Context, tenantID, id string) error
	// Delete performs a soft delete on the specified master purchase.
	Delete(ctx context.Context, tenantID, id string) error
}

// MasterPurchaseService defines the business-logic contract for master purchases.
type MasterPurchaseService interface {
	// Create persists a new master purchase after validating the account type.
	Create(ctx context.Context, tenantID string, input CreateMasterPurchaseInput) (*MasterPurchase, error)
	// GetByID retrieves a specific master purchase by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*MasterPurchase, error)
	// ListByTenant returns all active master purchases for the given tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]MasterPurchase, error)
	// ListByAccount returns all master purchases associated with a specific account.
	ListByAccount(ctx context.Context, tenantID, accountID string) ([]MasterPurchase, error)
	// ProjectInstallments computes the full installment schedule without persisting anything.
	ProjectInstallments(mp *MasterPurchase) []ProjectedInstallment
	// Update modifies an existing master purchase's metadata.
	Update(ctx context.Context, tenantID, id string, input UpdateMasterPurchaseInput) (*MasterPurchase, error)
	// Delete performs a soft delete on the specified master purchase.
	Delete(ctx context.Context, tenantID, id string) error
}

// CloseInvoiceResult reports the outcome of an invoice closing operation.
type CloseInvoiceResult struct {
	Errors         []error
	ProcessedCount int
}

// InvoiceCloser defines the contract for materializing installments at invoice-close time.
type InvoiceCloser interface {
	// CloseInvoice finds all pending master purchases for the account due on or before
	// closingDate, materializes the current installment as a transaction, and advances
	// paid_installments. Runs each master purchase in its own DB transaction.
	CloseInvoice(ctx context.Context, tenantID, accountID string, closingDate time.Time) (CloseInvoiceResult, error)
}
