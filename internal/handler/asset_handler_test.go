package handler

import (
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

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.InvestmentService{}
		h := NewAssetHandler(svc)

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
		svc := &mocks.InvestmentService{}
		h := NewAssetHandler(svc)

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

	tenantID := "tenant-1"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.InvestmentService{}
		h := NewAssetHandler(svc)

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
		svc := &mocks.InvestmentService{}
		h := NewAssetHandler(svc)

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
