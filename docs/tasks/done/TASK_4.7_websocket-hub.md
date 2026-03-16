# Task 4.7 — WebSocket Hub: Server-Side Broadcast Infrastructure

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Real-Time
> **Status:** ✅ `done`
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

- [x] `internal/platform/ws/hub.go` — central hub struct; per-tenant rooms (map of tenant ID → set of clients); `Register`, `Unregister`, `Broadcast` methods; goroutine-safe.
- [x] `internal/platform/ws/client.go` — individual WebSocket client; read pump (handles ping/pong and graceful close); write pump (serialises outgoing messages to the write buffer).
- [x] `internal/platform/ws/event.go` — `Event` struct with `TenantID`, `Type`, and `Payload` (pre-rendered HTML string for OOB swaps).
- [x] `internal/platform/ws/publisher.go` — `Publisher` interface for sending events to the hub; in-process `InProcessPublisher` implementation.
- [x] WebSocket upgrade handler `GET /ws` in the web server: validates auth cookie (reuses web auth middleware); registers client in hub.
- [x] Client-side Alpine.js plugin (`web/static/js/ws.js`): opens WebSocket, reconnects with exponential backoff, injects received HTML fragments into the HTMX swap pipeline.
- [x] Hub unit tests (`hub_test.go`): test Register, Unregister, Broadcast to single and multiple clients; test that messages are not delivered to unregistered clients; test tenant isolation (message for tenant A is not delivered to tenant B).
- [x] Connection limit: max 10 concurrent connections per tenant (configurable via `WS_MAX_CONNECTIONS_PER_TENANT` env var); return `1008 Policy Violation` on exceed.

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

---

## 5. Acceptance Criteria

- [x] Hub compiles and passes all unit tests (Register, Unregister, Broadcast, tenant isolation).
- [x] `Publish` does not block if no clients are connected for the target tenant.
- [x] Messages for tenant A are never delivered to clients of tenant B (isolation enforced in tests).
- [x] WebSocket upgrade requires a valid `moolah_token` cookie; unauthenticated upgrade returns `401` or `403`.
- [x] Max connections per tenant is enforced; exceeding limit returns close code `1008`.
- [x] Client write pump sets write deadlines to prevent slow-client goroutine leaks.
- [x] Client read pump handles pong frames and graceful close.
- [x] `ws.js` reconnects after connection loss with exponential backoff (max 30s).
- [x] `ws.js` feeds received HTML fragments into HTMX OOB swap pipeline.
- [x] `golangci-lint run ./internal/platform/ws/...` passes.
- [x] `gosec ./internal/platform/ws/...` passes.
- [x] `docs/ROADMAP.md` row 4.7 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                         |
| ---------- | ------ | ---------------------------------------------- |
| 2026-03-15 | Agent  | Task implemented, tested, and marked as done.  |
