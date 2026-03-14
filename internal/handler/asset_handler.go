package handler

import (
	"encoding/json"
	"net/http"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/go-playground/validator/v10"
)

// AssetHandler handles asset-related HTTP requests.
type AssetHandler struct {
	service  domain.InvestmentService
	validate *validator.Validate
}

// NewAssetHandler creates a new AssetHandler.
func NewAssetHandler(service domain.InvestmentService) *AssetHandler {
	return &AssetHandler{
		service:  service,
		validate: validator.New(),
	}
}

// ListAssets returns the global asset catalogue.
// @Summary		List assets
// @Description	Returns the global asset catalogue with optional filtering.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			asset_type	query		string	false	"Filter by asset type (stock, bond, fund, crypto, real_estate, income_source)"
// @Param			currency	query		string	false	"Filter by currency (e.g., USD, BRL)"
// @Success		200			{array}		domain.Asset
// @Failure		401			{object}	map[string]string	"Unauthorized"
// @Failure		500			{object}	map[string]string	"Internal server error"
// @Router			/v1/assets [get]
func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	params := domain.ListAssetsParams{}
	q := r.URL.Query()

	if t := q.Get("asset_type"); t != "" {
		at := domain.AssetType(t)
		params.AssetType = &at
	}
	if c := q.Get("currency"); c != "" {
		params.Currency = &c
	}

	assets, err := h.service.ListAssets(r.Context(), params)
	if err != nil {
		handleError(w, r, err, "failed to list assets")
		return
	}

	respondJSON(w, r, assets, http.StatusOK)
}

// GetAsset returns a single global asset.
// @Summary		Get asset
// @Description	Returns a single global asset by its ID, merging global values with tenant-specific overrides if they exist.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Asset ID"
// @Success		200	{object}	domain.Asset
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		404	{object}	map[string]string	"Asset not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/assets/{id} [get]
func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing asset id", http.StatusBadRequest)
		return
	}

	asset, err := h.service.GetAssetWithTenantConfig(r.Context(), tenantID, id)
	if err != nil {
		handleError(w, r, err, "failed to get asset")
		return
	}

	respondJSON(w, r, asset, http.StatusOK)
}

// CreateAsset creates a new global asset (Admin Only).
// @Summary		Create asset
// @Description	Creates a new global asset in the catalogue. Restricted to administrators.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			request	body		domain.CreateAssetInput	true	"Asset details"
// @Success		201		{object}	domain.Asset
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		422		{object}	map[string]string	"Validation error"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/v1/assets [post]
func (h *AssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateAssetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	asset, err := h.service.CreateAsset(r.Context(), input)
	if err != nil {
		handleError(w, r, err, "failed to create asset")
		return
	}

	respondJSON(w, r, asset, http.StatusCreated)
}

// DeleteAsset deletes a global asset (Admin Only).
// @Summary		Delete asset
// @Description	Soft-deletes a global asset from the catalogue. Restricted to administrators.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Asset ID"
// @Success		204	"No Content"
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/assets/{id} [delete]
func (h *AssetHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, r, "missing asset id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteAsset(r.Context(), id); err != nil {
		handleError(w, r, err, "failed to delete asset")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListMyConfigs returns the tenant's asset configurations.
// @Summary		List my asset configurations
// @Description	Returns all asset overrides (configs) for the current tenant.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Success		200	{array}		domain.TenantAssetConfig
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/v1/me/asset-configs [get]
func (h *AssetHandler) ListMyConfigs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	configs, err := h.service.ListTenantAssetConfigs(r.Context(), tenantID)
	if err != nil {
		handleError(w, r, err, "failed to list asset configs")
		return
	}

	respondJSON(w, r, configs, http.StatusOK)
}

// UpsertMyConfig creates or updates an asset configuration for the tenant.
// @Summary		Upsert asset configuration
// @Description	Creates or updates overrides for a global asset for the current tenant.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			asset_id	path		string							true	"Asset ID"
// @Param			request		body		domain.UpsertTenantAssetConfigInput	true	"Configuration details"
// @Success		200			{object}	domain.TenantAssetConfig
// @Failure		400			{object}	map[string]string				"Invalid request body"
// @Failure		401			{object}	map[string]string				"Unauthorized"
// @Failure		422			{object}	map[string]string				"Validation error"
// @Failure		500			{object}	map[string]string				"Internal server error"
// @Router			/v1/me/asset-configs/{asset_id} [put]
func (h *AssetHandler) UpsertMyConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	assetID := r.PathValue("asset_id")
	if assetID == "" {
		respondError(w, r, "missing asset id", http.StatusBadRequest)
		return
	}

	var input domain.UpsertTenantAssetConfigInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	input.AssetID = assetID // Force path param

	if err := h.validate.Struct(input); err != nil {
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	config, err := h.service.UpsertTenantAssetConfig(r.Context(), tenantID, input)
	if err != nil {
		handleError(w, r, err, "failed to upsert asset config")
		return
	}

	respondJSON(w, r, config, http.StatusOK)
}

// DeleteMyConfig deletes an asset configuration for the tenant.
// @Summary		Delete asset configuration
// @Description	Removes overrides for a global asset for the current tenant.
// @Tags			assets
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			asset_id	path		string	true	"Asset ID"
// @Success		204			"No Content"
// @Failure		401			{object}	map[string]string	"Unauthorized"
// @Failure		500			{object}	map[string]string	"Internal server error"
// @Router			/v1/me/asset-configs/{asset_id} [delete]
func (h *AssetHandler) DeleteMyConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantIDFromCtx(r.Context())
	if !ok {
		respondError(w, r, "unauthorized", http.StatusUnauthorized)
		return
	}

	assetID := r.PathValue("asset_id")
	if assetID == "" {
		respondError(w, r, "missing asset id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTenantAssetConfig(r.Context(), tenantID, assetID); err != nil {
		handleError(w, r, err, "failed to delete asset config")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
