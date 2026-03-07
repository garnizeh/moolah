# Task 1.1.14 — platform/middleware/idempotency.go: Redis-Backed Idempotency-Key Middleware

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement an `http.Handler` middleware that prevents duplicate financial records by enforcing an `Idempotency-Key` header on all state-mutating `POST` endpoints. The middleware checks a Redis-backed store before each request executes, either replaying a previously cached response or acquiring a processing lock and delegating to the handler. The store dependency is injected as an interface (`IdempotencyStore`) to allow 100% mockable unit tests.

---

## 2. Context & Motivation

In financial systems a client may submit the same request more than once — double-clicking a "Pay" button, a mobile client retrying after a network timeout, or an automated job replaying on failure. Without a deduplication layer, each request creates a distinct record (duplicate transaction, account, etc.).

The strategy is a Redis-backed idempotency key scoped per authenticated user with a 24-hour TTL:

- **Miss:** Acquire a `SETNX` lock, execute the handler, cache the response.
- **In-flight (locked):** Return `409 Conflict` — another goroutine is already processing the same key.
- **Hit (response cached):** Return the exact cached status + body without touching the database.

- Architecture reference: `docs/ARCHITECTURE.md` — Section 7.7 Idempotency
- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.14
- Depends on: task 1.1.15 (`IdempotencyStore` Redis implementation)
- Applied in routes: task 1.5.11

---

## 3. Scope

### In scope

- [ ] `internal/platform/middleware/idempotency.go` — `Idempotency(store IdempotencyStore) func(http.Handler) http.Handler`
- [ ] `IdempotencyStore` interface (defined in same file; implemented in task 1.1.15)
- [ ] `CachedResponse` struct (status code + body bytes)
- [ ] `responseRecorder` — captures handler output for caching
- [ ] Missing `Idempotency-Key` header → `400 Bad Request` with `code: "missing_idempotency_key"`
- [ ] In-flight key → `409 Conflict` with `code: "idempotency_key_in_flight"`
- [ ] Key length validation: max 255 characters
- [ ] Redis key scoped per user: `idempotency:{userID}:{clientKey}` (user extracted from request context set by `RequireAuth`)
- [ ] `internal/platform/middleware/idempotency_test.go` — unit tests with mocked `IdempotencyStore`

### Out of scope

- Redis implementation (`IdempotencyStore`) — covered by task 1.1.15
- Route wiring — covered by task 1.5.11
- `GET`, `PATCH`, `DELETE` endpoints (naturally idempotent; no header required)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                    | Purpose                                              |
| ------ | ------------------------------------------------------- | ---------------------------------------------------- |
| CREATE | `internal/platform/middleware/idempotency.go`           | Middleware + `IdempotencyStore` interface + types    |
| CREATE | `internal/platform/middleware/idempotency_test.go`      | Unit tests with mock store                           |

### Key interfaces / types

```go
// internal/platform/middleware/idempotency.go
package middleware

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "time"
)

const (
    idempotencyHeader = "Idempotency-Key"
    idempotencyTTL    = 24 * time.Hour
    keyMaxLen         = 255
)

// IdempotencyStore is the Redis-backed store injected into the middleware.
// Defined here so the middleware package owns the abstraction; the concrete
// implementation lives in internal/platform/idempotency/redis_store.go (task 1.1.15).
type IdempotencyStore interface {
    // Get retrieves a previously cached response. Returns (nil, nil) on a cache miss.
    Get(ctx context.Context, key string) (*CachedResponse, error)
    // SetLocked atomically acquires the processing lock (SETNX).
    // Returns true if the lock was acquired, false if already held by another request.
    SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error)
    // SetResponse stores the final response, replacing the processing lock.
    SetResponse(ctx context.Context, key string, resp CachedResponse, ttl time.Duration) error
}

// CachedResponse holds the HTTP status and body to replay on subsequent requests.
type CachedResponse struct {
    StatusCode int    `json:"status_code"`
    Body       []byte `json:"body"`
}

// responseRecorder wraps http.ResponseWriter to capture the status code and
// response body so the middleware can persist them in Redis after the handler runs.
type responseRecorder struct {
    http.ResponseWriter
    statusCode int
    body       *bytes.Buffer
}

// Idempotency returns a middleware that deduplicates mutating POST requests using
// the Idempotency-Key header and a Redis-backed IdempotencyStore.
// Must be placed AFTER RequireAuth in the middleware chain so that the user ID
// is available in the request context for key scoping.
func Idempotency(store IdempotencyStore) func(http.Handler) http.Handler { ... }
```

### Redis key format

```text
idempotency:{userID}:{clientKey}
```

`userID` is extracted from the request context (injected by `RequireAuth` middleware).  
`clientKey` is the raw `Idempotency-Key` header value.

### Error responses

