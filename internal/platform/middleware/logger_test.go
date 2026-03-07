package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	t.Run("logs every request with correct fields", func(t *testing.T) {
		t.Parallel()
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

		handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("ok"))
		}))

		req := httptest.NewRequest(http.MethodPost, "/test-path", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))

		var logEntry struct {
			Msg       string `json:"msg"`
			RequestID string `json:"request_id"`
			Method    string `json:"method"`
			Path      string `json:"path"`
			TenantID  string `json:"tenant_id"`
			UserID    string `json:"user_id"`
			Status    int    `json:"status"`
			Latency   int64  `json:"latency_ms"`
		}

		err := json.Unmarshal(logBuf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "request", logEntry.Msg)
		assert.Equal(t, http.MethodPost, logEntry.Method)
		assert.Equal(t, "/test-path", logEntry.Path)
		assert.Equal(t, http.StatusCreated, logEntry.Status)
		assert.GreaterOrEqual(t, logEntry.Latency, int64(0))
		assert.NotEmpty(t, logEntry.RequestID)
	})

	t.Run("captures authentic identity from context", func(t *testing.T) {
		t.Parallel()
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

		handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		tenantID := "01AN4Z0QN6W9NHY969YDSXDE5C"
		userID := "01AN4Z0QN6W9NHY969YDSXDE5D"

		ctx := context.WithValue(context.Background(), tenantIDKey, tenantID)
		ctx = context.WithValue(ctx, userIDKey, userID)

		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		var logEntry map[string]any
		err := json.Unmarshal(logBuf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, tenantID, logEntry["tenant_id"])
		assert.Equal(t, userID, logEntry["user_id"])
	})

	t.Run("handles default 200 status correctly", func(t *testing.T) {
		t.Parallel()
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

		handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// No WriteHeader called
			_, _ = w.Write([]byte("ok"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		var logEntry map[string]any
		err := json.Unmarshal(logBuf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.InDelta(t, float64(http.StatusOK), logEntry["status"], 0.01) // JSON unmarshals ints to float64 for map[string]any
	})
}
