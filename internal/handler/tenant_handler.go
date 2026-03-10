package handler

import (
	"encoding/json"
	"net/http"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
)

// TenantHandler handles tenant-related HTTP requests.
type TenantHandler struct {
	service  domain.TenantService
	validate *validator.Validate
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(service domain.TenantService) *TenantHandler {
	return &TenantHandler{
		service:  service,
		validate: validator.New(),
	}
}

// UpdateTenantRequest defines the payload for updating a tenant.
type UpdateTenantRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

// InviteUserRequest defines the payload for inviting a user.
type InviteUserRequest struct {
	Email string      `json:"email" validate:"required,email"`
	Role  domain.Role `json:"role"  validate:"required,oneof=owner member"`
}

// GetMe handles GET /v1/tenants/me
func (h *TenantHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenant, err := h.service.GetByID(r.Context(), tenantID)
	if err != nil {
		handleError(w, r, err, "failed to get tenant")
		return
	}

	respondJSON(w, r, tenant, http.StatusOK)
}

// UpdateMe handles PATCH /v1/tenants/me
func (h *TenantHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.UpdateTenantInput{
		Name: req.Name,
	}

	tenant, err := h.service.Update(r.Context(), tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to update tenant")
		return
	}

	respondJSON(w, r, tenant, http.StatusOK)
}

// InviteUser handles POST /v1/tenants/me/invite
func (h *TenantHandler) InviteUser(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.CreateUserInput{
		Email: req.Email,
		Role:  req.Role,
	}

	user, err := h.service.InviteUser(r.Context(), tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to invite user")
		return
	}

	respondJSON(w, r, user, http.StatusCreated)
}
