# Task 2.13 — Smoke Test: Phase 2 Happy Path (`internal/server/smoke_test.go`)

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Quality Gate
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Extend `internal/server/smoke_test.go` with `TestSmoke_Phase2HappyPath` — an end-to-end journey that exercises all Phase 2 routes against real Postgres and Redis containers. The test mirrors the Phase 1 pattern (`TestSmoke_Phase1HappyPath`): wire the full DI graph, spin up real containers, and walk through every Phase 2 endpoint in a logical sequence — from creating a credit card account and registering a master purchase to triggering invoice closing and asserting the materialised transaction.

---

## 2. Context & Motivation

`TestSmoke_Phase1HappyPath` (Task 1.6.4) established the pattern: a linear, numbered journey that validates every route is correctly wired, middleware is applied, and the integration contract between layers holds. Phase 2 introduces `MasterPurchaseHandler` and the `CloseInvoice` endpoint. Without a smoke test, misconfigurations in route registration, DI wiring, or middleware ordering can silently ship to production. This test is the final quality gate for Phase 2.

---

## 3. Scope

### In scope

- [ ] New test function `TestSmoke_Phase2HappyPath` in `internal/server/smoke_test.go`.
- [ ] DI wiring for Phase 2 dependencies: `MasterPurchaseRepository`, `MasterPurchaseService`, `InvoiceCloser`.
- [ ] Update `server.New(...)` call (or equivalent builder) to accept Phase 2 services.
- [ ] Journey steps covering all 7 Phase 2 endpoints (see §4).
- [ ] Assertion that `close-invoice` materialises a transaction in the DB.
- [ ] Assertion that the audit log contains a `SYSTEM`-actor entry after closing.
- [ ] Assertion that `projected_schedule` is returned in the `POST /v1/master-purchases` response.
- [ ] Idempotency replay verification for `POST /v1/master-purchases`.

### Out of scope

- Phase 1 smoke test modifications (must remain green and untouched).
- Scheduler/ticker testing (Phase 5).
- Error-path smoke steps (happy path only — error coverage lives in handler unit tests).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                      | Purpose                                         |
| ------ | ----------------------------------------- | ----------------------------------------------- |
| MODIFY | `internal/server/smoke_test.go`           | Add `TestSmoke_Phase2HappyPath` journey         |
| MODIFY | `internal/server/routes.go`               | Verify Phase 2 routes are registered (pre-req)  |
| MODIFY | `internal/server/server.go`               | Accept Phase 2 services in constructor if needed |

### Journey steps

| Step | Method | Path                                            | Expected Status | Notes                                          |
| ---- | ------ | ----------------------------------------------- | --------------- | ---------------------------------------------- |
| 00   | GET    | `/healthz`                                      | 200             | Sanity check                                   |
| 01–02| POST   | `/v1/auth/otp/request` + `/v1/auth/otp/verify`  | 202 + 200       | Reuse Phase 1 OTP pattern; get `accessToken`   |
| 03   | POST   | `/v1/accounts`                                  | 201             | Create `credit_card` account; capture `ccID`   |
| 04   | POST   | `/v1/categories`                                | 201             | Create expense category; capture `categoryID`  |
| 05   | POST   | `/v1/master-purchases`                          | 201             | Body: 3-instalment, R$12.00 total; capture `mpID`; assert `projected_schedule` has 3 entries |
| 05b  | POST   | `/v1/master-purchases` (replay)                 | 201             | Same idempotency key; assert `X-Cache: HIT` and same `mpID` |
| 06   | GET    | `/v1/master-purchases`                          | 200             | List endpoint returns non-empty slice          |
| 07   | GET    | `/v1/master-purchases/{mpID}`                   | 200             | Returns `projected_schedule` in response body  |
| 08   | GET    | `/v1/accounts/{ccID}/master-purchases`          | 200             | Account-scoped list; asserts `mpID` present    |
| 09   | PATCH  | `/v1/master-purchases/{mpID}`                   | 200             | Update description; assert new description     |
| 10   | POST   | `/v1/accounts/{ccID}/close-invoice`             | 200             | `processed_count == 1`; `errors` empty         |
| 11   | GET    | `/v1/transactions`                              | 200             | List contains materialised instalment tx       |
| 12   | DELETE | `/v1/accounts/{ccID}`                           | 409 / 422       | Blocked by open master purchases OR 204 if policies allow; document decision |

> **Note on step 12:** If the business rule is "cannot delete an account with open master purchases", document this as an explicit 422 assertion. Otherwise, soft-delete succeeds (204).

### DI additions (reference)

