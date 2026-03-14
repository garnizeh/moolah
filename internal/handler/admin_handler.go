package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/go-playground/validator/v10"
)

// AdminHandler handles system administration HTTP requests. These endpoints are restricted to sysadmin users and allow management of tenants, users, and viewing audit logs.
type AdminHandler struct {
	service  domain.AdminService
	validate *validator.Validate
}

// NewAdminHandler creates a new AdminHandler with the given AdminService.
func NewAdminHandler(service domain.AdminService) *AdminHandler {
	return &AdminHandler{
		service:  service,
		validate: validator.New(),
	}
}

// UpdateTenantPlanRequest represents the request body for updating a tenant's subscription plan.
type UpdateTenantPlanRequest struct {
	Plan domain.TenantPlan `json:"plan" validate:"required,oneof=free pro business"`
}

// ListTenants GET /v1/admin/tenants
//
// @Summary		List all tenants
// @Description	Returns a list of all tenants in the system. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			with_deleted	query		bool	false	"Include soft-deleted tenants"
// @Success		200				{array}		domain.Tenant
// @Failure		401				{object}	map[string]string	"Unauthorized"
// @Failure		403				{object}	map[string]string	"Forbidden"
// @Failure		500				{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/tenants [get]
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
//
// @Summary		Get tenant details
// @Description	Returns details of any tenant by ID. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Tenant ULID"
// @Success		200	{object}	domain.Tenant
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string		"Tenant not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/tenants/{id} [get]
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
//
// @Summary		Update tenant plan
// @Description	Changes the subscription plan of a tenant. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string					true	"Tenant ULID"
// @Param			request	body		UpdateTenantPlanRequest	true	"New plan details"
// @Success		200		{object}	domain.Tenant
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string thin]string	"Internal server error"
// @Router			/v1/admin/tenants/{id}/plan [patch]
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
//
// @Summary		Suspend tenant
// @Description	Suspends a tenant's access to the system. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path	string	true	"Tenant ULID"
// @Success		204	"Tenant suspended"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string		"Tenant not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/tenants/{id}/suspend [post]
func (h *AdminHandler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.SuspendTenant(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to suspend tenant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RestoreTenant POST /v1/admin/tenants/{id}/restore
//
// @Summary		Restore tenant
// @Description	Restores a suspended or soft-deleted tenant. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path	string	true	"Tenant ULID"
// @Success		204	"Tenant restored"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string		"Tenant not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/tenants/{id}/restore [post]
func (h *AdminHandler) RestoreTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.RestoreTenant(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to restore tenant")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HardDeleteTenant DELETE /v1/admin/tenants/{id}
//
// @Summary		Hard delete tenant
// @Description	Permanently deletes a tenant and all its data. Requires sysadmin role and matching X-Confirm-Token header.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id				path	string	true	"Tenant ULID"
// @Param			X-Confirm-Token	header	string	true	"Confirmation token (must match ID)"
// @Success		204				"Tenant deleted permanently"
// @Failure		400				{object}	map[string]string	"Invalid confirmation token"
// @Failure		401				{object}	map[string]string	"Unauthorized"
// @Failure		403				{object}	map[string]string	"Forbidden"
// @Failure		500				{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/tenants/{id} [delete]
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
//
// @Summary		List all users
// @Description	Returns a list of all users in the system across all tenants. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		domain.User
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/users [get]
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListAllUsers(r.Context())
	if err != nil {
		handleError(w, r, err, "failed to list users")
		return
	}
	respondJSON(w, r, users, http.StatusOK)
}

// GetUser GET /v1/admin/users/{id}
//
// @Summary		Get user details
// @Description	Returns details of any user by ID. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"User ULID"
// @Success		200	{object}	domain.User
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string		"User not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/users/{id} [get]
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
//
// @Summary		Force delete user
// @Description	Permanently deletes a user from the system. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path	string	true	"User ULID"
// @Success		204	"User deleted permanently"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/users/{id} [delete]
func (h *AdminHandler) ForceDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.ForceDeleteUser(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to force-delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListAuditLogs GET /v1/admin/audit-logs
//
// @Summary		List audit logs
// @Description	Returns a paginated list of system audit logs. Requires sysadmin role.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			limit	query		int	false	"Limit (default 50)"
// @Param			offset	query		int	false	"Offset (default 0)"
// @Success		200		{object}	map[string]any
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/admin/audit-logs [get]
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
