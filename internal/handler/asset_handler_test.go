package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAssetHandler_ListAssets(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		assets := []domain.Asset{{ID: "1", Name: "Asset 1"}}
		svc.On("ListAssets", mock.Anything, mock.Anything).Return(assets, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/assets", nil)
		rr := httptest.NewRecorder()

		h.ListAssets(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.Asset
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Asset 1", resp[0].Name)
	})

	t.Run("with filters", func(t *testing.T) {
		t.Parallel()
		svc.On("ListAssets", mock.Anything, mock.MatchedBy(func(p domain.ListAssetsParams) bool {
			return p.AssetType != nil && *p.AssetType == domain.AssetTypeStock && p.Currency != nil && *p.Currency == "USD"
		})).Return([]domain.Asset{}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/assets?asset_type=stock&currency=USD", nil)
		rr := httptest.NewRecorder()

		h.ListAssets(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAssetHandler_GetAsset(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		asset := &domain.Asset{ID: "1", Name: "Asset 1"}
		svc.On("GetAssetWithTenantConfig", mock.Anything, tenantID, "1").Return(asset, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/assets/1", nil)
		req.SetPathValue("id", "1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAsset(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp domain.Asset
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Asset 1", resp.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		svc.On("GetAssetWithTenantConfig", mock.Anything, tenantID, "99").Return(nil, domain.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/assets/99", nil)
		req.SetPathValue("id", "99")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAsset(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestAssetHandler_CreateAsset(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateAssetInput{
			Ticker:    "AAPL",
			Name:      "Apple Inc",
			Currency:  "USD",
			AssetType: domain.AssetTypeStock,
		}
		asset := &domain.Asset{ID: "1", Ticker: "AAPL", Name: "Apple Inc"}

		svc.On("CreateAsset", mock.Anything, input).Return(asset, nil).Once()

		bodyRaw, err := json.Marshal(input)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/assets", bytes.NewReader(bodyRaw))
		rr := httptest.NewRecorder()

		h.CreateAsset(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp domain.Asset
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "AAPL", resp.Ticker)
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateAssetInput{
			Ticker: "", // Required
		}

		bodyRaw, err := json.Marshal(input)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/assets", bytes.NewReader(bodyRaw))
		rr := httptest.NewRecorder()

		h.CreateAsset(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestAssetHandler_DeleteAsset(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc.On("DeleteAsset", mock.Anything, "1").Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/v1/assets/1", nil)
		req.SetPathValue("id", "1")
		rr := httptest.NewRecorder()

		h.DeleteAsset(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

func TestAssetHandler_ListMyConfigs(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		configs := []domain.TenantAssetConfig{{ID: "c1", AssetID: "a1"}}
		svc.On("ListTenantAssetConfigs", mock.Anything, tenantID).Return(configs, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/v1/me/asset-configs", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.ListMyConfigs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.TenantAssetConfig
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/me/asset-configs", nil)
		rr := httptest.NewRecorder()

		h.ListMyConfigs(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestAssetHandler_UpsertMyConfig(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		customName := "Custom Name"
		input := domain.UpsertTenantAssetConfigInput{
			AssetID: "a1",
			Name:    &customName,
		}
		config := &domain.TenantAssetConfig{ID: "c1", AssetID: "a1", Name: &customName}

		svc.On("UpsertTenantAssetConfig", mock.Anything, tenantID, input).Return(config, nil).Once()

		bodyRaw, err := json.Marshal(input)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/v1/me/asset-configs/a1", bytes.NewReader(bodyRaw))
		req.SetPathValue("asset_id", "a1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.UpsertMyConfig(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp domain.TenantAssetConfig
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Custom Name", *resp.Name)
	})
}

func TestAssetHandler_DeleteMyConfig(t *testing.T) {
	t.Parallel()

	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)
	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc.On("DeleteTenantAssetConfig", mock.Anything, tenantID, "a1").Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/v1/me/asset-configs/a1", nil)
		req.SetPathValue("asset_id", "a1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.DeleteMyConfig(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

func TestAssetHandler_PathValueMissing(t *testing.T) {
	t.Parallel()
	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)
	tenantID := "tenant-1"

	t.Run("GetAsset missing id", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/v1/assets/", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		// No PathValue set
		rr := httptest.NewRecorder()
		h.GetAsset(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("DeleteAsset missing id", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodDelete, "/v1/assets/", nil)
		rr := httptest.NewRecorder()
		h.DeleteAsset(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UpsertMyConfig missing id", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPut, "/v1/me/asset-configs/", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.UpsertMyConfig(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("DeleteMyConfig missing id", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodDelete, "/v1/me/asset-configs/", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.DeleteMyConfig(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAssetHandler_ServiceErrors(t *testing.T) {
	t.Parallel()
	svc := &mocks.InvestmentService{}
	h := NewAssetHandler(svc)
	tenantID := "tenant-1"

	t.Run("ListAssets error", func(t *testing.T) {
		t.Parallel()
		svc.On("ListAssets", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
		req := httptest.NewRequest(http.MethodGet, "/v1/assets", nil)
		rr := httptest.NewRecorder()
		h.ListAssets(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("CreateAsset error", func(t *testing.T) {
		t.Parallel()
		input := domain.CreateAssetInput{Ticker: "T", Name: "N", Currency: "USD", AssetType: domain.AssetTypeStock}
		svc.On("CreateAsset", mock.Anything, input).Return(nil, assert.AnError).Once()
		bodyRaw, err := json.Marshal(input)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/assets", bytes.NewReader(bodyRaw))
		rr := httptest.NewRecorder()
		h.CreateAsset(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("UpsertMyConfig bad body", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPut, "/v1/me/asset-configs/a1", bytes.NewReader([]byte("invalid")))
		req.SetPathValue("asset_id", "a1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.UpsertMyConfig(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("CreateAsset bad body", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodPost, "/v1/assets", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()
		h.CreateAsset(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UpsertMyConfig error", func(t *testing.T) {
		t.Parallel()
		input := domain.UpsertTenantAssetConfigInput{AssetID: "a1"}
		svc.On("UpsertTenantAssetConfig", mock.Anything, tenantID, input).Return(nil, assert.AnError).Once()
		bodyRaw, err := json.Marshal(input)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/v1/me/asset-configs/a1", bytes.NewReader(bodyRaw))
		req.SetPathValue("asset_id", "a1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.UpsertMyConfig(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("DeleteAsset error", func(t *testing.T) {
		t.Parallel()
		svc.On("DeleteAsset", mock.Anything, "1").Return(assert.AnError).Once()
		req := httptest.NewRequest(http.MethodDelete, "/v1/assets/1", nil)
		req.SetPathValue("id", "1")
		rr := httptest.NewRecorder()
		h.DeleteAsset(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("ListMyConfigs error", func(t *testing.T) {
		t.Parallel()
		svc.On("ListTenantAssetConfigs", mock.Anything, tenantID).Return(nil, assert.AnError).Once()
		req := httptest.NewRequest(http.MethodGet, "/v1/me/asset-configs", nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.ListMyConfigs(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("DeleteMyConfig error", func(t *testing.T) {
		t.Parallel()
		svc.On("DeleteTenantAssetConfig", mock.Anything, tenantID, "a1").Return(assert.AnError).Once()
		req := httptest.NewRequest(http.MethodDelete, "/v1/me/asset-configs/a1", nil)
		req.SetPathValue("asset_id", "a1")
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		h.DeleteMyConfig(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

//go:fix inline
//nolint:unused
func ptr[T any](v T) *T {
	return new(v)
}
