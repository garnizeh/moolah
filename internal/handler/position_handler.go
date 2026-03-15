package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// PositionHandler handles position-related HTTP requests.
type PositionHandler struct {
	service  domain.InvestmentService
	validate *validator.Validate
}

// NewPositionHandler creates a new PositionHandler.
func NewPositionHandler(service domain.InvestmentService) *PositionHandler {
	return &PositionHandler{
		service:  service,
		validate: validator.New(),
	}
}

// CreatePositionRequest defines the payload for creating a new position.
type CreatePositionRequest struct {
	PurchasedAt        time.Time         `json:"purchased_at"       validate:"required"`
	IncomeAmountCents  *int64            `json:"income_amount_cents"  validate:"omitempty,gte=0"`
	IncomeIntervalDays *int              `json:"income_interval_days" validate:"omitempty,gt=0"`
	IncomeRateBps      *int              `json:"income_rate_bps"      validate:"omitempty,gte=0"`
	NextIncomeAt       *time.Time        `json:"next_income_at"     validate:"omitempty"`
	MaturityAt         *time.Time        `json:"maturity_at"        validate:"omitempty"`
	Quantity           decimal.Decimal   `json:"quantity"           validate:"required"`
	Currency           string            `json:"currency"           validate:"required,len=3"`
	AccountID          string            `json:"account_id"         validate:"required"`
	IncomeType         domain.IncomeType `json:"income_type"        validate:"required,oneof=none dividend coupon rent interest salary"`
	AssetID            string            `json:"asset_id"           validate:"required"`
	AvgCostCents       int64             `json:"avg_cost_cents"     validate:"gte=0"`
	LastPriceCents     int64             `json:"last_price_cents"   validate:"gte=0"`
}

// UpdatePositionRequest defines the payload for updating an existing position.
type UpdatePositionRequest struct {
	Quantity           *decimal.Decimal   `json:"quantity"             validate:"omitempty"`
	AvgCostCents       *int64             `json:"avg_cost_cents"       validate:"omitempty,gte=0"`
	LastPriceCents     *int64             `json:"last_price_cents"     validate:"omitempty,gte=0"`
	IncomeType         *domain.IncomeType `json:"income_type"          validate:"omitempty,oneof=none dividend coupon rent interest salary"`
	IncomeIntervalDays *int               `json:"income_interval_days" validate:"omitempty,gt=0"`
	IncomeAmountCents  *int64             `json:"income_amount_cents"  validate:"omitempty,gte=0"`
	IncomeRateBps      *int               `json:"income_rate_bps"      validate:"omitempty,gte=0"`
	NextIncomeAt       *time.Time         `json:"next_income_at"       validate:"omitempty"`
	MaturityAt         *time.Time         `json:"maturity_at"          validate:"omitempty"`
}

// List handles GET /v1/positions
//
// @Summary		List positions
// @Description	Returns all positions belonging to the current tenant.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		domain.Position
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/positions [get]
func (h *PositionHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	positions, err := h.service.ListPositions(r.Context(), tenantID)
	if err != nil {
		handleError(w, r, err, "failed to list positions")
		return
	}

	respondJSON(w, r, positions, http.StatusOK)
}

// Create handles POST /v1/positions
//
// @Summary		Create position
// @Description	Creates a new investment position for the current tenant.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			Idempotency-Key	header	string					false	"Optional idempotency key (ULID format recommended)"
// @Param			request			body	CreatePositionRequest	true	"Position details"
// @Success		201				{object}	domain.Position
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/positions [post]
func (h *PositionHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.CreatePositionInput{
		AssetID:            req.AssetID,
		AccountID:          req.AccountID,
		Quantity:           req.Quantity,
		AvgCostCents:       req.AvgCostCents,
		LastPriceCents:     req.LastPriceCents,
		Currency:           req.Currency,
		PurchasedAt:        req.PurchasedAt,
		IncomeType:         req.IncomeType,
		IncomeIntervalDays: req.IncomeIntervalDays,
		IncomeAmountCents:  req.IncomeAmountCents,
		IncomeRateBps:      req.IncomeRateBps,
		NextIncomeAt:       req.NextIncomeAt,
		MaturityAt:         req.MaturityAt,
	}

	pos, err := h.service.CreatePosition(r.Context(), tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to create position")
		return
	}

	respondJSON(w, r, pos, http.StatusCreated)
}

// GetByID handles GET /v1/positions/{id}
//
// @Summary		Get position
// @Description	Returns a single position by its ID.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Position ID"
// @Success		200	{object}	domain.Position
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string	"Position not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/positions/{id} [get]
func (h *PositionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	pos, err := h.service.GetPosition(r.Context(), tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to get position")
		return
	}

	respondJSON(w, r, pos, http.StatusOK)
}

