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
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "tenant not found", http.StatusNotFound)
		default:
			slog.ErrorContext(r.Context(), "failed to get tenant", "error", err, "tenant_id", tenantID)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
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
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "tenant not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			respondError(w, r, "forbidden", http.StatusForbidden)
		default:
			slog.ErrorContext(r.Context(), "failed to update tenant", "error", err, "tenant_id", tenantID)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
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
		switch {
		case errors.Is(err, domain.ErrConflict):
			respondError(w, r, "user already exists", http.StatusConflict)
		case errors.Is(err, domain.ErrForbidden):
			respondError(w, r, "forbidden", http.StatusForbidden)
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "tenant not found", http.StatusNotFound)
		default:
			slog.ErrorContext(r.Context(), "failed to invite user", "error", err, "tenant_id", tenantID, "email", req.Email)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, r, user, http.StatusCreated)
}
