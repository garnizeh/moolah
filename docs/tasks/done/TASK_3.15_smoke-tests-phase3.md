# Task 3.15 — Mock Factory Updates + Phase 3 Smoke Tests

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Quality & Finalization
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Extend the centralized mock factory (`internal/testutil/mocks`) with all new Phase 3 interfaces, add Phase 3 seed factories to `internal/testutil/seeds`, and write Phase 3 smoke tests in `internal/server/` that exercise the full investment stack end-to-end against a real Postgres container. Mirrors the pattern established by Phase 2 smoke tests (Task 2.13).

---

## 2. Context & Motivation

Smoke tests validate the entire stack (handler → service → repository → DB) with a single HTTP round-trip per scenario. They catch integration issues (e.g., missing migrations, wiring bugs, wrong JSON serialization) that unit tests cannot catch in isolation.

All Phase 1 and Phase 2 interfaces have testify/mock implementations in `internal/testutil/mocks/`. Phase 3 must add mocks for the six new repository interfaces and the `InvestmentService` interface so that unit tests in handler and service packages can compile.

**Reference:** Phase 2 smoke test pattern in `internal/server/smoke_test.go`.

---

## 3. Scope

### In scope

- [x] Mock types for all six new repository interfaces (generated or hand-written via testify/mock):
  - `MockAssetRepository`
  - `MockTenantAssetConfigRepository`
  - `MockPositionRepository`
  - `MockPositionSnapshotRepository`
  - `MockPositionIncomeEventRepository`
  - `MockPortfolioSnapshotRepository`
- [x] `MockInvestmentService` — mock for `domain.InvestmentService`.
- [x] Seed factories in `internal/testutil/seeds`:
  - `SeedAsset`, `SeedTenantAssetConfig`, `SeedPosition`, `SeedPositionIncomeEvent`.
- [x] Smoke tests in `internal/server/smoke_test.go` (build tag `integration`):
  - `TestSmoke_CreateAsset` — admin creates asset; tenant reads with COALESCE override.
  - `TestSmoke_PositionLifecycle` — create position → receive income → check summary.
  - `TestSmoke_PortfolioSnapshot` — trigger manual snapshot; verify persisted aggregate.
  - `TestSmoke_TenantAssetConfig` — upsert config; verify `GET /v1/assets/{id}` returns overridden values.

### Out of scope

- Bruno collection updates (separate documentation task).
- Swagger regeneration (handled during each handler task, not aggregated here).
- Load or performance tests.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                         | Purpose                                           |
| ------ | ------------------------------------------------------------ | ------------------------------------------------- |
| MODIFY | `internal/testutil/mocks/querier_mock.go`                    | Final consolidation of all Phase 3 mock methods   |
| CREATE | `internal/testutil/mocks/investment_service_mock.go`         | `MockInvestmentService`                           |
| CREATE | `internal/testutil/mocks/asset_repository_mock.go`           | `MockAssetRepository`                             |
| CREATE | `internal/testutil/mocks/tenant_asset_config_repository_mock.go` | `MockTenantAssetConfigRepository`            |
| CREATE | `internal/testutil/mocks/position_repository_mock.go`        | `MockPositionRepository`                          |
| CREATE | `internal/testutil/mocks/position_income_event_repository_mock.go` | `MockPositionIncomeEventRepository`         |
| CREATE | `internal/testutil/mocks/portfolio_snapshot_repository_mock.go`    | `MockPortfolioSnapshotRepository`           |
| MODIFY | `internal/testutil/seeds/seeds.go`                           | Add `SeedAsset`, `SeedPosition`, etc.             |
| MODIFY | `internal/server/smoke_test.go`                              | Add Phase 3 smoke tests                           |

### Smoke test pattern

```go
//go:build integration

func TestSmoke_PositionLifecycle(t *testing.T) {
    t.Parallel()
    srv := newTestServer(t)
    tenant, token := seeds.SeedTenantWithToken(t, srv.DB)

    // 1. Create asset (admin)
    adminToken := seeds.SeedAdminToken(t, srv.DB)
    asset := httpPost(t, srv, "/v1/assets", adminToken, CreateAssetRequest{
        Ticker: "TEST-01", Name: "Test Corp", AssetType: "stock", Currency: "BRL",
    })

    // 2. Create position
    pos := httpPost(t, srv, "/v1/positions", token, CreatePositionRequest{
        AssetID: asset.ID, AccountID: tenant.InvestmentAccountID,
        Quantity: "10", AvgCostCents: 5000, Currency: "BRL",
        PurchasedAt: time.Now(), IncomeType: "dividend",
        IncomeIntervalDays: ptr(30), IncomeAmountCents: ptr(int64(150)),
        NextIncomeAt: ptr(time.Now().Add(30 * 24 * time.Hour)),
    })
    require.Equal(t, http.StatusCreated, pos.StatusCode)

    // 3. Check summary
    summary := httpGet(t, srv, "/v1/investments/summary", token)
    require.Equal(t, http.StatusOK, summary.StatusCode)
    require.Greater(t, summary.Body.TotalValueCents, int64(0))
}
```

---

## 5. Acceptance Criteria

- [x] All six new mock repository types compile and implement their respective interfaces.
- [x] `MockInvestmentService` compiles and implements `domain.InvestmentService`.
- [x] All four smoke tests pass against a Postgres testcontainer.
- [x] `TestSmoke_TenantAssetConfig` verifies COALESCE: overridden `name` is returned by `GET /v1/assets/{id}`.
- [x] `TestSmoke_PositionLifecycle` verifies that `GET /v1/investments/summary` returns `int64` values, not floats.
- [x] Seed factories follow the `seeds.Seed*` naming convention established in Phase 1/2.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.15 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                 |
| ---------- | ------ | -------------------------------------- |
| 2026-03-13 | —      | Task created (new)                     |
| 2026-03-15 | Agent  | Task completed; all integration tests pass |
