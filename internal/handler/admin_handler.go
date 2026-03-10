package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/go-playground/validator/v10"
)

type AdminHandler struct {
	service  domain.AdminService
	validate *validator.Validate
}

func NewAdminHandler(service domain.AdminService) *AdminHandler {
	return &AdminHandler{
		service:  service,
		validate: validator.New(),
	}
}

type UpdateTenantPlanRequest struct {
	Plan domain.TenantPlan `json:"plan" validate:"required,oneof=free pro business"`
}

// ListTenants GET /v1/admin/tenants
func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	withDeleted := r.URL.Query().Get("with_deleted") == "true"
	tenants, err := h.service.ListAllTenants(r.Context(), withDeleted)
	if err != nil {
		handleError(w, r, err, "failed to list tenants")
		return
	}
	respondJSON(w, r, tenants, http.StatusOK)
}

// GetTenant GET /v1/admin/tenants/{id}
func (h *AdminHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenant, err := h.service.GetTenantByID(r.Context(), id)
	if err != nil {
		handleError(w, r, err, "failed to get tenant")
		return
	}
	respondJSON(w, r, tenant, http.StatusOK)
}

// UpdateTenantPlan PATCH /v1/admin/tenants/{id}/plan
func (h *AdminHandler) UpdateTenantPlan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req UpdateTenantPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tenant, err := h.service.UpdateTenantPlan(r.Context(), id, req.Plan)
	if err != nil {
		handleError(w, r, err, "failed to update tenant plan")
		return
	}
	respondJSON(w, r, tenant, http.StatusOK)
}

// SuspendTenant POST /v1/admin/tenants/{id}/suspend
func (h *AdminHandler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.SuspendTenant(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to suspend tenant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RestoreTenant POST /v1/admin/tenants/{id}/restore
func (h *AdminHandler) RestoreTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.RestoreTenant(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to restore tenant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HardDeleteTenant DELETE /v1/admin/tenants/{id}
func (h *AdminHandler) HardDeleteTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	token := r.Header.Get("X-Confirm-Token")

	// Pre-validation to avoid calling service if token is missing/wrong
	if token != id {
		respondError(w, r, "missing or invalid confirmation token", http.StatusBadRequest)
		return
	}

	if err := h.service.HardDeleteTenant(r.Context(), id, token); err != nil {
		handleError(w, r, err, "failed to hard-delete tenant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListUsers GET /v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListAllUsers(r.Context())
	if err != nil {
		handleError(w, r, err, "failed to list users")
		return
	}
	respondJSON(w, r, users, http.StatusOK)
}

// GetUser GET /v1/admin/users/{id}
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		handleError(w, r, err, "failed to get user")
		return
	}
	respondJSON(w, r, user, http.StatusOK)
}

// ForceDeleteUser DELETE /v1/admin/users/{id}
func (h *AdminHandler) ForceDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.ForceDeleteUser(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to force-delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListAuditLogs GET /v1/admin/audit-logs
func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	params := domain.ListAuditLogsParams{
		Limit:  50,
		Offset: 0,
	}

	if l := q.Get("limit"); l != "" {
		if val, err := strconv.ParseInt(l, 10, 32); err == nil && val > 0 && val <= 1000 {
			params.Limit = int32(val)
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.ParseInt(o, 10, 32); err == nil && val >= 0 && val <= 1000000 {
			params.Offset = int32(val)
		}
	}

	logs, err := h.service.ListAuditLogs(r.Context(), params)
	if err != nil {
		handleError(w, r, err, "failed to list audit logs")
		return
	}

	respondJSON(w, r, map[string]any{
		"data":   logs,
		"limit":  params.Limit,
		"offset": params.Offset,
	}, http.StatusOK)
}