```json
// 400 — missing header
{ "error": { "code": "missing_idempotency_key", "message": "Idempotency-Key header is required." } }

// 400 — key too long
{ "error": { "code": "invalid_idempotency_key", "message": "Idempotency-Key must not exceed 255 characters." } }

// 409 — in-flight
{ "error": { "code": "idempotency_key_in_flight", "message": "A request with this Idempotency-Key is already being processed." } }
```

### SQL queries (sqlc)

N/A — middleware operates only against Redis.

### API endpoints (if applicable)

N/A — pure middleware; routes wired in task 1.5.11.

### Endpoints that will require this middleware

| Method | Path                    | Why                                            |
| ------ | ----------------------- | ---------------------------------------------- |
| `POST` | `/v1/transactions`      | Creates a financial record — must never duplicate |
| `POST` | `/v1/accounts`          | Creates an account                             |
| `POST` | `/v1/categories`        | Creates a category                             |
| `POST` | `/v1/master-purchases`  | Creates installment purchase (Phase 2)         |
| `POST` | `/v1/tenants/me/users`  | Invites a user — idempotent invite by email    |

### Error cases to handle

| Scenario                          | HTTP Status | Code                          |
| --------------------------------- | ----------- | ----------------------------- |
| `Idempotency-Key` header absent   | 400         | `missing_idempotency_key`     |
| `Idempotency-Key` exceeds 255 chars | 400       | `invalid_idempotency_key`     |
| Key currently locked (in-flight)  | 409         | `idempotency_key_in_flight`   |
| Key already processed (cache hit) | —           | Replay cached status + body   |
| Redis `Get` error                 | 500         | Pass through standard error   |
| Redis `SetLocked` error           | 500         | Pass through standard error   |
| Handler returns 5xx               | —           | Do NOT cache error responses  |

---

## 5. Acceptance Criteria

- [ ] `Idempotency-Key` header absent → `400` with `code: "missing_idempotency_key"`.
- [ ] `Idempotency-Key` > 255 characters → `400` with `code: "invalid_idempotency_key"`.
- [ ] First request (cache miss): handler executes; response cached in Redis; client receives handler's response.
- [ ] Duplicate request (cache hit): handler NOT called; client receives identical cached response.
- [ ] In-flight key (lock held): returns `409` with `code: "idempotency_key_in_flight"`.
- [ ] Handler `5xx` response is NOT cached (next retry will re-execute the handler).
- [ ] Redis key is scoped as `idempotency:{userID}:{clientKey}` (different users with same key are isolated).
- [ ] All exported types and functions have Go doc comments.
- [ ] `IdempotencyStore` interface is injected — no direct Redis import in this file.
- [ ] Unit tests use a mock `IdempotencyStore`; no real Redis required.
- [ ] Test coverage for `idempotency.go` ≥ 90%.
- [ ] `golangci-lint run ./internal/platform/middleware/...` passes with zero issues.
- [ ] `gosec ./internal/platform/middleware/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.14 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                                  | Type     | Status         |
| ----------------------------------------------------------- | -------- | -------------- |
| `RequireAuth` middleware (injects `userID` into context)    | Upstream | ✅ done (1.1.9) |
| `IdempotencyStore` Redis implementation (task 1.1.15)       | Upstream | 🔵 backlog      |
| Route wiring (task 1.5.11)                                  | Downstream | 🔵 backlog    |

---

## 7. Testing Plan

### Unit tests (`internal/platform/middleware/idempotency_test.go`, no build tag)

Use `httptest.NewRecorder`, `httptest.NewRequest`, and a mock `IdempotencyStore`.

- **Missing header:** request with no `Idempotency-Key`; assert `400` + correct error code.
- **Key too long:** header with 256-character key; assert `400` + correct error code.
- **Cache miss (first request):** mock `Get` returns `(nil, nil)`; mock `SetLocked` returns `(true, nil)`; stub handler returns `201`; assert client gets `201`; assert `SetResponse` called with correct status + body.
- **Cache hit (duplicate):** mock `Get` returns a `CachedResponse{201, body}`; assert handler NOT called; assert client gets `201` with same body.
- **In-flight (lock held):** mock `Get` returns `(nil, nil)`; mock `SetLocked` returns `(false, nil)`; assert `409` + correct error code; assert handler NOT called.
- **Handler 5xx not cached:** mock `Get` Miss + `SetLocked` OK; stub handler returns `500`; assert `SetResponse` NOT called.
- **Redis error on Get:** mock `Get` returns `(nil, err)`; assert `500`.

### Integration tests

N/A — the integration between middleware and real Redis is covered by task 1.1.15 (`redis_store_test.go`).

---

## 8. Change Log

| Date | Author | Change |
| ---- | ------ | ------ |
| 2026-03-07 | — | Task created |
