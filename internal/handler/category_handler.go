package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
)

type CategoryHandler struct {
	service  domain.CategoryService
	validate *validator.Validate
}

func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service:  service,
		validate: validator.New(),
	}
}

type CreateCategoryRequest struct {
	ParentID *string             `json:"parent_id" validate:"omitempty"`
	Icon     *string             `json:"icon"      validate:"omitempty"`
	Color    *string             `json:"color"     validate:"omitempty"`
	Name     string              `json:"name"      validate:"required,min=1,max=100"`
	Type     domain.CategoryType `json:"type"      validate:"required"`
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name"  validate:"omitempty,min=1,max=100"`
	Icon  *string `json:"icon"  validate:"omitempty"`
	Color *string `json:"color" validate:"omitempty"`
}

// List handles GET /v1/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	catType := r.URL.Query().Get("type")
	parentID := r.URL.Query().Get("parent_id")

	var categories []domain.Category
	var err error

	if parentID != "" {
		categories, err = h.service.ListChildren(r.Context(), tenantID, parentID)
	} else {
		categories, err = h.service.ListByTenant(r.Context(), tenantID)
	}

	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list categories", "error", err, "tenant_id", tenantID)
		respondError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	// Filter by type if provided
	if catType != "" {
		filtered := make([]domain.Category, 0)
		for _, c := range categories {
			if string(c.Type) == catType {
				filtered = append(filtered, c)
			}
		}
		categories = filtered
	}

	respondJSON(w, r, categories, http.StatusOK)
}

// Create handles POST /v1/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.CreateCategoryInput{
		Name: req.Name,
		Type: req.Type,
	}

	if req.ParentID != nil {
		input.ParentID = *req.ParentID
	}
	if req.Icon != nil {
		input.Icon = *req.Icon
	}
	if req.Color != nil {
		input.Color = *req.Color
	}

	category, err := h.service.Create(r.Context(), tenantID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConflict):
			respondError(w, r, "category already exists", http.StatusConflict)
		case errors.Is(err, domain.ErrInvalidInput):
			respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		default:
			slog.ErrorContext(r.Context(), "failed to create category", "error", err, "tenant_id", tenantID)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, r, category, http.StatusCreated)
}

// GetByID handles GET /v1/categories/{id}
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing category id", http.StatusBadRequest)
		return
	}

	category, err := h.service.GetByID(r.Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(w, r, "category not found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(r.Context(), "failed to get category", "error", err, "tenant_id", tenantID, "category_id", id)
		respondError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, r, category, http.StatusOK)
}

// Update handles PATCH /v1/categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing category id", http.StatusBadRequest)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.UpdateCategoryInput{
		Name:  req.Name,
		Icon:  req.Icon,
		Color: req.Color,
	}

	category, err := h.service.Update(r.Context(), tenantID, id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "category not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrConflict):
			respondError(w, r, "category name conflict", http.StatusConflict)
		default:
			slog.ErrorContext(r.Context(), "failed to update category", "error", err, "tenant_id", tenantID, "category_id", id)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, r, category, http.StatusOK)
}

// Delete handles DELETE /v1/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing category id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), tenantID, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(w, r, "category not found", http.StatusNotFound)
			return
		}
		slog.ErrorContext(r.Context(), "failed to delete category", "error", err, "tenant_id", tenantID, "category_id", id)
		respondError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
