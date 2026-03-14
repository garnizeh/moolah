package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/garnizeh/moolah/internal/domain"
)

// CategoryService provides business logic for managing categories, including creation, retrieval, updating, and deletion of categories. It also handles related audit logging.
type categoryService struct {
	categoryRepo domain.CategoryRepository
	auditRepo    domain.AuditRepository
}

// NewCategoryService creates a new instance of CategoryService.
func NewCategoryService(categoryRepo domain.CategoryRepository, auditRepo domain.AuditRepository) domain.CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		auditRepo:    auditRepo,
	}
}

// Create creates a new category for the specified tenant, and logs the creation action in the audit trail. It also enforces a maximum hierarchy depth of 1 (i.e., a category can have a parent, but not a grandparent).
func (s *categoryService) Create(ctx context.Context, tenantID string, input domain.CreateCategoryInput) (*domain.Category, error) {
	// 1. Hierarchy depth validation.
	if input.ParentID != "" {
		parent, err := s.categoryRepo.GetByID(ctx, tenantID, input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("category service: failed to fetch parent category: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("category service: parent category not found: %w", domain.ErrNotFound)
		}

		// A child cannot have its own parent (max depth 1).
		if parent.ParentID != "" {
			return nil, fmt.Errorf("category service: hierarchy depth exceeded (max 1): %w", domain.ErrInvalidInput)
		}
	}

	// 2. Persist category.
	category, err := s.categoryRepo.Create(ctx, tenantID, input)
	if err != nil {
		return nil, fmt.Errorf("category service: failed to create category: %w", err)
	}

	// 3. Write audit log.
	newValues, err := json.Marshal(map[string]any{
		"parent_id": category.ParentID,
		"name":      category.Name,
		"type":      category.Type,
		"icon":      category.Icon,
		"color":     category.Color,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal audit trail for category creation", "error", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    "system", // Placeholder: should be context userID in future
		EntityType: "category",
		EntityID:   category.ID,
		Action:     domain.AuditActionCreate,
		NewValues:  newValues,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create audit log for category creation", "error", err)
	}

	return category, nil
}

// GetByID retrieves a category by its ID and tenant ID. It returns domain.ErrNotFound if the category does not exist.
func (s *categoryService) GetByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category service: failed to find category: %w", err)
	}
	return category, nil
}

// ListByTenant returns all categories for a given tenant.
func (s *categoryService) ListByTenant(ctx context.Context, tenantID string) ([]domain.Category, error) {
	categories, err := s.categoryRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("category service: failed to list categories: %w", err)
	}
	return categories, nil
}

// ListChildren returns all child categories for a given tenant and parent category ID.
func (s *categoryService) ListChildren(ctx context.Context, tenantID, parentID string) ([]domain.Category, error) {
	categories, err := s.categoryRepo.ListChildren(ctx, tenantID, parentID)
	if err != nil {
		return nil, fmt.Errorf("category service: failed to list subcategories: %w", err)
	}
	return categories, nil
}

// Update modifies an existing category's details, and logs the update action in the audit trail, including the old and new values of the changed fields.
func (s *categoryService) Update(ctx context.Context, tenantID, id string, input domain.UpdateCategoryInput) (*domain.Category, error) {
	oldCategory, err := s.categoryRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("category service: failed to fetch existing category: %w", err)
	}

	category, err := s.categoryRepo.Update(ctx, tenantID, id, input)
	if err != nil {
		return nil, fmt.Errorf("category service: failed to update category: %w", err)
	}

	newValuesMap := make(map[string]any)
	oldValuesMap := make(map[string]any)

	if input.Name != nil && *input.Name != oldCategory.Name {
		newValuesMap["name"] = *input.Name
		oldValuesMap["name"] = oldCategory.Name
	}
	if input.Icon != nil && *input.Icon != oldCategory.Icon {
		newValuesMap["icon"] = *input.Icon
		oldValuesMap["icon"] = oldCategory.Icon
	}
	if input.Color != nil && *input.Color != oldCategory.Color {
		newValuesMap["color"] = *input.Color
		oldValuesMap["color"] = oldCategory.Color
	}

	if len(newValuesMap) > 0 {
		oldValues, err := json.Marshal(oldValuesMap)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal old values for category update audit", "error", err)
		}
		newValues, err := json.Marshal(newValuesMap)
		if err != nil {
			slog.ErrorContext(ctx, "failed to marshal new values for category update audit", "error", err)
		}

		_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
			TenantID:   tenantID,
			ActorID:    "system",
			EntityType: "category",
			EntityID:   id,
			Action:     domain.AuditActionUpdate,
			OldValues:  oldValues,
			NewValues:  newValues,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to create audit log for category update", "error", err)
		}
	}

	return category, nil
}

// Delete performs a soft delete of the category, and logs the deletion action in the audit trail. It first checks if the category exists before attempting deletion.
func (s *categoryService) Delete(ctx context.Context, tenantID, id string) error {
	// Fetch to ensure it exists.
	_, err := s.categoryRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("category service: failed to locate category for deletion: %w", err)
	}

	_, err = s.auditRepo.Create(ctx, domain.CreateAuditLogInput{
		TenantID:   tenantID,
		ActorID:    "system",
		EntityType: "category",
		EntityID:   id,
		Action:     domain.AuditActionSoftDelete,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create audit log for category deletion", "error", err)
	}

	if err := s.categoryRepo.Delete(ctx, tenantID, id); err != nil {
		return fmt.Errorf("category service: failed to delete category: %w", err)
	}

	return nil
}
