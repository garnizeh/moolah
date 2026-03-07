# Task 1.1.11 — platform/middleware/logger.go: Request Logging Middleware

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement an `http.Handler` middleware that emits a structured `slog` log line for every HTTP request. The log entry includes method, path, HTTP status code, latency, `tenant_id`, and `user_id`. It uses the context values injected by `RequireAuth` (task 1.1.9) for the identity fields — missing values are logged as empty strings rather than causing an error.

---

## 2. Context & Motivation

Request-level observability is a production requirement. A structured log entry per request enables correlation across services, SLA monitoring, and debugging without relying on APM agents in Phase 1. The middleware must wrap the `ResponseWriter` to capture the status code written by the handler, since `net/http` does not expose this after the fact.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.11
- Depends on: task 1.1.2 (`pkg/logger`), task 1.1.9 (`middleware/auth.go` — context helpers)
- Applied by: task 1.5.2 (`cmd/api/server.go`) — outermost middleware in the chain

---

## 3. Scope

### In scope

- [x] `internal/platform/middleware/logger.go` — `RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler`
- [x] `responseWriter` wrapper that captures the HTTP status code
- [x] Log fields: `method`, `path`, `status`, `latency_ms`, `tenant_id`, `user_id`, `request_id`
- [x] `internal/platform/middleware/logger_test.go` — unit tests

### Out of scope

- Body/payload logging (security risk — never log request bodies in Phase 1)
- Distributed trace propagation (deferred to task 5.3 OpenTelemetry)
- Log sampling (deferred to Phase 5)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                | Purpose                                     |
| ------ | --------------------------------------------------- | ------------------------------------------- |
| CREATE | `internal/platform/middleware/logger.go`            | Structured request-logging middleware       |
| CREATE | `internal/platform/middleware/logger_test.go`       | Unit tests                                  |

### Key interfaces / types

```go
// internal/platform/middleware/logger.go
package middleware

import (
    "log/slog"
    "net/http"
    "time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
    http.ResponseWriter
    status      int
    wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
    if !rw.wroteHeader {
        rw.status = code
        rw.wroteHeader = true
        rw.ResponseWriter.WriteHeader(code)
    }
}

// RequestLogger returns a middleware that logs each request as a structured slog entry.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

            next.ServeHTTP(rw, r)

            tenantID, _ := TenantIDFromCtx(r.Context())
            userID, _   := UserIDFromCtx(r.Context())

            logger.InfoContext(r.Context(), "request",
                slog.String("method",     r.Method),
                slog.String("path",       r.URL.Path),
                slog.Int("status",        rw.status),
                slog.Int64("latency_ms",  time.Since(start).Milliseconds()),
                slog.String("tenant_id",  tenantID),
                slog.String("user_id",    userID),
            )
        })
    }
}
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A — middleware, not a handler.

### Error cases to handle

| Scenario                                    | Behaviour                                  |
| ------------------------------------------- | ------------------------------------------ |
| `tenant_id` / `user_id` absent from context | Log empty string — not an error            |
| Handler panics                              | Panic propagates; log line not emitted (recovery is a separate middleware concern) |
| `WriteHeader` called multiple times         | `responseWriter` captures only the first call |

---

## 5. Acceptance Criteria

- [x] Every request produces exactly one structured log line containing `method`, `path`, `status`, `latency_ms`.
- [x] `tenant_id` and `user_id` are populated when `RequireAuth` ran before the logger.
- [x] `tenant_id` and `user_id` are empty strings (not omitted) when the route is unauthenticated.
- [x] `status` reflects the actual HTTP status code written by the handler, not a hard-coded default.
- [x] `responseWriter.WriteHeader` deduplication prevents doubled status captures.
- [x] Test coverage for `logger.go` = 100%.
- [x] `golangci-lint run ./internal/platform/middleware/...` passes with zero issues.
- [x] `gosec ./internal/platform/middleware/...` passes with zero issues.
- [x] `docs/ROADMAP.md` row 1.1.11 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                       | Type     | Status     |
| ------------------------------------------------ | -------- | ---------- |
| Task 1.1.2 `pkg/logger` — `*slog.Logger`        | Upstream | ✅ done |
| Task 1.1.9 `middleware/auth.go` — context helpers | Upstream | ✅ done |
| Go 1.21+ (`log/slog` stdlib)                    | Runtime  | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`internal/platform/middleware/logger_test.go`, no build tag)

Use `httptest.NewRecorder`, `httptest.NewRequest`, and a `slog.Logger` writing to a `bytes.Buffer`.

- **Status captured correctly:** handler writes `204`; assert log entry `status=204`.
- **Default 200:** handler does not call `WriteHeader`; assert log entry `status=200`.
- **Latency positive:** assert `latency_ms >= 0`.
- **Auth context present:** inject `tenantID` and `userID` into context before calling middleware; assert log fields populated.
- **Auth context absent:** no context values; assert `tenant_id=""` and `user_id=""` in log.
- **Log output is valid JSON:** parse buffer as JSON; assert `method`, `path`, `status` keys present.

### Integration tests

N/A

---

## 8. Open Questions

| # | Question                                                    | Owner | Resolution |
| - | ----------------------------------------------------------- | ----- | ---------- |
| 1 | Should we add a `request_id` (UUID/ULID) for correlation?   | —     | Yes — generate one per request using `pkg/ulidutil.New()` and add to both the log line and a `X-Request-ID` response header. |

---

## 9. Change Log

| Date       | Author | Change                         |
| ---------- | ------ | ------------------------------ |
| 2026-03-07 | —      | Task created from roadmap 1.1.11 |
| 2026-03-07 | GitHub Copilot | Implementation of Request Logger with context and Request ID |
