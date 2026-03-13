package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMasterPurchaseHandler_Create(t *testing.T) {
	t.Parallel()

	tenantID := "tenant_123"
	userID := "user_456"
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := &mocks.MasterPurchaseService{}
		h := NewMasterPurchaseHandler(svc)

		reqBody := CreateMasterPurchaseRequest{
			AccountID:            "acc_1",
			CategoryID:           "cat_1",
			Description:          "iPhone",
			TotalAmountCents:     120000,
			InstallmentCount:     12,
			ClosingDay:           10,
			FirstInstallmentDate: now,
		}

		svc.On("Create", mock.Anything, tenantID, mock.Anything).Return(&domain.MasterPurchase{ID: "mp_1"}, nil)

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("validation failed", func(t *testing.T) {
		t.Parallel()
		h := NewMasterPurchaseHandler(nil)

		reqBody := CreateMasterPurchaseRequest{
			Description: "", // Invalid
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/v1/master-purchases", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}
