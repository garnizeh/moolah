# Task 1.1.15 — platform/idempotency/redis_store.go: IdempotencyStore Redis Implementation

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the concrete `IdempotencyStore` interface (defined in task 1.1.14) backed by Redis using `github.com/redis/go-redis/v9`. The store uses `SETNX` for atomic lock acquisition and `SET` to persist final responses. A companion integration test validates the full lifecycle (miss → lock → cache → replay → expiry) against a real Redis instance spun up via `testcontainers-go`.

---

## 2. Context & Motivation

Task 1.1.14 defines the `IdempotencyStore` interface and the middleware logic. This task delivers the production-ready Redis-backed implementation that satisfies that interface. The separation keeps the middleware package free of Redis dependencies and fully unit-testable with mocks.

Key Redis operations:

- `GET key` — check for a cached response or in-flight lock.
- `SET key "locked" NX EX ttl` — atomically acquire processing lock (SETNX semantics via SET NX).
- `SET key <json> EX ttl` — store the final response, replacing the lock.

- Architecture reference: `docs/ARCHITECTURE.md` — Section 7.7 Idempotency (Redis Key Lifecycle)
- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.15
- Depends on: task 1.1.14 (`IdempotencyStore` interface)

---

## 3. Scope

### In scope

- [ ] `internal/platform/idempotency/redis_store.go` — `RedisStore` struct implementing `middleware.IdempotencyStore`
- [ ] `NewRedisStore(client *redis.Client) *RedisStore` constructor
- [ ] `Get`, `SetLocked`, `SetResponse` method implementations
- [ ] Serialization/deserialization of `CachedResponse` as JSON in Redis
- [ ] `internal/platform/idempotency/redis_store_integration_test.go` — testcontainers-go integration tests (build tag `//go:build integration`)

### Out of scope

- Middleware logic — covered by task 1.1.14
- Route wiring — covered by task 1.5.11
- Redis Sentinel / Cluster config (deferred to Phase 5 — 5.6)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                            | Purpose                                               |
| ------ | --------------------------------------------------------------- | ----------------------------------------------------- |
| CREATE | `internal/platform/idempotency/redis_store.go`                  | Concrete `IdempotencyStore` backed by Redis           |
| CREATE | `internal/platform/idempotency/redis_store_integration_test.go` | Integration tests using `testcontainers-go`           |

### Key interfaces / types

```go
// internal/platform/idempotency/redis_store.go
package idempotency

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/garnizeh/moolah/internal/platform/middleware"
)

const lockedSentinel = "locked"

// RedisStore implements middleware.IdempotencyStore using Redis.
type RedisStore struct {
    client *redis.Client
}

// NewRedisStore returns a new RedisStore wrapping the provided Redis client.
func NewRedisStore(client *redis.Client) *RedisStore {
    return &RedisStore{client: client}
}

// Get retrieves a previously cached response.
// Returns (nil, nil) on a cache miss; returns (nil, err) on a Redis error.
// Returns (nil, nil) if the key holds the in-flight "locked" sentinel
// (the middleware interprets this as an in-flight key via SetLocked).
func (s *RedisStore) Get(ctx context.Context, key string) (*middleware.CachedResponse, error) { ... }

// SetLocked atomically acquires the processing lock using SET NX EX.
// Returns true if the lock was acquired; false if already held.
func (s *RedisStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) { ... }

// SetResponse stores the final response JSON, replacing the processing lock.
func (s *RedisStore) SetResponse(ctx context.Context, key string, resp middleware.CachedResponse, ttl time.Duration) error { ... }
```

### Redis key lifecycle

```text
t=0       SetLocked("idempotency:u:01HZ...", 24h)  → SET NX EX  → "locked"
t=1       SetResponse(key, {201, body}, 24h)        → SET EX     → JSON
t=5       Get("idempotency:u:01HZ...")              → GET        → JSON → CachedResponse
t=24h+1   Key expires automatically
```

### Serialization format

`CachedResponse` is stored as a JSON string in Redis:

```json
{"status_code": 201, "body": "<base64-encoded body bytes>"}
```

