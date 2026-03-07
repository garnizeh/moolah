# Task 1.1.1 — pkg/ulidutil: Thread-Safe Monotonic ULID Factory

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement a package-level ULID generator that is safe for concurrent use and produces monotonically increasing identifiers within the same millisecond. Every primary key in the system is a ULID string; this package is therefore a hard dependency for every other Phase 1 task.

---

## 2. Context & Motivation

The project mandates ULID for all primary keys (see `docs/ARCHITECTURE.md` — Identity section). ULIDs are 26-character, lexicographically sortable, and URL-safe. The standard `ulid.Make()` function is not goroutine-safe when using a monotonic entropy source; we need a wrapper that serialises access via a mutex.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.1

---

## 3. Scope

### In scope

- [ ] `pkg/ulidutil/ulid.go` — exported `New() string` factory
- [ ] Thread safety via `sync.Mutex` wrapping a single `ulid.MonotonicEntropy` source
- [ ] `pkg/ulidutil/ulid_test.go` — unit + race-detector tests

### Out of scope

- Parsing / validation of existing ULID strings (not needed in Phase 1)
- Batch generation helpers (deferred until required)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                       | Purpose                              |
| ------ | -------------------------- | ------------------------------------ |
| CREATE | `pkg/ulidutil/ulid.go`     | Thread-safe monotonic ULID generator |
| CREATE | `pkg/ulidutil/ulid_test.go`| Unit tests + race detector coverage  |

### Key interfaces / types

```go
// pkg/ulidutil/ulid.go
package ulidutil

import (
    "crypto/rand"
    "sync"
    "time"

    "github.com/oklog/ulid/v2"
)

var (
    mu      sync.Mutex
    entropy = ulid.Monotonic(rand.Reader, 0)
)

// New returns a new ULID string. Safe for concurrent use.
func New() string {
    mu.Lock()
    id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
    mu.Unlock()
    return id.String()
}
```

### SQL queries (sqlc)

N/A — this is a utility package with no database interaction.

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario                              | Handling                                                              |
| ------------------------------------- | --------------------------------------------------------------------- |
| `ulid.MustNew` panics (entropy error) | Acceptable — `crypto/rand` failure is an unrecoverable OS-level error |

---

## 5. Acceptance Criteria

- [ ] `New()` returns a 26-character uppercase string.
- [ ] Two consecutive calls within the same millisecond produce strictly increasing strings.
- [ ] `go test -race ./pkg/ulidutil/...` passes with zero data-race warnings.
- [ ] Test coverage for `pkg/ulidutil` = 100%.
- [ ] `golangci-lint run ./pkg/ulidutil/...` passes with zero issues.
- [ ] `gosec ./pkg/ulidutil/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.1 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                           | Type     | Status      |
| ------------------------------------ | -------- | ----------- |
| `github.com/oklog/ulid/v2` in go.mod | External | 🔵 backlog  |
| Phase 0 complete (module scaffolded) | Upstream | ✅ done     |

---

## 7. Testing Plan

### Unit tests (`pkg/ulidutil/ulid_test.go`, no build tag)

- **Happy path:** `New()` returns a non-empty 26-character string.
- **Uniqueness:** 10 000 consecutive calls produce 10 000 distinct values.
- **Monotonic ordering:** slice of 1 000 ULIDs generated in sequence is sorted ascending.
- **Concurrency:** 50 goroutines each calling `New()` 200 times — must produce 10 000 distinct values with `-race` enabled.

### Integration tests

N/A — no database interaction.

---

## 8. Open Questions

| # | Question                                       | Owner | Resolution |
| - | ---------------------------------------------- | ----- | ---------- |
| 1 | Should `New()` return `string` or `ulid.ULID`? | —     | Return `string` — callers never need the binary type; avoids leaking the dependency. |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.1 |