```go
// In TestSmoke_Phase2HappyPath:
mpRepo     := repository.NewMasterPurchaseRepository(pgDB.Queries)
mpSvc      := service.NewMasterPurchaseService(mpRepo, accountRepo)
invoiceSvc := service.NewInvoiceCloser(mpRepo, transactionRepo, auditRepo, mpSvc, pgDB.Pool)

srv := server.New(
    "0",
    authSvc,
    tenantSvc,
    accountSvc,
    categorySvc,
    transactionSvc,
    adminSvc,
    mpSvc,         // ← Phase 2
    invoiceSvc,    // ← Phase 2
    idempotencyStore,
    rateLimiterStore,
    tokenParser,
)
```

### Key assertions

```go
// Step 05 — projected_schedule returned in creation response
var createResp struct {
    MasterPurchase    domain.MasterPurchase         `json:"master_purchase"`
    ProjectedSchedule []domain.ProjectedInstallment `json:"projected_schedule"`
}
decodeJSON(t, resp, &createResp)
require.Len(t, createResp.ProjectedSchedule, 3, "expected 3 projected instalments")

// Step 10 — close-invoice processes 1 master purchase
var closeResp struct {
    ProcessedCount int      `json:"processed_count"`
    Errors         []string `json:"errors"`
}
decodeJSON(t, resp, &closeResp)
assert.Equal(t, 1, closeResp.ProcessedCount)
assert.Empty(t, closeResp.Errors)

// Step 11 — materialised transaction visible in list
var txList struct {
    Data []domain.Transaction `json:"data"`
}
decodeJSON(t, resp, &txList)
var found bool
for _, tx := range txList.Data {
    if tx.MasterPurchaseID == mpID {
        found = true
        assert.Equal(t, int64(400), tx.AmountCents) // 1200 / 3
    }
}
assert.True(t, found, "materialised instalment transaction must appear in list")
```

---

## 5. Acceptance Criteria

- [ ] `TestSmoke_Phase2HappyPath` compiles and passes with `go test -tags integration ./internal/server/...`.
- [ ] All 12+ journey steps are present and numbered with `t.Run("NN_...", ...)`.
- [ ] `t.Parallel()` called at top of test function.
- [ ] `TestSmoke_Phase1HappyPath` continues to pass unchanged.
- [ ] `POST /v1/master-purchases` response includes `projected_schedule` with correct count.
- [ ] Idempotency replay on `POST /v1/master-purchases` asserts `X-Cache: HIT` and same ID.
- [ ] `POST /v1/accounts/{id}/close-invoice` returns `processed_count == 1` and empty `errors`.
- [ ] Materialised transaction with correct `amount_cents` and `master_purchase_id` appears in `/v1/transactions`.
- [ ] Test uses `require.NoError` for every fallible call (including `resp.Body.Close()`).
- [ ] No `_ = ...` ignoring errors.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.13 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                           | Type     | Status       |
| ---------------------------------------------------- | -------- | ------------ |
| Task 2.6 — `MasterPurchaseHandler` implemented       | Upstream | 🔵 backlog   |
| Task 2.7 — `InvoiceCloser` service                   | Upstream | 🔵 backlog   |
| Task 2.8 — `CloseInvoice` endpoint                   | Upstream | 🔵 backlog   |
| Task 2.10 — `domain.ActorSystem` constant            | Upstream | 🔵 backlog   |
| Phase 2 routes registered in `internal/server/routes.go` | Upstream | 🔵 backlog |
| `internal/testutil/containers` (Phase 1)             | Upstream | ✅ done      |
| `TestSmoke_Phase1HappyPath` pattern (reference)      | Upstream | ✅ done      |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — this is an integration smoke test by nature.

### Integration tests (`//go:build integration`)

- **File:** `internal/server/smoke_test.go` (extended)
- **Build tag:** `integration`
- **Container:** PostgreSQL + Redis via `testcontainers-go` (shared `containers.NewPostgresDB` + `containers.NewRedisClient`).
- **Run:** `make integration` or `go test -tags integration -v -run TestSmoke_Phase2 ./internal/server/...`

---

## 8. Open Questions

| # | Question                                                                                     | Owner | Resolution |
| - | -------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `server.New(...)` accept Phase 2 services as additional parameters, or should a builder pattern be introduced? | — | Extend the existing variadic or struct-based constructor — follow the Phase 1 precedent. |
| 2 | Step 12: should deleting a credit card account with open master purchases be blocked (422) or allowed (204)? | — | Decide during Task 2.6 implementation; document the outcome here. |
| 3 | Should the journey also verify the Phase 1 smoke test can still run against the Phase 2 schema (migration compatibility)? | — | Yes — CI runs both smoke tests in the same pipeline; no extra step needed. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
