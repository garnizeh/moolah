package ws

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (clients only send pong/close frames).
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin should be more restrictive in production.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the WebSocket connection and the hub.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan string
	tenantID string
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, tenantID string, conn *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		tenantID: tenantID,
		conn:     conn,
		send:     make(chan string, 256),
	}
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		select {
		case c.hub.unregister <- c:
		default:
			// Fallback: the hub loop might already be terminating or the channel is blocked
			// but we must not hang the readPump.
		}
		if err := c.conn.Close(); err != nil {
			slog.Error("ws: close connection failed in readPump", "error", err)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		slog.Error("ws: set read deadline failed", "error", err)
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			slog.Error("ws: set read deadline failed in pong", "error", err)
		}
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("ws: read error", "error", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			slog.Error("ws: close connection failed in writePump", "error", err)
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Error("ws: set write deadline failed", "error", err)
			}
			if !ok {
				// The hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					slog.Error("ws: write close message failed", "error", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				slog.Error("ws: next writer failed", "error", err)
				return
			}
			if _, err := w.Write([]byte(message)); err != nil {
				slog.Error("ws: write message failed", "error", err)
				_ = w.Close()
				return
			}

			if err := w.Close(); err != nil {
				slog.Error("ws: close writer failed", "error", err)
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Error("ws: set write deadline failed in ping", "error", err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("ws: write ping message failed", "error", err)
				return
			}
		}
	}
}

// UpgradeHandler returns an http.HandlerFunc that upgrades the connection to WebSocket.
func UpgradeHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value(middleware.TenantIDKey).(string)
		if !ok || tenantID == "" {
			http.Error(w, "Unauthorized: missing tenant_id", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("ws: upgrade failed", "error", err)
			return
		}

		client := NewClient(hub, tenantID, conn)
		hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}
