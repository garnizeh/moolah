package handler

import (
	"encoding/json"
	"net/http"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
)

// AccountHandler handles account-related HTTP requests.
type AccountHandler struct {
	service  domain.AccountService
	validate *validator.Validate
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(service domain.AccountService) *AccountHandler {
	return &AccountHandler{
		service:  service,
		validate: validator.New(),
	}
}

// CreateAccountRequest defines the payload for creating a new account.
type CreateAccountRequest struct {
	Name         string             `json:"name"          validate:"required,min=1,max=100"`
	Type         domain.AccountType `json:"type"          validate:"required,oneof=checking savings credit_card investment"`
	Currency     string             `json:"currency"      validate:"required,len=3"`
	InitialCents int64              `json:"initial_cents" validate:"required"`
}

// UpdateAccountRequest defines the payload for updating an account.
type UpdateAccountRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=1,max=100"`
	Currency *string `json:"currency" validate:"omitempty,len=3"`
}

// List handles GET /v1/accounts
//
// @Summary		List accounts
// @Description	Returns all accounts belonging to the current tenant.
// @Tags			accounts
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		domain.Account
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/accounts [get]
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := h.service.ListByTenant(r.Context(), tenantID)
	if err != nil {
		handleError(w, r, err, "failed to list accounts")
		return
	}

	respondJSON(w, r, accounts, http.StatusOK)
}

// Create handles POST /v1/accounts
//
// @Summary		Create account
// @Description	Creates a new financial account for the current tenant.
// @Tags			accounts
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			Idempotency-Key	header	string					false	"Optional idempotency key (ULID format recommended)"
// @Param			request			body	CreateAccountRequest	true	"Account details"
// @Success		201				{object}	domain.Account
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	userID, userOK := middleware.UserIDFromCtx(r.Context())
	if !ok || !userOK {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.CreateAccountInput{
		UserID:       userID,
		Name:         req.Name,
		Type:         req.Type,
		Currency:     req.Currency,
		InitialCents: req.InitialCents,
	}

	account, err := h.service.Create(r.Context(), tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to create account")
		return
	}

	respondJSON(w, r, account, http.StatusCreated)
}

// GetByID handles GET /v1/accounts/{id}
//
// @Summary		Get account
// @Description	Returns details of a specific account by ID.
// @Tags			accounts
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Account ULID"
// @Success		200	{object}	domain.Account
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string		"Account not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/accounts/{id} [get]
func (h *AccountHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing account id", http.StatusBadRequest)
		return
	}

	account, err := h.service.GetByID(r.Context(), tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to get account")
		return
	}

	respondJSON(w, r, account, http.StatusOK)
}

// Update handles PATCH /v1/accounts/{id}
//
// @Summary		Update account
// @Description	Updates details of an existing account.
// @Tags			accounts
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id				path	string					true	"Account ULID"
// @Param			Idempotency-Key	header	string					false	"Optional idempotency key (ULID format recommended)"
// @Param			request			body	UpdateAccountRequest	true	"Update fields"
// @Success		200				{object}	domain.Account
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		404		{object}	map[string]string		"Account not found"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/accounts/{id} [patch]
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing account id", http.StatusBadRequest)
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.UpdateAccountInput{
		Name:     req.Name,
		Currency: req.Currency,
	}

	account, err := h.service.Update(r.Context(), tenantID, id, input)
	if err != nil {
		handleError(w, r, err, "failed to update account")
		return
	}

	respondJSON(w, r, account, http.StatusOK)
}

// Delete handles DELETE /v1/accounts/{id}
//
// @Summary		Delete account
// @Description	Soft-deletes an account and its transactions.
// @Tags			accounts
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path	string	true	"Account ULID"
// @Success		204	"Account deleted"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string		"Account not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/accounts/{id} [delete]
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing account id", http.StatusBadRequest)
		return
	}

	err := h.service.Delete(r.Context(), tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to delete account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
