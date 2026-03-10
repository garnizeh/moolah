package handler

import (
	"encoding/json"
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
//
// @Summary		List categories
// @Description	Returns all categories for the current tenant. Can be filtered by type or parent ID.
// @Tags			categories
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			type		query		string	false	"Category type (income/expense)"
// @Param			parent_id	query		string	false	"Parent category ULID"
// @Success		200			{array}		domain.Category
// @Failure		401			{object}	map[string]string	"Unauthorized"
// @Failure		500			{object}	map[string]string	"Internal server error"
// @Router			/v1/categories [get]
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
		handleError(w, r, err, "failed to list categories")
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
//
// @Summary		Create category
// @Description	Creates a new category for the current tenant.
// @Tags			categories
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		CreateCategoryRequest	true	"Category details"
// @Success		201		{object}	domain.Category
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/categories [post]
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
		handleError(w, r, err, "failed to create category")
		return
	}

	respondJSON(w, r, category, http.StatusCreated)
}

// GetByID handles GET /v1/categories/{id}
//
// @Summary		Get category
// @Description	Returns details of a specific category by ID.
// @Tags			categories
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Category ULID"
// @Success		200	{object}	domain.Category
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string		"Category not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/categories/{id} [get]
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
		handleError(w, r, err, "failed to get category")
		return
	}

	respondJSON(w, r, category, http.StatusOK)
}

// Update handles PATCH /v1/categories/{id}
//
// @Summary		Update category
// @Description	Updates details of an existing category.
// @Tags			categories
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string					true	"Category ULID"
// @Param			request	body		UpdateCategoryRequest	true	"Update fields"
// @Success		200		{object}	domain.Category
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		404		{object}	map[string]string		"Category not found"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/categories/{id} [patch]
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
		handleError(w, r, err, "failed to update category")
		return
	}

	respondJSON(w, r, category, http.StatusOK)
}

// Delete handles DELETE /v1/categories/{id}
//
// @Summary		Delete category
// @Description	Soft-deletes a category. This may be blocked if there are transactions associated with it.
// @Tags			categories
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path	string	true	"Category ULID"
// @Success		204	"Category deleted"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string		"Category not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/categories/{id} [delete]
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
		handleError(w, r, err, "failed to delete category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
