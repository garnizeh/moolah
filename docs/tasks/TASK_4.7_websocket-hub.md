# Task 4.7 — WebSocket Hub: Server-Side Broadcast Infrastructure

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Real-Time
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the WebSocket broadcast hub in `internal/platform/ws/`. The hub manages per-tenant connection rooms: when the API side pushes an event (e.g. balance changed, income received, snapshot completed), all connected browser clients for that tenant receive the update in real time. On the client side, Alpine.js manages the WebSocket connection with automatic reconnect. HTMX OOB (out-of-band) swaps are used to update specific DOM fragments without a full page reload.

---

## 2. Context & Motivation

Without real-time push, users must refresh the page to see balance changes after a transaction is created or to see new income events. A WebSocket hub solves this elegantly at minimal complexity cost.

**Architecture:**

- The web server (`cmd/web/`) runs the hub.
- Browser clients connect to `GET /ws` (upgraded to WebSocket).
- The API server (or background jobs) can publish events to the hub via a shared in-process channel or Redis pub/sub (if deployed as separate processes).
- Events are HTML fragments (Templ-rendered partials) with HTMX OOB `hx-swap-oob` attributes, so the browser applies them without any client-side JavaScript parsing.

**Scope for MVP:** Single-process deployment (web and API in the same binary or co-located). Redis pub/sub is designed as an interface so it can be swapped in later without breaking the hub.

**Depends on:** Task 4.2 (web server), Task 4.6 (authentication — WebSocket connections require auth).

---

## 3. Scope

### In scope

- [ ] `internal/platform/ws/hub.go` — central hub struct; per-tenant rooms (map of tenant ID → set of clients); `Register`, `Unregister`, `Broadcast` methods; goroutine-safe.
- [ ] `internal/platform/ws/client.go` — individual WebSocket client; read pump (handles ping/pong and graceful close); write pump (serialises outgoing messages to the write buffer).
- [ ] `internal/platform/ws/event.go` — `Event` struct with `TenantID`, `Type`, and `Payload` (pre-rendered HTML string for OOB swaps).
- [ ] `internal/platform/ws/publisher.go` — `Publisher` interface for sending events to the hub; in-process `InProcessPublisher` implementation.
- [ ] WebSocket upgrade handler `GET /ws` in the web server: validates auth cookie (reuses web auth middleware); registers client in hub.
- [ ] Client-side Alpine.js plugin (`web/static/js/ws.js`): opens WebSocket, reconnects with exponential backoff, injects received HTML fragments into the HTMX swap pipeline.
- [ ] Hub unit tests (`hub_test.go`): test Register, Unregister, Broadcast to single and multiple clients; test that messages are not delivered to unregistered clients; test tenant isolation (message for tenant A is not delivered to tenant B).
- [ ] Connection limit: max 10 concurrent connections per tenant (configurable via `WS_MAX_CONNECTIONS_PER_TENANT` env var); return `1008 Policy Violation` on exceed.

### Out of scope

- Redis pub/sub backend (deferred — interface is defined, in-process impl only for MVP).
- Presence/typing indicators.
- Message persistence / replay (events are fire-and-forget).
- Binary WebSocket frames (text frames only; HTML payloads).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                    | Purpose                                           |
| ------ | --------------------------------------- | ------------------------------------------------- |
| CREATE | `internal/platform/ws/hub.go`          | Connection hub; per-tenant rooms                  |
| CREATE | `internal/platform/ws/client.go`       | Individual WebSocket connection lifecycle         |
| CREATE | `internal/platform/ws/event.go`        | Event type definitions                            |
| CREATE | `internal/platform/ws/publisher.go`    | Publisher interface + in-process implementation   |
| CREATE | `internal/platform/ws/hub_test.go`     | Unit tests: broadcast, isolation, cleanup         |
| CREATE | `web/static/js/ws.js`                  | Alpine.js WebSocket plugin (reconnect logic)      |
| MODIFY | `cmd/web/main.go`                      | Start hub goroutine; register `/ws` route         |

### Hub interface

