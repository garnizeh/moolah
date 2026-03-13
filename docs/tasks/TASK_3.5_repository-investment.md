# Task 3.5 — Investment Repository Implementation + Integration Tests

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Data Access
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the concrete repository structs that fulfil the `AssetRepository`, `PositionRepository`, and `PortfolioSnapshotRepository` interfaces defined in Task 3.3. Each repository wraps the generated `sqlc` queries from Task 3.4. Full integration test coverage is required, exercising each repository method against a real Postgres container via `testcontainers-go`.

---

## 2. Context & Motivation

The project separates the repository interface (defined in `internal/domain/`) from its concrete implementation (lived in `internal/platform/repository/`). This enables service-layer unit tests to mock the repository without hitting a database. The concrete implementation is only exercised by integration tests.

Phase 1 established `AccountRepository`, `TransactionRepository`, etc. Phase 2 added `MasterPurchaseRepository`. All follow the same constructor pattern: `NewXRepository(q *sqlc.Queries) *XRepository`. Phase 3 must be consistent.

---

## 3. Scope

### In scope

- [ ] `internal/platform/repository/asset_repository.go` — implements `domain.AssetRepository`.
- [ ] `internal/platform/repository/position_repository.go` — implements `domain.PositionRepository`.
- [ ] `internal/platform/repository/portfolio_snapshot_repository.go` — implements `domain.PortfolioSnapshotRepository`.
- [ ] Integration tests for each repository file (build tag `integration`).
- [ ] Registration of new repositories in mock factory (`internal/testutil/mocks`) if needed.

### Out of scope

- Service layer (Task 3.6).
- DI wiring in `cmd/api/main.go` (Task 3.7 or final wiring task).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                                   | Purpose                                        |
| ------ | ---------------------------------------------------------------------- | ---------------------------------------------- |
| CREATE | `internal/platform/repository/asset_repository.go`                    | Concrete `AssetRepository` implementation      |
| CREATE | `internal/platform/repository/asset_repository_integration_test.go`   | Integration tests for asset repo               |
| CREATE | `internal/platform/repository/position_repository.go`                 | Concrete `PositionRepository` implementation   |
| CREATE | `internal/platform/repository/position_repository_integration_test.go`| Integration tests for position repo            |
| CREATE | `internal/platform/repository/portfolio_snapshot_repository.go`       | Concrete `PortfolioSnapshotRepository` impl    |
| CREATE | `internal/platform/repository/portfolio_snapshot_repository_integration_test.go` | Integration tests for snapshot repo |
| MODIFY | `internal/testutil/mocks/querier_mock.go`                              | Add mock methods for new sqlc queries          |

### Constructor pattern

```go
// asset_repository.go
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
        Name:      input.Name,
        AssetType: sqlc.AssetType(input.AssetType),
        Currency:  input.Currency,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create asset: %w", err)
    }
    return mapAsset(row), nil
}
// ... GetByID, GetByTicker, List
```

### Integration test pattern

```go
//go:build integration

func TestAssetRepository_Create(t *testing.T) {
    t.Parallel()
    pgDB := containers.NewPostgresDB(t)
    repo := repository.NewAssetRepository(pgDB.Queries)

    t.Run("creates asset successfully", func(t *testing.T) {
        t.Parallel()
        asset, err := repo.Create(context.Background(), domain.CreateAssetInput{
            Ticker:    "AAPL",
            Name:      "Apple Inc.",
            AssetType: domain.AssetTypeStock,
            Currency:  "USD",
        })
        require.NoError(t, err)
        assert.NotEmpty(t, asset.ID)
        assert.Equal(t, "AAPL", asset.Ticker)
    })
}
```

---

## 5. Acceptance Criteria

- [ ] All three repository structs implement their respective domain interfaces (compiler-enforced via `var _ domain.AssetRepository = (*AssetRepository)(nil)`).
- [ ] Every method wraps errors with context: `fmt.Errorf("failed to <action>: %w", err)`.
- [ ] `domain.ErrNotFound` is returned when `pgx.ErrNoRows` is encountered.
- [ ] All integration tests use `t.Parallel()` (both parent and subtests).
- [ ] Tenant isolation is verified: creating a position for tenant A is not visible to tenant B.
- [ ] Soft-delete is verified: deleted positions are excluded from list queries.
- [ ] `require.NoError(t, err)` used for every fallible call, including `resp.Body.Close()`.
- [ ] Test coverage for new repository code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.5 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                              | Type     | Status     |
| ------------------------------------------------------- | -------- | ---------- |
| Task 3.2 — Migrations must be applied before testing    | Upstream | 🔵 backlog |
| Task 3.3 — Domain interfaces defined                    | Upstream | 🔵 backlog |
| Task 3.4 — `sqlc generate` output available             | Upstream | 🔵 backlog |
| `internal/testutil/containers.NewPostgresDB`            | Upstream | ✅ done    |
| `internal/testutil/seeds` factories                     | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests

N/A — repository implementations are pure data-access code, fully covered by integration tests.

### Integration tests (`//go:build integration`)

- **Files:** `*_integration_test.go` for each repository.
- **Cases per repository:**
  - Create and retrieve by ID.
  - List returns correct results filtered by tenant.
  - Update modifies only the specified fields.
  - Soft delete excludes record from subsequent list queries.
  - Cross-tenant query returns zero rows.
  - `GetByID` on a missing record returns `domain.ErrNotFound`.

---

## 8. Open Questions

| # | Question                                                                                      | Owner | Resolution |
| - | --------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `NewAssetRepository` accept `*sqlc.Queries` or the `sqlc.Querier` interface for mockability? | — | Follow Phase 2 pattern (use `sqlc.Querier` interface) |
| 2 | Should `PortfolioSnapshotRepository.Create` be upsert (on conflict update) or strict insert? | —     | Strict insert for MVP; upsert if re-snapshotting is required |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
