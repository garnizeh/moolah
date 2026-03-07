# Task 1.1.9 — platform/middleware/auth.go: PASETO Validation + Context Injection

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Implement two `net/http` middleware functions: `RequireAuth` validates the PASETO token from the `Authorization: Bearer <token>` header and injects the parsed claims into the request context; `RequireRole` is a higher-order middleware that enforces a minimum role level, returning `403` if the authenticated user's role is insufficient.

---

## 2. Context & Motivation

Every protected endpoint in the API needs to verify the caller's identity and tenant before any business logic runs. Middleware is the correct boundary for this cross-cutting concern — it decouples auth from handlers and ensures tenant_id is always available in context for repository calls. Downstream packages (`service`, `handler`) must never call `pkg/pasetoutils` directly.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.9
- Depends on: task 1.1.4 (`pkg/pasetoutils`), task 1.2.2 (`domain/role.go`)
- Consumed by: task 1.5.2 (`cmd/api/server.go`) — middleware chain

---

## 3. Scope

### In scope

- [ ] `internal/platform/middleware/auth.go` — `RequireAuth` and `RequireRole`
- [ ] Context key types (unexported) for `tenantID`, `userID`, `role`
- [ ] Exported helper functions: `TenantIDFromCtx`, `UserIDFromCtx`, `RoleFromCtx`
- [ ] `internal/platform/middleware/auth_test.go` — unit tests with mock token scenarios

### Out of scope

- Token refresh (handled in `service/auth_service.go`)
- Session storage / revocation (deferred to Phase 5)
- API key authentication (not in scope)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                          | Purpose                                        |
| ------ | --------------------------------------------- | ---------------------------------------------- |
| CREATE | `internal/platform/middleware/auth.go`        | `RequireAuth`, `RequireRole`, context helpers  |
| CREATE | `internal/platform/middleware/auth_test.go`   | Table-driven unit tests                        |

### Key interfaces / types

```go
// internal/platform/middleware/auth.go
package middleware

import (
    "context"
    "net/http"

    "github.com/garnizeh/moolah/pkg/pasetoutils"
    "github.com/garnizeh/moolah/internal/domain"
)

// Unexported context key types — prevent collisions with other packages.
type contextKey int

const (
    tenantIDKey contextKey = iota
    userIDKey
    roleKey
)

// RequireAuth validates the Bearer token and injects claims into ctx.
// Returns 401 if the header is missing or the token is invalid/expired.
func RequireAuth(tokenParser func(string) (*pasetoutils.Claims, error)) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // extract Bearer token, call tokenParser, inject into ctx, call next
        })
    }
}

// RequireRole returns a middleware that enforces the caller has at least `role`.
// Must be chained after RequireAuth.
// Returns 403 if the role is insufficient.
func RequireRole(role domain.Role) func(http.Handler) http.Handler { ... }

// TenantIDFromCtx extracts the tenant ID injected by RequireAuth.
// Returns ("", false) if not present.
func TenantIDFromCtx(ctx context.Context) (string, bool) { ... }

// UserIDFromCtx extracts the user ID injected by RequireAuth.
func UserIDFromCtx(ctx context.Context) (string, bool) { ... }

// RoleFromCtx extracts the role injected by RequireAuth.
func RoleFromCtx(ctx context.Context) (domain.Role, bool) { ... }
```

### HTTP response format for auth errors

```json
{ "error": { "code": "UNAUTHORIZED", "message": "missing or invalid token" } }
{ "error": { "code": "FORBIDDEN",    "message": "insufficient role" } }
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A — middleware, not a handler.

### Error cases to handle

| Scenario                             | HTTP Status | Error Code      |
| ------------------------------------ | ----------- | --------------- |
| Missing `Authorization` header       | 401         | `UNAUTHORIZED`  |
| Malformed `Bearer` prefix            | 401         | `UNAUTHORIZED`  |
| Token invalid (bad signature / key)  | 401         | `UNAUTHORIZED`  |
| Token expired                        | 401         | `TOKEN_EXPIRED` |
| Role insufficient                    | 403         | `FORBIDDEN`     |

---

## 5. Acceptance Criteria

- [ ] `RequireAuth` passes the request through when a valid, non-expired token is present.
- [ ] `RequireAuth` returns `401` with JSON body when token is missing, invalid, or expired.
- [ ] `TenantIDFromCtx` / `UserIDFromCtx` / `RoleFromCtx` return values injected by `RequireAuth`.
- [ ] `RequireRole(domain.RoleAdmin)` returns `403` for a `member`-role token.
- [ ] `RequireRole(domain.RoleMember)` passes for any authenticated user.
- [ ] Context helpers return `false` on a context that did not pass through `RequireAuth`.
- [ ] Test coverage for `internal/platform/middleware/auth.go` = 100%.
- [ ] `golangci-lint run ./internal/platform/middleware/...` passes with zero issues.
- [ ] `gosec ./internal/platform/middleware/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.9 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status     |
| ---------------------------------------------- | -------- | ---------- |
| Task 1.1.4 `pkg/pasetoutils` — `Parse` function | Upstream | 🔵 backlog |
| Task 1.2.2 `domain/role.go` — `Role` type      | Upstream | 🔵 backlog |
| Phase 0 complete (module scaffolded)           | Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`internal/platform/middleware/auth_test.go`, no build tag)

Use `httptest.NewRecorder` and `httptest.NewRequest`. Inject a stub `tokenParser` function — no real PASETO keys needed in unit tests.

- **Valid token:** stub returns valid claims; assert next handler called; assert context values set.
- **Missing header:** assert 401; assert next handler NOT called.
- **Malformed header (no `Bearer` prefix):** assert 401.
- **Expired token:** stub returns `pasetoutils.ErrTokenExpired`; assert 401 with `"TOKEN_EXPIRED"` code.
- **Invalid token:** stub returns `pasetoutils.ErrTokenInvalid`; assert 401 with `"UNAUTHORIZED"` code.
- **RequireRole — sufficient role:** `admin` token + `RequireRole(domain.RoleMember)` → passes.
- **RequireRole — insufficient role:** `member` token + `RequireRole(domain.RoleAdmin)` → 403.
- **Context helpers on empty ctx:** return zero values and `false`.

### Integration tests

N/A — middleware behavior is fully testable with `httptest`.

---

## 8. Open Questions

| # | Question                                                          | Owner | Resolution |
| - | ----------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `RequireAuth` accept the PASETO key directly or a parser function? | — | Accept a parser function (`func(string) (*Claims, error)`) — easier to mock in tests and decouples key management from middleware. |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.9 |
