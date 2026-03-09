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
		slog.ErrorContext(r.Context(), "failed to list accounts", "error", err, "tenant_id", tenantID)
		respondError(w, r, "internal server error", http.StatusInternalServerError)
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
		switch {
		case errors.Is(err, domain.ErrConflict):
			respondError(w, r, "account name conflict", http.StatusConflict)
		case errors.Is(err, domain.ErrForbidden):
			respondError(w, r, "forbidden", http.StatusForbidden)
		default:
			slog.ErrorContext(r.Context(), "failed to create account", "error", err, "tenant_id", tenantID)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
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
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "account not found", http.StatusNotFound)
		default:
			slog.ErrorContext(r.Context(), "failed to get account", "error", err, "tenant_id", tenantID, "account_id", id)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
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
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "account not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			respondError(w, r, "forbidden", http.StatusForbidden)
		case errors.Is(err, domain.ErrConflict):
			respondError(w, r, "account name conflict", http.StatusConflict)
		default:
			slog.ErrorContext(r.Context(), "failed to update account", "error", err, "tenant_id", tenantID, "account_id", id)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
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
		switch {
		case errors.Is(err, domain.ErrNotFound):
			respondError(w, r, "account not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrForbidden):
			respondError(w, r, "forbidden", http.StatusForbidden)
		default:
			slog.ErrorContext(r.Context(), "failed to delete account", "error", err, "tenant_id", tenantID, "account_id", id)
			respondError(w, r, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
