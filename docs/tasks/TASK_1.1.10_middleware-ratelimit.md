# Task 1.1.10 — platform/middleware/ratelimit.go: Token-Bucket OTP Rate Limiter

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement an `http.Handler` middleware that applies a per-email token-bucket rate limit to the OTP request endpoint (`POST /auth/otp/request`). The limit is 5 requests per 15-minute window. Exceeding the limit returns `429 Too Many Requests` with a `Retry-After` header. Limiters are kept in an in-memory map with a background cleanup goroutine to prevent unbounded growth.

---

## 2. Context & Motivation

Without rate limiting, the OTP endpoint is vulnerable to email bombing (flooding a victim's inbox) and credential-stuffing enumeration. The `golang.org/x/time/rate` token-bucket limiter is the recommended approach for in-process rate limiting. Per-email limiting (rather than per-IP) is preferred because multiple users may share an IP (NAT, office network) but a single user should never need more than 5 OTP codes in 15 minutes.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.10
- Security reference: `docs/ARCHITECTURE.md` — Authentication section
- Applied to: `POST /auth/otp/request` route in task 1.5.3

---

## 3. Scope

### In scope

- [ ] `internal/platform/middleware/ratelimit.go` — `OTPRateLimiter(next http.Handler) http.Handler`
- [ ] Per-email limiter map, protected by `sync.RWMutex`
- [ ] Background goroutine that prunes stale limiters every 15 minutes
- [ ] `Retry-After` response header (seconds until next token is available)
- [ ] `internal/platform/middleware/ratelimit_test.go` — unit tests

### Out of scope

- Distributed rate limiting via Redis (deferred — in-process is sufficient for Phase 1 single-instance deployment)
- Per-IP limiting (complements email limiting but deferred)
- Admin override / allowlist (deferred to Phase 4)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                | Purpose                                     |
| ------ | --------------------------------------------------- | ------------------------------------------- |
| CREATE | `internal/platform/middleware/ratelimit.go`         | Token-bucket rate limiter middleware        |
| CREATE | `internal/platform/middleware/ratelimit_test.go`    | Unit tests                                  |

### Key interfaces / types

```go
// internal/platform/middleware/ratelimit.go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

const (
    otpRateLimit  = 5               // max requests
    otpRatePeriod = 15 * time.Minute
)

type emailLimiter struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

// OTPRateLimiter returns a middleware that enforces per-email OTP rate limiting.
// It reads the email from the parsed JSON body; falls through (no limit) if email
// cannot be extracted — the handler itself will validate the request.
func OTPRateLimiter() func(http.Handler) http.Handler { ... }
```

> **Note on body reading:** The middleware must read and buffer the request body to extract the email, then restore it for the downstream handler using `io.NopCloser(bytes.NewReader(body))`.

### HTTP response for rate-limited requests

```json
{ "error": { "code": "RATE_LIMITED", "message": "Too many OTP requests. Please wait before retrying." } }
```

Headers: `Retry-After: <seconds>`, `Content-Type: application/json`

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A — middleware applied to `POST /auth/otp/request`.

### Error cases to handle

| Scenario                            | HTTP Status | Behaviour                              |
| ----------------------------------- | ----------- | -------------------------------------- |
| Limit exceeded                      | 429         | Return JSON error + `Retry-After`      |
| Email not in request body           | Pass through | Let the handler return the validation error |
| Body read error                     | Pass through | Let the handler fail gracefully        |

---

## 5. Acceptance Criteria

- [ ] 5 requests in < 15 minutes from the same email: all pass.
- [ ] 6th request from the same email within the window: returns `429`.
- [ ] `Retry-After` header is present and is a positive integer (seconds).
- [ ] After the window expires (time.Sleep or mocked clock), the 6th request passes.
- [ ] Different emails are limited independently.
- [ ] Background cleanup removes entries unseen for > 15 minutes.
- [ ] Test coverage for `ratelimit.go` ≥ 90%.
- [ ] `golangci-lint run ./internal/platform/middleware/...` passes with zero issues.
- [ ] `gosec ./internal/platform/middleware/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.10 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                      | Type     | Status     |
| ----------------------------------------------- | -------- | ---------- |
| `golang.org/x/time/rate` added to `go.mod`      | External | 🔵 backlog |
| Phase 0 complete (module scaffolded)            | Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`internal/platform/middleware/ratelimit_test.go`, no build tag)

Use `httptest.NewRecorder` and `httptest.NewRequest` with JSON bodies.

- **Under limit:** send 5 requests with `email: "a@b.com"`; assert all return 200 (from stub next handler).
- **Over limit:** send 6th request; assert 429 + `Retry-After` header present.
- **Different emails:** `a@b.com` at limit; `c@d.com` sends first request; assert 200.
- **Empty body:** send request with no body; assert request passes through to handler.
- **Cleanup test:** advance the internal clock (inject time source); verify stale entries are removed.

### Integration tests

N/A — fully testable in-process.

---

## 8. Open Questions

| # | Question                                                             | Owner | Resolution |
| - | -------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should the rate limit parameters (5 req / 15 min) be configurable via `pkg/config`? | — | Yes — accept as constructor parameters; defaults match spec. |
| 2 | Read body to extract email, or require callers to pass email via a custom header? | — | Read body and restore with `io.NopCloser` — cleaner API contract for the OTP endpoint. |

---

## 9. Change Log

| Date       | Author | Change                         |
| ---------- | ------ | ------------------------------ |
| 2026-03-07 | —      | Task created from roadmap 1.1.10 |