// Update handles PATCH /v1/positions/{id}
//
// @Summary		Update position
// @Description	Updates an existing position's details.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id				path	string					true	"Position ID"
// @Param			Idempotency-Key	header	string					false	"Optional idempotency key"
// @Param			request			body	UpdatePositionRequest	true	"Update details"
// @Success		200				{object}	domain.Position
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		404		{object}	map[string]string	"Position not found"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/positions/{id} [patch]
func (h *PositionHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	var req UpdatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	input := domain.UpdatePositionInput{
		Quantity:           req.Quantity,
		AvgCostCents:       req.AvgCostCents,
		LastPriceCents:     req.LastPriceCents,
		IncomeType:         req.IncomeType,
		IncomeIntervalDays: req.IncomeIntervalDays,
		IncomeAmountCents:  req.IncomeAmountCents,
		IncomeRateBps:      req.IncomeRateBps,
		NextIncomeAt:       req.NextIncomeAt,
		MaturityAt:         req.MaturityAt,
	}

	pos, err := h.service.UpdatePosition(r.Context(), tenantID, id, input)
	if err != nil {
		handleError(w, r, err, "failed to update position")
		return
	}

	respondJSON(w, r, pos, http.StatusOK)
}

// Delete handles DELETE /v1/positions/{id}
//
// @Summary		Delete position
// @Description	Soft-deletes an existing position (closes/resigns).
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Position ID"
// @Success		204	"No content"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string	"Position not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/positions/{id} [delete]
func (h *PositionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if err := h.service.DeletePosition(r.Context(), tenantID, id); err != nil {
		handleError(w, r, err, "failed to delete position")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListByAccount handles GET /v1/accounts/{id}/positions
//
// @Summary		List positions by account
// @Description	Returns all positions associated with a specific account.
// @Tags			accounts
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Account ID"
// @Success		200	{array}		domain.Position
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/accounts/{id}/positions [get]
func (h *PositionHandler) ListByAccount(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	accountID := r.PathValue("id")
	positions, err := h.service.ListPositionsByAccount(r.Context(), tenantID, accountID)
	if err != nil {
		handleError(w, r, err, "failed to list positions for account")
		return
	}

	respondJSON(w, r, positions, http.StatusOK)
}

// ListIncomeEvents handles GET /v1/income-events
//
// @Summary		List all income events
// @Description	Returns all income events (receivables) for the current tenant.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		domain.PositionIncomeEvent
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/income-events [get]
func (h *PositionHandler) ListIncomeEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	events, err := h.service.ListIncomeEvents(r.Context(), tenantID, "")
	if err != nil {
		handleError(w, r, err, "failed to list income events")
		return
	}

	respondJSON(w, r, events, http.StatusOK)
}

// ListPendingIncomeEvents handles GET /v1/income-events/pending
//
// @Summary		List pending income events
// @Description	Returns only pending income events for the current tenant.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		domain.PositionIncomeEvent
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/income-events/pending [get]
func (h *PositionHandler) ListPendingIncomeEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	events, err := h.service.ListIncomeEvents(r.Context(), tenantID, string(domain.ReceivableStatusPending))
	if err != nil {
		handleError(w, r, err, "failed to list pending income events")
		return
	}

	respondJSON(w, r, events, http.StatusOK)
}

// MarkIncomeReceived handles PATCH /v1/income-events/{id}/receive
//
// @Summary		Mark income as received
// @Description	Marks a pending income event as received.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id				path	string	true	"Income Event ID"
// @Param			Idempotency-Key	header	string	false	"Optional idempotency key"
// @Success		200				{object}	domain.PositionIncomeEvent
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		404		{object}	map[string]string	"Income event not found"
// @Failure		409		{object}	map[string]string	"Conflict (already received/cancelled)"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/income-events/{id}/receive [patch]
func (h *PositionHandler) MarkIncomeReceived(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	event, err := h.service.MarkIncomeReceived(r.Context(), tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to mark income as received")
		return
	}

	respondJSON(w, r, event, http.StatusOK)
}

// CancelIncome handles PATCH /v1/income-events/{id}/cancel
//
// @Summary		Cancel income event
// @Description	Cancels (writes off) an income event.
// @Tags			positions
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Income Event ID"
// @Success		200	{object}	domain.PositionIncomeEvent
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string	"Income event not found"
// @Failure		409	{object}	map[string]string	"Conflict (already received/cancelled)"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/income-events/{id}/cancel [patch]
func (h *PositionHandler) CancelIncome(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	event, err := h.service.CancelIncome(r.Context(), tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to cancel income event")
		return
	}

	respondJSON(w, r, event, http.StatusOK)
}

// TriggerSnapshot handles POST /v1/portfolio/snapshot
//
// @Summary		Trigger manual snapshot
// @Description	Manually triggers a portfolio-wide snapshot for the current tenant.
// @Tags			portfolio
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			Idempotency-Key	header	string	false	"Optional idempotency key"
// @Success		201				{object}	domain.PortfolioSnapshot
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		409		{object}	map[string]string	"Snapshot already exists for today"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/portfolio/snapshot [post]
func (h *PositionHandler) TriggerSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	snapshot, err := h.service.TakeSnapshot(r.Context(), tenantID)
	if err != nil {
		handleError(w, r, err, "failed to trigger snapshot")
		return
	}

	respondJSON(w, r, snapshot, http.StatusCreated)
}

// GetSummary handles GET /v1/investments/summary
//
// @Summary		Get portfolio summary
// @Description	Returns the net worth and allocation breakdown.
// @Tags			portfolio
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{object}	domain.PortfolioSummary
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/investments/summary [get]
func (h *PositionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	summary, err := h.service.GetPortfolioSummary(r.Context(), tenantID)
	if err != nil {
		handleError(w, r, err, "failed to get portfolio summary")
		return
	}

	respondJSON(w, r, summary, http.StatusOK)
}
