package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Lifecycle(t *testing.T) {
	t.Parallel()

	hub := NewHub(10)
	ctx := t.Context()
	go hub.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant")
		if tenantID == "" {
			http.Error(w, "missing tenant", http.StatusUnauthorized)
			return
		}
		// Injects tenantID into context as the middleware would
		ctx := context.WithValue(r.Context(), middleware.TenantIDKey, tenantID)
		UpgradeHandler(hub)(w, r.WithContext(ctx))
	}))
	t.Cleanup(func() {
		server.Close()
	})

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}

	t.Run("successful connection and message receipt", func(t *testing.T) {
		t.Parallel()
		conn, resp, err := dialer.Dial(wsURL+"?tenant=tenant-1", nil)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close()
		}()
		require.NoError(t, resp.Body.Close())

		// Test broadcast receipt
		payload := "<div>Update</div>"
		err = hub.Publish(Event{
			TenantID: "tenant-1",
			Payload:  payload,
		})
		require.NoError(t, err)

		_, message, errRead := conn.ReadMessage()
		require.NoError(t, errRead)
		assert.Equal(t, payload, string(message))
	})

	t.Run("unauthorized upgrade", func(t *testing.T) {
		t.Parallel()
		_, resp, err := dialer.Dial(wsURL, nil) // No tenant query param
		require.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			if resp.Body != nil {
				require.NoError(t, resp.Body.Close())
			}
		}
	})

	t.Run("ping pong heartbeat", func(t *testing.T) {
		t.Parallel()
		conn, resp, err := dialer.Dial(wsURL+"?tenant=tenant-ping", nil)
		require.NoError(t, err)
		defer func() {
			_ = conn.Close()
		}()
		require.NoError(t, resp.Body.Close())

		// At least verify we can send a pong back, connection remains open
		err = conn.WriteMessage(websocket.PongMessage, nil)
		require.NoError(t, err)
	})
}

func TestClient_ReadPumpExit(t *testing.T) {
	t.Parallel()

	hub := NewHub(10)
	ctx := t.Context()
	go hub.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.TenantIDKey, "tenant-exit")
		UpgradeHandler(hub)(w, r.WithContext(ctx))
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}

	conn, resp, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	// Closing the client connection should trigger hub unregistration
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	require.NoError(t, err)

	// Wait a bit for the cleanup
	time.Sleep(150 * time.Millisecond)

	// Should be zero clients
	hub.mu.RLock()
	room := hub.rooms["tenant-exit"]
	hub.mu.RUnlock()
	assert.Empty(t, room)
}
