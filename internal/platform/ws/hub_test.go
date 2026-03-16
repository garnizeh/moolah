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

func TestHub_Broadcast(t *testing.T) {
	t.Parallel()

	hub := NewHub(10)
	ctx := t.Context()
	go hub.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.TenantIDKey, "tenant-1")
		UpgradeHandler(hub)(w, r.WithContext(ctx))
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}

	conn1, res1, err1 := dialer.Dial(wsURL, nil)
	require.NoError(t, err1)
	defer func() { _ = conn1.Close() }()
	if res1 != nil {
		_ = res1.Body.Close()
	}

	conn2, res2, err2 := dialer.Dial(wsURL, nil)
	require.NoError(t, err2)
	defer func() { _ = conn2.Close() }()
	if res2 != nil {
		_ = res2.Body.Close()
	}

	event := Event{
		TenantID: "tenant-1",
		Type:     "test",
		Payload:  "<div id='test'>Hello</div>",
	}
	errPublish := hub.Publish(event)
	require.NoError(t, errPublish)

	for _, conn := range []*websocket.Conn{conn1, conn2} {
		if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Fatalf("failed to set read deadline: %v", err)
		}
		_, message, errRead := conn.ReadMessage()
		require.NoError(t, errRead)
		assert.Equal(t, event.Payload, string(message))
	}
}

func TestHub_TenantIsolation(t *testing.T) {
	t.Parallel()

	hub := NewHub(10)
	ctx := t.Context()
	go hub.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant")
		ctx := context.WithValue(r.Context(), middleware.TenantIDKey, tenantID)
		UpgradeHandler(hub)(w, r.WithContext(ctx))
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}

	conn1, res1, err1 := dialer.Dial(wsURL+"?tenant=tenant-1", nil)
	require.NoError(t, err1)
	defer func() { _ = conn1.Close() }()
	if res1 != nil {
		_ = res1.Body.Close()
	}

	conn2, res2, err2 := dialer.Dial(wsURL+"?tenant=tenant-2", nil)
	require.NoError(t, err2)
	defer func() { _ = conn2.Close() }()
	if res2 != nil {
		_ = res2.Body.Close()
	}

	event := Event{
		TenantID: "tenant-1",
		Payload:  "secret-1",
	}
	errPublish := hub.Publish(event)
	require.NoError(t, errPublish)

	if err := conn1.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}
	_, msg, errRead := conn1.ReadMessage()
	require.NoError(t, errRead)
	assert.Equal(t, "secret-1", string(msg))

	if errDeadline := conn2.SetReadDeadline(time.Now().Add(200 * time.Millisecond)); errDeadline != nil {
		t.Fatalf("failed to set read deadline: %v", errDeadline)
	}
	_, _, errRead2 := conn2.ReadMessage()
	require.Error(t, errRead2)
}

func TestHub_MaxConnections(t *testing.T) {
	t.Parallel()

	hub := NewHub(1)
	ctx := t.Context()
	go hub.Run(ctx)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.TenantIDKey, "tenant-1")
		UpgradeHandler(hub)(w, r.WithContext(ctx))
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}

	// Conn 1: OK
	conn1, res1, err1 := dialer.Dial(wsURL, nil)
	require.NoError(t, err1)
	defer func() { _ = conn1.Close() }()
	if res1 != nil {
		_ = res1.Body.Close()
	}

	// Conn 2: Should be rejected
	conn2, res2, err2 := dialer.Dial(wsURL, nil)
	if err2 == nil {
		defer func() { _ = conn2.Close() }()
		if res2 != nil {
			_ = res2.Body.Close()
		}

		if err := conn2.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			t.Fatalf("failed to set read deadline: %v", err)
		}
		_, _, errRead := conn2.ReadMessage()
		require.Error(t, errRead)
	} else if res2 != nil {
		_ = res2.Body.Close()
	}
}

func TestHub_PublishFull(t *testing.T) {
	t.Parallel()

	hub := NewHub(1)
	// Do NOT start go hub.Run(ctx)

	for range 256 {
		err := hub.Publish(Event{TenantID: "t", Payload: "p"})
		require.NoError(t, err)
	}

	err := hub.Publish(Event{TenantID: "t", Payload: "p"})
	require.Error(t, err)
}
