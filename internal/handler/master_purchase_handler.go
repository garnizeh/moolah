package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
)

// MasterPurchaseHandler handles HTTP requests for master purchase management.
type MasterPurchaseHandler struct {
	service  domain.MasterPurchaseService
	validate *validator.Validate
}

// NewMasterPurchaseHandler creates a new MasterPurchaseHandler instance.
func NewMasterPurchaseHandler(service domain.MasterPurchaseService) *MasterPurchaseHandler {
	return &MasterPurchaseHandler{
		service:  service,
		validate: validator.New(),
	}
}

// CreateMasterPurchaseRequest defines the incoming payload for a master purchase.
type CreateMasterPurchaseRequest struct {
	FirstInstallmentDate time.Time `json:"first_installment_date"  validate:"required"`
	AccountID            string    `json:"account_id"             validate:"required"`
	CategoryID           string    `json:"category_id"            validate:"required"`
	Description          string    `json:"description"            validate:"required,min=1,max=255"`
	TotalAmountCents     int64     `json:"total_amount_cents"      validate:"required,gt=0"`
	InstallmentCount     int32     `json:"installment_count"      validate:"required,min=2,max=48"`
	ClosingDay           int32     `json:"closing_day"            validate:"required,min=1,max=28"`
}

// UpdateMasterPurchaseRequest defines the patchable fields.
type UpdateMasterPurchaseRequest struct {
	CategoryID  *string `json:"category_id"  validate:"omitempty"`
	Description *string `json:"description"  validate:"omitempty,min=1,max=255"`
}

// Create handles POST /v1/master-purchases
func (h *MasterPurchaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	userID, userOK := middleware.UserIDFromCtx(ctx)
	if !ok || !userOK {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateMasterPurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	input := domain.CreateMasterPurchaseInput{
		AccountID:            req.AccountID,
		CategoryID:           req.CategoryID,
		UserID:               userID,
		Description:          req.Description,
		TotalAmountCents:     req.TotalAmountCents,
		InstallmentCount:     req.InstallmentCount,
		ClosingDay:           req.ClosingDay,
		FirstInstallmentDate: req.FirstInstallmentDate,
	}

	mp, err := h.service.Create(ctx, tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to create master purchase")
		return
	}

	respondJSON(w, r, mp, http.StatusCreated)
}

// GetByID handles GET /v1/master-purchases/{id}
func (h *MasterPurchaseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing id", http.StatusBadRequest)
		return
	}

	mp, err := h.service.GetByID(ctx, tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to get master purchase")
		return
	}

	respondJSON(w, r, mp, http.StatusOK)
}

// ListByTenant handles GET /v1/master-purchases
func (h *MasterPurchaseHandler) ListByTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Optional account filter
	accountID := r.URL.Query().Get("account_id")
	var mps []domain.MasterPurchase
	var err error

	if accountID != "" {
		mps, err = h.service.ListByAccount(ctx, tenantID, accountID)
	} else {
		mps, err = h.service.ListByTenant(ctx, tenantID)
	}

	if err != nil {
		handleError(w, r, err, "failed to list master purchases")
		return
	}

	respondJSON(w, r, mps, http.StatusOK)
}

// Project handles GET /v1/master-purchases/{id}/projection
func (h *MasterPurchaseHandler) Project(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	mp, err := h.service.GetByID(ctx, tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to get master purchase for projection")
		return
	}

	projection := h.service.ProjectInstallments(mp)
	respondJSON(w, r, projection, http.StatusOK)
}

// Update handles PATCH /v1/master-purchases/{id}
func (h *MasterPurchaseHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	var req UpdateMasterPurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, r, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	input := domain.UpdateMasterPurchaseInput{
		CategoryID:  req.CategoryID,
		Description: req.Description,
	}

	mp, err := h.service.Update(ctx, tenantID, id, input)
	if err != nil {
		handleError(w, r, err, "failed to update master purchase")
		return
	}

	respondJSON(w, r, mp, http.StatusOK)
}

// Delete handles DELETE /v1/master-purchases/{id}
func (h *MasterPurchaseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := middleware.TenantIDFromCtx(ctx)
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if err := h.service.Delete(ctx, tenantID, id); err != nil {
		handleError(w, r, err, "failed to delete master purchase")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
