# Task 1.6.4 — Full Phase 1 API Smoke Test

> **Roadmap Ref:** Phase 1 — MVP › 1.6 Quality Gate
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-11
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Execute a full end-to-end smoke test of the Phase 1 API against a live running server (using `testcontainers-go` or `docker-compose` in CI) that validates the complete happy-path user journey: OTP auth → tenant management → account CRUD → category CRUD → transaction CRUD. The smoke test must run as a dedicated CI job that succeeds only when the entire workflow completes without errors.

---

## 2. Context & Motivation

Unit and integration tests cover individual layers in isolation. A smoke test validates that the assembled system — DI wiring, middleware chain, handler → service → repository → DB — works as a whole. This is the final quality gate before Phase 1 is declared shippable. Without it, integration bugs at layer boundaries (e.g., JWT context propagation, idempotency middleware, error mapping) can slip through. See roadmap item 1.6.4.

---

## 3. Scope

### In scope

- [x] A Go `//go:build integration` test in `internal/server/smoke_test.go` (or `cmd/api/smoke_test.go`) that:
  - Spins up PostgreSQL + Redis via `testcontainers-go`.
  - Starts the full HTTP server in-process using `httptest.NewServer`.
  - Walks through the complete Phase 1 happy-path journey (see §4).
- [x] A `make smoke-test` target that runs only the smoke test.
- [x] A `smoke-test` CI job in `.github/workflows/ci.yml` that runs after `integration-tests`.
- [x] Idempotency key usage verified: sending the same mutation request twice yields the same response.
- [x] Auth token rotation verified: `RefreshToken` returns a new valid token.

### Out of scope

- Negative / error-path smoke tests (covered by unit tests in each handler).
- Load or performance testing.
- Frontend / WebSocket end-to-end testing.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                              | Purpose                                               |
| ------ | --------------------------------- | ----------------------------------------------------- |
| CREATE | `internal/server/smoke_test.go`   | Full journey smoke test with testcontainers           |
| MODIFY | `Makefile`                        | Add `smoke-test` target                               |
| MODIFY | `.github/workflows/ci.yml`        | Add `smoke-test` job after `integration-tests`        |

### Happy-path journey (ordered steps)

| Step | Method + Path                         | Assertion                                          |
| ---- | ------------------------------------- | -------------------------------------------------- |
| 1    | `GET /healthz`                        | 200 OK                                             |
| 2    | `POST /v1/auth/otp/request`           | 202 Accepted; OTP saved in DB                      |
| 3    | `POST /v1/auth/otp/verify`            | 200 OK; `token` + `refresh_token` in body          |
| 4    | `GET /v1/tenants/me`                  | 200 OK; tenant matches seeded data                 |
| 5    | `PATCH /v1/tenants/me`                | 200 OK; tenant name updated                        |
| 6    | `POST /v1/accounts`                   | 201 Created; `id` in body                          |
| 7    | `GET /v1/accounts/{id}`               | 200 OK; matches created account                    |
| 8    | `PATCH /v1/accounts/{id}`             | 200 OK; account name updated                       |
| 9    | `GET /v1/accounts`                    | 200 OK; list contains the created account          |
| 10   | `POST /v1/categories`                 | 201 Created; `id` in body                          |
| 11   | `GET /v1/categories/{id}`             | 200 OK; matches created category                   |
| 12   | `POST /v1/transactions`               | 201 Created; `id` in body                          |
| 13   | `GET /v1/transactions/{id}`           | 200 OK; matches created transaction                |
| 14   | `GET /v1/transactions`                | 200 OK; list contains the created transaction      |
| 15   | `DELETE /v1/transactions/{id}`        | 204 No Content                                     |
| 16   | `DELETE /v1/accounts/{id}`            | 204 No Content                                     |
| 17   | `POST /v1/auth/token/refresh`         | 200 OK; new token returned                         |
| 18   | Replay step 6 with same `Idempotency-Key` | 201 same body (idempotency cache hit)          |

### Smoke test structure

