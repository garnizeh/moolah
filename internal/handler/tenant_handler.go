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
//
// @Summary		Get current tenant
// @Description	Returns the details of the tenant (household) associated with the authenticated user.
// @Tags			tenants
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	domain.Tenant
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string	"Tenant not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/tenants/me [get]
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
//
// @Summary		Update current tenant
// @Description	Updates the details of the current tenant (e.g., household name).
// @Tags			tenants
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			Idempotency-Key	header	string				false	"Optional idempotency key (ULID format recommended)"
// @Param			request			body	UpdateTenantRequest	true	"Update fields"
// @Success		200				{object}	domain.Tenant
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/tenants/me [patch]
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
//
// @Summary		Invite user
// @Description	Invites a new user (family member) to the current tenant.
// @Tags			tenants
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			Idempotency-Key	header	string				false	"Optional idempotency key (ULID format recommended)"
// @Param			request			body	InviteUserRequest	true	"Invite details"
// @Success		201				{object}	domain.User
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		403		{object}	map[string]string	"Forbidden (admin only)"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/tenants/me/invite [post]
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