> `Body` is marshaled as a JSON byte array (Go `[]byte` → base64 in JSON encoding). No special handling needed beyond standard `encoding/json`.

### SQL queries (sqlc)

N/A — this task is Redis-only.

### API endpoints (if applicable)

N/A — infrastructure component; no HTTP endpoints.

### Error cases to handle

| Scenario                              | Behaviour                                                      |
| ------------------------------------- | -------------------------------------------------------------- |
| Redis unavailable on `Get`            | Return `(nil, err)` — middleware responds 500                  |
| Redis unavailable on `SetLocked`      | Return `(false, err)` — middleware responds 500                |
| Redis unavailable on `SetResponse`    | Return `err` — middleware logs warning; response still sent    |
| Key holds `"locked"` sentinel on Get  | Return `(nil, nil)` — caller (SetLocked) handles via NX check  |
| JSON unmarshal error on Get           | Return `(nil, err)` — treat as unrecoverable cache corruption  |

---

## 5. Acceptance Criteria

- [ ] `RedisStore` implements `middleware.IdempotencyStore` (compile-time interface assertion: `var _ middleware.IdempotencyStore = (*RedisStore)(nil)`).
- [ ] `Get` on a missing key returns `(nil, nil)`.
- [ ] `Get` on a `"locked"` key returns `(nil, nil)` (treated as miss — `SetLocked` determines in-flight state).
- [ ] `Get` on a cached response key returns a correctly deserialized `*CachedResponse`.
- [ ] `SetLocked` on a fresh key returns `(true, nil)` and sets the key with the correct TTL.
- [ ] `SetLocked` on an already-locked key returns `(false, nil)`.
- [ ] `SetResponse` replaces the lock value with the serialized `CachedResponse` and refreshes the TTL.
- [ ] After TTL expires, `Get` returns `(nil, nil)` (key no longer exists).
- [ ] All exported types and functions have Go doc comments.
- [ ] Compile-time interface assertion present in source file.
- [ ] Integration tests use `//go:build integration` build tag and `testcontainers-go` Redis container.
- [ ] Test coverage for `redis_store.go` ≥ 90%.
- [ ] `golangci-lint run ./internal/platform/idempotency/...` passes with zero issues.
- [ ] `gosec ./internal/platform/idempotency/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.15 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                                  | Type       | Status          |
| ----------------------------------------------------------- | ---------- | --------------- |
| `github.com/redis/go-redis/v9` in `go.mod` / `vendor/`     | External   | needs adding     |
| `middleware.IdempotencyStore` + `middleware.CachedResponse` | Upstream   | 🔵 backlog (1.1.14) |
| `testcontainers-go` (Redis module) in `go.mod`              | Test-only  | ✅ done (in vendor) |
| Route wiring (task 1.5.11)                                  | Downstream | 🔵 backlog       |

---

## 7. Testing Plan

### Integration tests (`redis_store_integration_test.go`, `//go:build integration`)

Spin up a Redis container with `testcontainers-go` (`testcontainers/redis`). Use a dedicated key prefix per test to avoid collisions.

- **Get miss:** call `Get` on a non-existent key; assert `(nil, nil)`.
- **SetLocked → Get (in-flight):** call `SetLocked`; call `Get`; assert `(nil, nil)` (locked sentinel is opaque to `Get`).
- **SetLocked twice:** first call returns `(true, nil)`; second call on the same key returns `(false, nil)`.
- **SetResponse → Get (cache hit):** call `SetLocked`; call `SetResponse` with a `CachedResponse{201, []byte(`{"id":"x"}`)}}`; call `Get`; assert returned struct matches.
- **TTL expiry:** `SetLocked` with 1-second TTL; `time.Sleep(2s)`; `Get`; assert `(nil, nil)`.
- **SetLocked refresh:** `SetResponse` overwrites the lock; a subsequent `SetLocked` on the same key returns `(false, nil)` (response already cached).

### Unit tests

N/A — `RedisStore` is a thin integration wrapper; all logic is tested via integration tests against a real Redis container. Mocking Redis internals would add no confidence.

---

## 8. Change Log

| Date | Author | Change |
| ---- | ------ | ------ |
| 2026-03-07 | — | Task created |