```go
// internal/platform/ws/publisher.go

// Publisher is the interface through which other services push events to connected clients.
type Publisher interface {
    // Publish sends an event to all WebSocket clients of the given tenant.
    Publish(ctx context.Context, event Event) error
}

// Event represents a push notification to be delivered to browser clients.
type Event struct {
    TenantID string // target tenant; required
    Type     string // e.g. "balance_updated", "income_received", "snapshot_ready"
    Payload  string // pre-rendered HTML fragment with hx-swap-oob attribute
}
```

### Hub structure

```go
// internal/platform/ws/hub.go

type Hub struct {
    // Registered clients per tenant
    rooms map[string]map[*Client]struct{}
    mu    sync.RWMutex

    // Channels
    register   chan *Client
    unregister chan *Client
    broadcast  chan Event

    // Config
    maxConnsPerTenant int
}

func NewHub(maxConnsPerTenant int) *Hub { ... }

// Run starts the hub event loop. Must be called in a goroutine.
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

func (h *Hub) Publish(_ context.Context, event Event) error {
    select {
    case h.broadcast <- event:
        return nil
    default:
        return fmt.Errorf("ws: broadcast channel full")
    }
}
```

### Client read/write pumps

```go
// internal/platform/ws/client.go

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 512 // bytes — clients only send pong frames
)

type Client struct {
    hub      *Hub
    tenantID string
    conn     *websocket.Conn // net/http upgraded connection
    send     chan string      // outgoing HTML fragments
}

// writePump serialises messages from send channel to the WebSocket connection.
func (c *Client) writePump() { ... }

// readPump reads from WebSocket (only pong/close frames expected from browser).
func (c *Client) readPump() { ... }
```

### Client-side Alpine.js plugin (`ws.js`)

```javascript
// web/static/js/ws.js
// Automatically connects to /ws and feeds received HTML into HTMX OOB swap.
document.addEventListener('alpine:init', () => {
    Alpine.plugin((Alpine) => {
        let ws, reconnectDelay = 1000;

        function connect() {
            ws = new WebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`);

            ws.onmessage = (event) => {
                // HTMX OOB: inject HTML fragment directly into the DOM
                htmx.process(htmx.parseHTML(`<div>${event.data}</div>`)[0]);
            };

            ws.onclose = () => {
                // Exponential backoff reconnect (max 30s)
                setTimeout(connect, Math.min(reconnectDelay *= 2, 30000));
            };

            ws.onopen = () => { reconnectDelay = 1000; };
        }

        // Only connect on authenticated pages (check for auth cookie presence)
        if (document.cookie.includes('moolah_token')) connect();
    });
});
```

### Event types to support at launch

| Type | Trigger | OOB Target |
| ---- | ------- | ---------- |
| `balance_updated` | Transaction created/deleted | `#account-balance-{id}` card fragment |
| `income_received` | MarkIncomeReceived API call | `#income-event-{id}` row fragment |
| `snapshot_ready` | Portfolio snapshot job completes | `#portfolio-summary` card fragment |

### WebSocket route registration

```go
// In cmd/web/main.go
hub := ws.NewHub(cfg.WSMaxConnsPerTenant)
go hub.Run(ctx)

mux.Handle("GET /ws", wsMiddleware(hub, tokenValidator))
```

---

## 5. Acceptance Criteria

- [ ] Hub compiles and passes all unit tests (Register, Unregister, Broadcast, tenant isolation).
- [ ] `Publish` does not block if no clients are connected for the target tenant.
- [ ] Messages for tenant A are never delivered to clients of tenant B (isolation enforced in tests).
- [ ] WebSocket upgrade requires a valid `moolah_token` cookie; unauthenticated upgrade returns `403`.
- [ ] Max connections per tenant is enforced; exceeding limit returns close code `1008`.
- [ ] Client write pump sets write deadlines to prevent slow-client goroutine leaks.
- [ ] Client read pump handles pong frames and graceful close.
- [ ] `ws.js` reconnects after connection loss with exponential backoff (max 30s).
- [ ] `ws.js` feeds received HTML fragments into HTMX OOB swap pipeline.
- [ ] `golangci-lint run ./internal/platform/ws/...` passes.
- [ ] `gosec ./internal/platform/ws/...` passes.
- [ ] `docs/ROADMAP.md` row 4.7 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-15 | —      | Task created (new) |
