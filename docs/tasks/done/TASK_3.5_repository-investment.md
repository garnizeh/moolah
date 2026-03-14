# Task 3.5 — Repository: `AssetRepository` + `TenantAssetConfigRepository` + Integration Tests

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Data Access
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement concrete repository structs that satisfy the `domain.AssetRepository` and `domain.TenantAssetConfigRepository` interfaces (defined in Task 3.3) by wrapping the sqlc-generated queries from Task 3.4. Full integration test coverage is required using `testcontainers-go`.

---

## 2. Context & Motivation

The repository layer is the only place that calls sqlc-generated code. It maps `sqlc` structs to domain entities and wraps errors with context. Service-layer unit tests will never import this package — they mock the interface defined in `internal/domain/`.

The `TenantAssetConfigRepository.Upsert` method is a critical path: it must use the `ON CONFLICT … DO UPDATE` pattern (sqlc `UpsertTenantAssetConfig` query) so that a second call for the same `(tenant_id, asset_id)` updates rather than creates.

**Reference:** ADR-003 §2.1, §2.7; Phase 2 pattern in `internal/platform/repository/master_purchase_repository.go`.

---

## 3. Scope

### In scope

- [ ] `internal/platform/repository/asset_repository.go` — implements `domain.AssetRepository`.
- [ ] `internal/platform/repository/asset_repository_integration_test.go` — integration tests (build tag `integration`).
- [ ] `internal/platform/repository/tenant_asset_config_repository.go` — implements `domain.TenantAssetConfigRepository`.
- [ ] `internal/platform/repository/tenant_asset_config_repository_integration_test.go` — integration tests.
- [ ] Add mock methods for new query types to `internal/testutil/mocks/querier_mock.go`.

### Out of scope

- Position-family repositories (Task 3.12).
- Service layer (Task 3.6).
- DI wiring in `cmd/api/main.go` (deferred to handler tasks).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                                             | Purpose                                       |
| ------ | -------------------------------------------------------------------------------- | --------------------------------------------- |
| CREATE | `internal/platform/repository/asset_repository.go`                              | `AssetRepository` concrete implementation     |
| CREATE | `internal/platform/repository/asset_repository_integration_test.go`             | Integration tests for asset repo              |
| CREATE | `internal/platform/repository/tenant_asset_config_repository.go`                | `TenantAssetConfigRepository` implementation  |
| CREATE | `internal/platform/repository/tenant_asset_config_repository_integration_test.go` | Integration tests                           |
| MODIFY | `internal/testutil/mocks/querier_mock.go`                                        | Add mocks for new sqlc query methods          |

### Constructor pattern

```go
type AssetRepository struct {
    q *sqlc.Queries
}

func NewAssetRepository(q *sqlc.Queries) *AssetRepository {
    return &AssetRepository{q: q}
}

func (r *AssetRepository) Create(ctx context.Context, input domain.CreateAssetInput) (*domain.Asset, error) {
    row, err := r.q.CreateAsset(ctx, sqlc.CreateAssetParams{
        ID:        ulid.New(),
        Ticker:    input.Ticker,
        Isin:      pgtype.Text{String: deref(input.ISIN), Valid: input.ISIN != nil},
        Name:      input.Name,
        AssetType: sqlc.AssetType(input.AssetType),
        Currency:  input.Currency,
        Details:   pgtype.Text{String: deref(input.Details), Valid: input.Details != nil},
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create asset: %w", err)
    }
    return mapAsset(row), nil
}
```

### Integration test pattern

```go
//go:build integration

func TestAssetRepository_Create(t *testing.T) {
    t.Parallel()
    db, cleanup := containers.NewPostgresDB(t)
    defer cleanup()

    q  := sqlc.New(db)
    r  := repository.NewAssetRepository(q)

    asset, err := r.Create(ctx, domain.CreateAssetInput{
        Ticker:    "AAPL",
        Name:      "Apple Inc.",
        AssetType: domain.AssetTypeStock,
        Currency:  "USD",
    })
    require.NoError(t, err)
    require.Equal(t, "AAPL", asset.Ticker)
}
```

---

## 5. Acceptance Criteria

- [x] `AssetRepository` implements all methods of `domain.AssetRepository`.
- [x] `TenantAssetConfigRepository` implements all methods of `domain.TenantAssetConfigRepository`.
- [x] `Upsert` on an existing `(tenant_id, asset_id)` updates rather than creating a duplicate.
- [x] `GetByAssetID` returns `domain.ErrAssetConfigNotFound` when no active config exists.
- [x] Every integration test calls `t.Parallel()` inside the subtest.
- [x] All error returns are wrapped with `fmt.Errorf("…: %w", err)`.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.5 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                             |
| ---------- | ------ | -------------------------------------------------- |
| 2026-03-13 | —      | Task created; rewritten for ADR v3 (asset + tenant_asset_config repos only) |
| 2026-03-13 | github-copilot | Implemented AssetRepository and TenantAssetConfigRepository with integration tests. |
