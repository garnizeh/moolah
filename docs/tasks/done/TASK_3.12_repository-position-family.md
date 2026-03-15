# Task 3.12 — Repository: Position Family + Integration Tests

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Data Access
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-14
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement concrete repository structs satisfying `domain.PositionRepository`, `domain.PositionSnapshotRepository`, `domain.PositionIncomeEventRepository`, and `domain.PortfolioSnapshotRepository` (defined in Task 3.9), wrapping the sqlc queries from Task 3.11. Full integration test coverage via `testcontainers-go` is required.

---

## 2. Context & Motivation

Task 3.5 delivered the asset subsystem repositories. This task delivers the four remaining position-family repositories. The most complex method is `UpdateStatus` in `PositionIncomeEventRepository`, which must validate the state transition (`pending → received | cancelled`) and return the updated record.

Append-only tables (`position_snapshots`, `portfolio_snapshots`) have no `Update` or `Delete` methods — any attempt to call them should return a compile-time error (simply don't define them).

**Reference:** ADR-003 §3.3–3.6; Phase 2 pattern in `internal/platform/repository/`.

---

## 3. Scope

### In scope

- [x] `internal/platform/repository/position_repository.go` — implements `domain.PositionRepository`.
- [x] `internal/platform/repository/position_repository_integration_test.go`.
- [x] `internal/platform/repository/position_snapshot_repository.go` — implements `domain.PositionSnapshotRepository`.
- [x] `internal/platform/repository/position_snapshot_repository_integration_test.go`.
- [x] `internal/platform/repository/position_income_event_repository.go` — implements `domain.PositionIncomeEventRepository`.
- [x] `internal/platform/repository/position_income_event_repository_integration_test.go`.
- [x] `internal/platform/repository/portfolio_snapshot_repository.go` — implements `domain.PortfolioSnapshotRepository`.
- [x] `internal/platform/repository/portfolio_snapshot_repository_integration_test.go`.
- [x] Add mock methods for new query types to `internal/testutil/mocks/querier_mock.go`.

### Out of scope

- Asset/TenantAssetConfig repositories (Task 3.5).
- Service layer (Task 3.6).
- DI wiring in `cmd/api/main.go`.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                                                    | Purpose                                   |
| ------ | --------------------------------------------------------------------------------------- | ----------------------------------------- |
| CREATE | `internal/platform/repository/position_repository.go`                                  | `PositionRepository` implementation       |
| CREATE | `internal/platform/repository/position_repository_integration_test.go`                 | Integration tests                         |
| CREATE | `internal/platform/repository/position_snapshot_repository.go`                         | `PositionSnapshotRepository` impl         |
| CREATE | `internal/platform/repository/position_snapshot_repository_integration_test.go`        | Integration tests                         |
| CREATE | `internal/platform/repository/position_income_event_repository.go`                     | `PositionIncomeEventRepository` impl      |
| CREATE | `internal/platform/repository/position_income_event_repository_integration_test.go`    | Integration tests                         |
| CREATE | `internal/platform/repository/portfolio_snapshot_repository.go`                        | `PortfolioSnapshotRepository` impl        |
| CREATE | `internal/platform/repository/portfolio_snapshot_repository_integration_test.go`       | Integration tests                         |
| MODIFY | `internal/testutil/mocks/querier_mock.go`                                               | Add mocks for new sqlc query methods      |

### Constructor pattern (same for all four)

```go
type PositionRepository struct { q *sqlc.Queries }
func NewPositionRepository(q *sqlc.Queries) *PositionRepository { ... }
```

### Critical method

```go
// UpdateStatus in PositionIncomeEventRepository transitions pending→received|cancelled.
// The DB constraint does not enforce state transitions; the repository enforces:
// if current status != "pending", return domain.ErrConflict.
func (r *PositionIncomeEventRepository) UpdateStatus(
    ctx context.Context,
    tenantID, id string,
    status domain.ReceivableStatus,
    receivedAt *time.Time,
) (*domain.PositionIncomeEvent, error) { ... }
```

---

## 5. Acceptance Criteria

- [x] All four repositories implement their respective `domain.*Repository` interfaces.
- [x] `PositionRepository.ListDueIncome` returns only positions where `next_income_at <= before` and `income_type != 'none'`.
- [x] `PositionIncomeEventRepository.UpdateStatus` returns `domain.ErrConflict` if current status is not `pending`.
- [x] Append-only repos (`PositionSnapshot`, `PortfolioSnapshot`) have no `Update` or `Delete` methods.
- [x] Every integration test calls `t.Parallel()`.
- [x] All errors are wrapped with meaningful context.
- [x] Mock querier additions compile without errors.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.12 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                          |
| ---------- | ------ | ----------------------------------------------- |
| 2026-03-14 | Copilot| Position family repositories logic and 100% UT  |
| 2026-03-14 | Copilot| Integration tests and security/lint fixes (G115)|
| 2026-03-13 | —      | Task created (new)                              |
