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