```go
//go:build integration

package server_test

func TestSmoke_Phase1HappyPath(t *testing.T) {
    t.Parallel()

    // 1. Start containers
    pgContainer  := containers.NewPostgresContainer(t)
    redisContainer := containers.NewRedisContainer(t)

    // 2. Build server
    cfg := buildTestConfig(pgContainer, redisContainer)
    srv := buildServer(t, cfg)
    ts  := httptest.NewServer(srv)
    t.Cleanup(ts.Close)

    client := ts.Client()
    base   := ts.URL

    // 3. Walk through journey steps
    t.Run("healthz", func(t *testing.T) { ... })
    t.Run("auth", func(t *testing.T) { ... })
    // ...
}
```

### Makefile target

```makefile
## smoke-test: Run Phase 1 end-to-end smoke test (requires Docker)
smoke-test:
  @echo "🚀 Running Phase 1 smoke test..."
  @$(GO) test -v -race -count=1 -tags=integration -timeout=300s \
    -run TestSmoke_Phase1HappyPath \
    ./internal/server/...
```

### CI job

```yaml
smoke-test:
  name: Smoke Test
  runs-on: ubuntu-latest
  needs: integration-tests
  steps:
    - uses: actions/checkout@v6
    - uses: actions/setup-go@v6
      with:
        go-version-file: 'go.mod'
        cache: true
    - name: Run Phase 1 smoke test
      run: make smoke-test
      env:
        TESTCONTAINERS_RYUK_DISABLED: "false"
```

---

## 5. Acceptance Criteria

- [x] `TestSmoke_Phase1HappyPath` passes end-to-end with a live PostgreSQL + Redis.
- [x] All 18 journey steps are covered and individually asserted.
- [x] Idempotency replay (step 18) is validated.
- [x] `make smoke-test` exits 0 locally (with Docker running).
- [x] CI `smoke-test` job passes on `main` branch.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status     |
| ---------------------------------------------- | -------- | ---------- |
| Task 1.1.16 — `testutil/containers`            | Upstream | ✅ done    |
| Task 1.5.1 — `cmd/api/main.go` DI wiring       | Upstream | ✅ done    |
| Task 1.5.3 — Routes registered                 | Upstream | ✅ done    |
| Task 1.5.4–1.5.9 — All handlers implemented    | Upstream | ✅ done    |
| Task 1.5.11 — Idempotency middleware wired      | Upstream | ✅ done    |
| Docker available on CI runner                   | External | ✅ done    |

---

## 7. Testing Plan

### Smoke test vs. integration tests

| Aspect              | Integration Tests              | Smoke Test                          |
| ------------------- | ------------------------------ | ----------------------------------- |
| Layer tested        | Repository (DB only)           | Full stack (HTTP → DB)              |
| Server running      | No                             | Yes (`httptest.NewServer`)          |
| Scope               | Per-package, isolated          | Cross-layer journey                 |
| Run time            | Fast (~10–30s)                 | Slower (~60–120s)                   |

### Failure modes to guard against

- Missing DI wiring (e.g., handler constructed without required service).
- JWT context not propagated through middleware chain.
- Idempotency middleware not applied to the correct routes.
- Database migration not auto-applied on server start.

---

## 8. Open Questions

| # | Question                                                           | Owner | Resolution                                                       |
| - | ------------------------------------------------------------------ | ----- | ---------------------------------------------------------------- |
| 1 | Should the smoke test also cover the admin (`/v1/admin/*`) routes? | —     | Add as a second sub-journey (`TestSmoke_AdminJourney`) in the same file. |
| 2 | Should we run the smoke test on every PR or only on `main` pushes? | —     | Run on every PR to catch regressions early; it runs after integration tests to keep fast-feedback PRs from being gated unnecessarily long. |
| 3 | Where should OTP be intercepted in tests (no live SMTP)?           | —     | Read the OTP directly from the `otp_requests` DB table or inject a mock mailer that records the sent code. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-10 | Copilot| Task created from roadmap |
| 2026-03-11 | Copilot| Implemented smoke test, CI, Makefile and fixed handler wiring |
