package ws

import (
	"context"
)

// Event represents a push notification to be delivered to browser clients.
type Event struct {
	TenantID string // target tenant; required
	Type     string // e.g. "balance_updated", "income_received", "snapshot_ready"
	Payload  string // pre-rendered HTML fragment with hx-swap-oob attribute
}

// Publisher is the interface through which other services push events to connected clients.
type Publisher interface {
	// Publish sends an event to all WebSocket clients of the given tenant.
	Publish(ctx context.Context, event Event) error
}

// InProcessPublisher is a basic implementation of Publisher that forwards events to the Hub.
type InProcessPublisher struct {
	hub *Hub
}

// NewInProcessPublisher creates a new InProcessPublisher.
func NewInProcessPublisher(hub *Hub) *InProcessPublisher {
	return &InProcessPublisher{hub: hub}
}

// Publish sends an event to the Hub's broadcast channel.
func (p *InProcessPublisher) Publish(_ context.Context, event Event) error {
	return p.hub.Publish(event)
}
