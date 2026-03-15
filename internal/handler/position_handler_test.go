package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPositionHandler_List(t *testing.T) {
	t.Parallel()
	noopHandler := slog.NewTextHandler(io.Discard, nil)
	slog.SetDefault(slog.New(noopHandler))

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.InvestmentService)
		h := NewPositionHandler(service)

		tenantID := "tenant-1"
		positions := []domain.Position{{ID: "pos-1", AssetID: "asset-1"}}
		service.On("ListPositions", mock.Anything, tenantID).Return(positions, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/positions", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.List(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp []domain.Position
		err := json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		require.Len(t, resp, 1)
	})
}

func TestPositionHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.InvestmentService)
		h := NewPositionHandler(service)

		tenantID := "tenant-1"
		reqBody := CreatePositionRequest{
			AssetID:        "asset-1",
			AccountID:      "acc-1",
			Quantity:       decimal.NewFromFloat(10.5),
			AvgCostCents:   1000,
			LastPriceCents: 1100,
			Currency:       "USD",
			PurchasedAt:    time.Now(),
			IncomeType:     domain.IncomeTypeDividend,
		}

		service.On("CreatePosition", mock.Anything, tenantID, mock.MatchedBy(func(in domain.CreatePositionInput) bool {
			return in.AssetID == reqBody.AssetID && in.Quantity.Equal(reqBody.Quantity)
		})).Return(&domain.Position{ID: "pos-1"}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPost, "/v1/positions", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.InvestmentService)
		h := NewPositionHandler(service)

		reqBody := CreatePositionRequest{
			AssetID: "", // Invalid
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, "tenant-1")
		req := httptest.NewRequest(http.MethodPost, "/v1/positions", bytes.NewReader(body)).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestPositionHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.InvestmentService)
		h := NewPositionHandler(service)

		tenantID := "tenant-1"
		posID := "pos-1"
		newQty := decimal.NewFromInt(20)
		reqBody := UpdatePositionRequest{
			Quantity: &newQty,
		}

		service.On("UpdatePosition", mock.Anything, tenantID, posID, mock.MatchedBy(func(in domain.UpdatePositionInput) bool {
			return in.Quantity != nil && in.Quantity.Equal(newQty)
		})).Return(&domain.Position{ID: posID, Quantity: newQty}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/positions/pos-1", bytes.NewReader(body)).WithContext(ctx)
		req.SetPathValue("id", posID)
		rr := httptest.NewRecorder()

		h.Update(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestPositionHandler_MarkIncomeReceived(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.InvestmentService)
		h := NewPositionHandler(service)

		tenantID := "tenant-1"
		eventID := "evt-1"
		service.On("MarkIncomeReceived", mock.Anything, tenantID, eventID).Return(&domain.PositionIncomeEvent{ID: eventID, Status: domain.ReceivableStatusReceived}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodPatch, "/v1/income-events/evt-1/receive", nil).WithContext(ctx)
		req.SetPathValue("id", eventID)
		rr := httptest.NewRecorder()

		h.MarkIncomeReceived(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestPositionHandler_GetSummary(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		service := new(mocks.InvestmentService)
		h := NewPositionHandler(service)

		tenantID := "tenant-1"
		service.On("GetPortfolioSummary", mock.Anything, tenantID).Return(&domain.PortfolioSummary{TotalValueCents: 10000}, nil)

		ctx := context.WithValue(context.Background(), middleware.TenantIDKey, tenantID)
		req := httptest.NewRequest(http.MethodGet, "/v1/investments/summary", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetSummary(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})
}
