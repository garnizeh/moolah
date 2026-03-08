# Task 1.5.11 — Wire `Idempotency` middleware on all mutating `POST` routes

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Apply the Redis-backed `Idempotency` middleware (Task 1.1.14) to all mutating `POST` routes in `cmd/api/routes.go`. This guarantees safe request replay: duplicate submissions with the same `Idempotency-Key` header return the cached response instead of creating duplicate records.

---

## 2. Context & Motivation

The idempotency middleware (`platform/middleware/idempotency.go`) and its Redis store (`platform/idempotency/redis_store.go`) are already implemented (Tasks 1.1.14–1.1.15) but are not yet wired into any routes. Mobile and web clients that retry failed requests without deduplication would create duplicate transactions, accounts, or categories. This task closes that gap before Phase 1 ships. See roadmap item 1.5.11.

---

## 3. Scope

### In scope

- [ ] Apply `IdempotencyMiddleware` to the following routes in `cmd/api/routes.go`:
  - `POST /v1/auth/otp/request`
  - `POST /v1/auth/otp/verify`
  - `POST /v1/accounts`
  - `POST /v1/categories`
  - `POST /v1/transactions`
  - `POST /v1/tenants/me/invite`
  - `PATCH /v1/accounts/{id}`
  - `PATCH /v1/categories/{id}`
  - `PATCH /v1/transactions/{id}`
  - `PATCH /v1/tenants/me`
- [ ] Verify the `Idempotency-Key` header is documented in Swaggo annotations (Task 1.5.10).

### Out of scope

- `DELETE` routes — idempotency for deletes is handled by soft-delete semantics (re-deleting a deleted record returns `404`, which is the correct idempotent response).
- Read (`GET`) routes — inherently idempotent.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                  | Purpose                                          |
| ------ | --------------------- | ------------------------------------------------ |
| MODIFY | `cmd/api/routes.go`   | Wrap mutating handlers with IdempotencyMiddleware|

### Middleware application pattern

```go
// Apply after RequireAuth so that userID is available in context for key scoping.
mux.Handle("POST /v1/accounts",
    middleware.RequireAuth(
        middleware.Idempotency(idempotencyStore)(
            accountHandler.Create,
        ),
    ),
)
```

The idempotency key is scoped per `userID + Idempotency-Key` header to prevent cross-user replay.

---

## 5. Acceptance Criteria

- [ ] All listed `POST` and `PATCH` routes use the `IdempotencyMiddleware`.
- [ ] Middleware is applied **after** `RequireAuth` in the chain so `userID` is in context.
- [ ] Duplicate requests with the same `Idempotency-Key` return the cached response.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                  | Type     | Status     |
| ------------------------------------------- | -------- | ---------- |
| Task 1.1.14 — Idempotency middleware        | Upstream | ✅ done    |
| Task 1.1.15 — Redis idempotency store       | Upstream | ✅ done    |
| Task 1.5.3 — `cmd/api/routes.go`            | Upstream | 🔵 backlog |
| Task 1.5.4–1.5.9 — Handler implementations  | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests

- Add a test to `cmd/api/routes_test.go` that verifies a duplicate `POST` request with the same `Idempotency-Key` is short-circuited by the middleware (mock store returns cached response).

### Integration / smoke test

- Task 1.6.4 collection sends a duplicate transaction creation and asserts the second response matches the first (status + body).

---

## 8. Open Questions

| # | Question                                                     | Owner | Resolution |
| - | ------------------------------------------------------------ | ----- | ---------- |
| 1 | Should `PATCH` routes require `Idempotency-Key` or make it optional? | — | Optional but honoured when present — matches industry standard (Stripe). |
| 2 | TTL for idempotency keys?                                    | —     | 24 hours, as set in Task 1.1.14. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
