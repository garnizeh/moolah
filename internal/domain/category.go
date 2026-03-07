package domain

import (
	"context"
	"time"
)

// CategoryType mirrors the database enum for transaction categories.
type CategoryType string

const (
	// CategoryTypeIncome representing income transactions (e.g. salary).
	CategoryTypeIncome CategoryType = "income"
	// CategoryTypeExpense representing expense transactions (e.g. groceries).
	CategoryTypeExpense CategoryType = "expense"
	// CategoryTypeTransfer representing transfers between accounts.
	CategoryTypeTransfer CategoryType = "transfer"
)

// Category is a tenant-scoped label for classifying transactions.
// It supports one level of parent-child hierarchy.
type Category struct {
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at,omitempty"`
	ID        string       `json:"id"`
	TenantID  string       `json:"tenant_id"`
	ParentID  string       `json:"parent_id"` // Empty string means root category
	Name      string       `json:"name"`
	Icon      string       `json:"icon"`  // Optional emoji or icon identifier
	Color     string       `json:"color"` // Optional hex color, e.g. "#FF5733"
	Type      CategoryType `json:"type"`
}

// IsRoot returns true when the category has no parent.
func (c *Category) IsRoot() bool {
	return c.ParentID == ""
}

// CreateCategoryInput is the value object used for creating a new category.
type CreateCategoryInput struct {
	ParentID string       `validate:"omitempty"`
	Name     string       `validate:"required,min=1,max=100"`
	Icon     string       `validate:"omitempty,max=10"`
	Color    string       `validate:"omitempty,hexcolor"`
	Type     CategoryType `validate:"required,oneof=income expense transfer"`
}

// UpdateCategoryInput is the value object used for updating an existing category.
type UpdateCategoryInput struct {
	Name  *string `validate:"omitempty,min=1,max=100"`
	Icon  *string `validate:"omitempty,max=10"`
	Color *string `validate:"omitempty,hexcolor"`
}

// CategoryRepository defines persistence operations for categories.
// It ensures that all data is isolated by tenant.
type CategoryRepository interface {
	// Create persists a new category for the specified tenant.
	Create(ctx context.Context, tenantID string, input CreateCategoryInput) (*Category, error)

	// GetByID retrieves a specific category by its ID and tenant ID.
	GetByID(ctx context.Context, tenantID, id string) (*Category, error)

	// ListByTenant returns all active categories for the given tenant.
	ListByTenant(ctx context.Context, tenantID string) ([]Category, error)

	// ListChildren returns all subcategories for a given parent within a tenant.
	ListChildren(ctx context.Context, tenantID, parentID string) ([]Category, error)

	// Update modifies an existing category's metadata.
	Update(ctx context.Context, tenantID, id string, input UpdateCategoryInput) (*Category, error)

	// Delete performs a soft delete on the specified category.
	Delete(ctx context.Context, tenantID, id string) error
}
