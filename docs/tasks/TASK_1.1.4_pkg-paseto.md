# Task 1.1.4 — pkg/paseto: PASETO v4 Local Token Seal/Parse

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement a `pkg/paseto` package that wraps PASETO v4 local (symmetric) token creation and verification. The package exposes typed `Claims`, a `Seal` function (issues a token), and a `Parse` function (validates and extracts claims). Consumed by the auth service (issue tokens) and the auth middleware (validate tokens on every protected request).

---

## 2. Context & Motivation

The project uses PASETO v4 local tokens instead of JWT. PASETO v4 local uses `XChaCha20-Poly1305` AEAD — tamper-proof and encrypted by default, removing the JWT `alg:none` attack class entirely. The library `aidanwoods.dev/go-paseto` is the canonical Go implementation.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.4
- Consumed by: task 1.1.9 (`middleware/auth.go`), task 1.4.1 (`service/auth_service.go`)

---

## 3. Scope

### In scope

- [ ] `pkg/paseto/paseto.go` — `Claims` struct, `Seal`, `Parse`
- [ ] `pkg/paseto/paseto_test.go` — round-trip, expiry, tamper, wrong-key tests
- [ ] Claims carry: `TenantID`, `UserID`, `Role`, `IssuedAt`, `ExpiresAt`

### Out of scope

- PASETO v4 public (asymmetric) tokens — not needed for this use case
- Token refresh logic (handled in `service/auth_service.go`)
- Token revocation / denylist (deferred to Phase 5)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                              | Purpose                              |
| ------ | --------------------------------- | ------------------------------------ |
| CREATE | `pkg/paseto/paseto.go`       | Seal, Parse, Claims types            |
| CREATE | `pkg/paseto/paseto_test.go`  | Unit tests for all token scenarios   |

### Key interfaces / types

```go
// pkg/paseto/paseto.go
package paseto

import (
    "errors"
    "time"

    "aidanwoods.dev/go-paseto"
)

// ErrTokenExpired is returned when a valid token is past its expiry time.
var ErrTokenExpired = errors.New("paseto: token expired")

// ErrTokenInvalid is returned when the token cannot be decrypted or parsed.
var ErrTokenInvalid = errors.New("paseto: token invalid")

// Claims holds the application-specific payload embedded in the token.
type Claims struct {
    TenantID  string    `json:"tenant_id"`
    UserID    string    `json:"user_id"`
    Role      string    `json:"role"`
    IssuedAt  time.Time `json:"iat"`
    ExpiresAt time.Time `json:"exp"`
}

// Seal encrypts claims into a PASETO v4 local token string.
// key must be exactly 32 bytes (use paseto.NewV4SymmetricKey()).
func Seal(claims Claims, key paseto.V4SymmetricKey) (string, error) { ... }

// Parse decrypts and validates a token, returning the embedded Claims.
// Returns ErrTokenExpired or ErrTokenInvalid on failure.
func Parse(token string, key paseto.V4SymmetricKey) (*Claims, error) { ... }
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario                       | Sentinel Error        |
| ------------------------------ | --------------------- |
| Token past `ExpiresAt`         | `ErrTokenExpired`     |
| Ciphertext tampered / bad key  | `ErrTokenInvalid`     |
| Malformed token string         | `ErrTokenInvalid`     |

---

## 5. Acceptance Criteria

- [ ] `Seal` → `Parse` round-trip with the same key returns identical `Claims`.
- [ ] `Parse` returns `ErrTokenExpired` for a token with `ExpiresAt` in the past.
- [ ] `Parse` returns `ErrTokenInvalid` when decrypting with a different key.
- [ ] `Parse` returns `ErrTokenInvalid` for a randomly modified token string.
- [ ] `Claims` fields `TenantID`, `UserID`, `Role` survive the round-trip unchanged.
- [ ] Test coverage for `pkg/paseto` = 100%.
- [ ] `golangci-lint run ./pkg/paseto/...` passes with zero issues.
- [ ] `gosec ./pkg/paseto/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.4 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                    | Type     | Status     |
| --------------------------------------------- | -------- | ---------- |
| `aidanwoods.dev/go-paseto` added to `go.mod`  | External | 🔵 backlog |
| Phase 0 complete (module scaffolded)          | Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`pkg/paseto/paseto_test.go`, no build tag)

- **Round-trip:** seal with TTL=1h; parse immediately; assert all claims fields equal.
- **Expired:** seal with TTL=-1s (already expired); parse; assert `errors.Is(err, ErrTokenExpired)`.
- **Wrong key:** seal with key A; parse with key B; assert `errors.Is(err, ErrTokenInvalid)`.
- **Tampered ciphertext:** flip a byte in the middle of the token string; assert `ErrTokenInvalid`.
- **Empty token string:** assert `ErrTokenInvalid`.

### Integration tests

N/A

---

## 8. Open Questions

| # | Question                                              | Owner | Resolution |
| - | ----------------------------------------------------- | ----- | ---------- |
| 1 | Store the symmetric key as hex string in config or as raw bytes? | — | Hex string in `PASETO_SECRET_KEY` env var; decode to 32 bytes in `pkg/config`. |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.4 |
