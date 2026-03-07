package repository

import (
	"context"
	"fmt"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/pkg/ulid"
	"github.com/jackc/pgx/v5/pgtype"
)

type categoryRepo struct {
	q sqlc.Querier
}

// NewCategoryRepository creates a new concrete implementation of domain.CategoryRepository.
func NewCategoryRepository(q sqlc.Querier) domain.CategoryRepository {
	return &categoryRepo{q: q}
}

// Create persists a new category for the specified tenant.
func (r *categoryRepo) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	parentID := pgtype.Text{}
	if input.ParentID != "" {
		parentID = pgtype.Text{String: input.ParentID, Valid: true}
	}

	arg := sqlc.CreateCategoryParams{
		ID:       ulid.New(),
		TenantID: tenantID,
		ParentID: parentID,
		Name:     input.Name,
		Icon:     pgtype.Text{String: input.Icon, Valid: input.Icon != ""},
		Color:    pgtype.Text{String: input.Color, Valid: input.Color != ""},
		Type:     sqlc.CategoryType(input.Type),
	}

	row, err := r.q.CreateCategory(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", TranslateError(err))
	}

	return mapCategory(row), nil
}

// GetByID retrieves a specific category by its ID and tenant ID.
func (r *categoryRepo) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, sqlc.GetCategoryByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", TranslateError(err))
	}

	return mapCategory(row), nil
}

// ListByTenant returns all active categories for the given tenant.
func (r *categoryRepo) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	rows, err := r.q.ListCategoriesByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", TranslateError(err))
	}

	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, *mapCategory(row))
	}

	return categories, nil
}

// ListChildren returns all subcategories for a given parent within a tenant.
func (r *categoryRepo) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	rows, err := r.q.ListChildCategories(ctx, sqlc.ListChildCategoriesParams{
		TenantID: tenantID,
		ParentID: pgtype.Text{String: parentID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list child categories: %w", TranslateError(err))
	}

	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, *mapCategory(row))
	}

	return categories, nil
}

// Update modifies an existing category's metadata.
func (r *categoryRepo) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	// First get current category to handle partial updates
	current, err := r.q.GetCategoryByID(ctx, sqlc.GetCategoryByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get category for update: %w", TranslateError(err))
	}

	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}

	icon := current.Icon.String
	if input.Icon != nil {
		icon = *input.Icon
	}

	color := current.Color.String
	if input.Color != nil {
		color = *input.Color
	}

	row, err := r.q.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		TenantID: tenantID,
		ID:       id,
		ParentID: current.ParentID,
		Name:     name,
		Icon:     pgtype.Text{String: icon, Valid: icon != ""},
		Color:    pgtype.Text{String: color, Valid: color != ""},
		Type:     current.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", TranslateError(err))
	}

	return mapCategory(row), nil
}

// Delete performs a soft delete on the specified category.
func (r *categoryRepo) Delete(ctx context.Context, tenantID, id string) error {
	err := r.q.SoftDeleteCategory(ctx, sqlc.SoftDeleteCategoryParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", TranslateError(err))
	}

	return nil
}

func mapCategory(row sqlc.Category) *domain.Category {
	parentID := ""
	if row.ParentID.Valid {
		parentID = row.ParentID.String
	}

	return &domain.Category{
		ID:        row.ID,
		TenantID:  row.TenantID,
		ParentID:  parentID,
		Name:      row.Name,
		Icon:      row.Icon.String,
		Color:     row.Color.String,
		Type:      domain.CategoryType(row.Type),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: &row.DeletedAt.Time,
	}
}
