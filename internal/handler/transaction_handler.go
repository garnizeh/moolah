package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
)

// TransactionHandler handles HTTP requests for transaction management.
type TransactionHandler struct {
	service  domain.TransactionService
	validate *validator.Validate
}

// NewTransactionHandler creates a new TransactionHandler instance.
func NewTransactionHandler(service domain.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		service:  service,
		validate: validator.New(),
	}
}

// CreateTransactionRequest defines the incoming payload for creating a transaction.
type CreateTransactionRequest struct {
	OccurredAt  time.Time              `json:"occurred_at"  validate:"required"`
	AccountID   string                 `json:"account_id"   validate:"required"`
	CategoryID  string                 `json:"category_id"  validate:"required"`
	Description string                 `json:"description"  validate:"required,min=1,max=255"`
	Type        domain.TransactionType `json:"type"         validate:"required,oneof=income expense transfer"`
	AmountCents int64                  `json:"amount_cents" validate:"required,gt=0"`
}

// UpdateTransactionRequest defines the incoming payload for updating a transaction.
type UpdateTransactionRequest struct {
	OccurredAt  *time.Time `json:"occurred_at"  validate:"omitempty"`
	CategoryID  *string    `json:"category_id"  validate:"omitempty"`
	Description *string    `json:"description"  validate:"omitempty,min=1,max=255"`
	AmountCents *int64     `json:"amount_cents" validate:"omitempty,gt=0"`
}

// Create handles POST /v1/transactions
//
// @Summary		Create transaction
// @Description	Creates a new financial transaction (income, expense, or transfer).
// @Tags			transactions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		CreateTransactionRequest	true	"Transaction details"
// @Success		201		{object}	domain.Transaction
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	userID, userOK := middleware.UserIDFromCtx(ctx)
	if !ok || !userOK {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	input := domain.CreateTransactionInput{
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		UserID:      userID,
		Type:        req.Type,
		AmountCents: req.AmountCents,
		Description: req.Description,
		OccurredAt:  req.OccurredAt,
	}

	tx, err := h.service.Create(ctx, tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to create transaction")
		return
	}

	respondJSON(w, r, tx, http.StatusCreated)
}

// GetByID handles GET /v1/transactions/{id}
//
// @Summary		Get transaction
// @Description	Returns details of a specific transaction by ID.
// @Tags			transactions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Transaction ULID"
// @Success		200	{object}	domain.Transaction
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string		"Transaction not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/transactions/{id} [get]
func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing transaction id", http.StatusBadRequest)
		return
	}

	tx, err := h.service.GetByID(ctx, tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to fetch transaction")
		return
	}

	respondJSON(w, r, tx, http.StatusOK)
}

// Update handles PATCH /v1/transactions/{id}
//
// @Summary		Update transaction
// @Description	Updates details of an existing transaction.
// @Tags			transactions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id		path		string						true	"Transaction ULID"
// @Param			request	body		UpdateTransactionRequest	true	"Update fields"
// @Success		200		{object}	domain.Transaction
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		404		{object}	map[string]string		"Transaction not found"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/transactions/{id} [patch]
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing transaction id", http.StatusBadRequest)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.UpdateTransactionInput{
		OccurredAt:  req.OccurredAt,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		AmountCents: req.AmountCents,
	}

	tx, err := h.service.Update(ctx, tenantID, id, input)
	if err != nil {
		handleError(w, r, err, "failed to update transaction")
		return
	}

	respondJSON(w, r, tx, http.StatusOK)
}

// Delete handles DELETE /v1/transactions/{id}
//
// @Summary		Delete transaction
// @Description	Soft-deletes a transaction.
// @Tags			transactions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path	string	true	"Transaction ULID"
// @Success		204	"Transaction deleted"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string		"Transaction not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/transactions/{id} [delete]
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing transaction id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, tenantID, id); err != nil {
		handleError(w, r, err, "failed to delete transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /v1/transactions
//
// @Summary		List transactions
// @Description	Returns a list of transactions for the current tenant. Can be filtered by account, category, type, and date range.
// @Tags			transactions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			account_id	query		string	false	"Account ULID"
// @Param			category_id	query		string	false	"Category ULID"
// @Param			type		query		string	false	"Transaction type (income/expense/transfer)"
// @Param			start_date	query		string	false	"Start date (RFC3339)"
// @Param			end_date	query		string	false	"End date (RFC3339)"
// @Param			limit		query		int		false	"Limit (default 50)"
// @Param			offset		query		int		false	"Offset (default 0)"
// @Success		200			{array}		domain.Transaction
// @Failure		401			{object}	map[string]string	"Unauthorized"
// @Failure		500			{object}	map[string]string	"Internal server error"
// @Router			/v1/transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	params := domain.ListTransactionsParams{
		AccountID:  q.Get("account_id"),
		CategoryID: q.Get("category_id"),
		Type:       domain.TransactionType(q.Get("type")),
		Limit:      50,
		Offset:     0,
	}

	if s := q.Get("start_date"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			params.StartDate = &t
		}
	}
	if s := q.Get("end_date"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			params.EndDate = &t
		}
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

	txs, err := h.service.List(ctx, tenantID, params)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list transactions", "error", err, "tenant_id", tenantID)
		respondError(w, r, "internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, r, map[string]any{
		"data":   txs,
		"limit":  params.Limit,
		"offset": params.Offset,
	}, http.StatusOK)
}
