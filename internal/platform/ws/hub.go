package ws

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients per tenant
	rooms      map[string]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mu         sync.RWMutex

	// Config
	maxConnsPerTenant int
}

// NewHub creates a new WebSocket hub.
func NewHub(maxConnsPerTenant int) *Hub {
	return &Hub{
		rooms:             make(map[string]map[*Client]struct{}),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan Event, 256),
		maxConnsPerTenant: maxConnsPerTenant,
	}
}

// Run starts the hub's main event loop.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case event := <-h.broadcast:
			h.deliverEvent(event)
		}
	}
}

// Publish sends an event to the hub for broadcasting.
func (h *Hub) Publish(event Event) error {
	select {
	case h.broadcast <- event:
		return nil
	default:
		return fmt.Errorf("ws: broadcast channel full")
	}
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[c.tenantID]; !ok {
		h.rooms[c.tenantID] = make(map[*Client]struct{})
	}

	if len(h.rooms[c.tenantID]) >= h.maxConnsPerTenant {
		// Enforce policy violation
		if err := c.conn.WriteMessage(websocket.ClosePolicyViolation, []byte("too many connections per tenant")); err != nil {
			slog.Error("ws: write policy violation message failed", "error", err)
		}
		if err := c.conn.Close(); err != nil {
			slog.Error("ws: close connection failed after policy violation", "error", err)
		}
		return
	}

	h.rooms[c.tenantID][c] = struct{}{}
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[c.tenantID]; ok {
		if _, ok := room[c]; ok {
			delete(room, c)
			close(c.send)
			if len(room) == 0 {
				delete(h.rooms, c.tenantID)
			}
		}
	}
}

func (h *Hub) deliverEvent(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[event.TenantID]; ok {
		for client := range room {
			select {
			case client.send <- event.Payload:
			default:
				// If client buffer is full, unregister it
				go func(c *Client) {
					h.unregister <- c
				}(client)
			}
		}
	}
}
