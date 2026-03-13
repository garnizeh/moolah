# Task 3.3 — Domain: `Asset` + `TenantAssetConfig` Entities & Repository Interfaces

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the `Asset` and `TenantAssetConfig` domain entities, input types, and their repository interfaces in `internal/domain/asset.go`. This is the first domain file for Phase 3 and establishes the global asset catalogue and per-tenant override pattern from ADR-003 §2.1 and §2.7.

---

## 2. Context & Motivation

Phase 3 introduces a two-layer asset model: a global `assets` table (admin-managed, no `tenant_id`) and a `tenant_asset_configs` table (per-tenant sparse overrides). The domain layer defines the Go structs and interfaces that all downstream layers (sqlc, repository, service, handler) depend on.

Following the project convention, every repository must be defined as an interface in `internal/domain/` before any implementation is written. See `internal/domain/account.go` for the Phase 1 pattern.

**Reference:** ADR-003 §2.1, §2.7, §3.1, §3.2.

---

## 3. Scope

### In scope

- [x] `AssetType` string enum with constants for all six values from ADR (`stock`, `bond`, `fund`, `crypto`, `real_estate`, `income_source`).
- [x] `Asset` struct with JSON tags (all fields from ADR §3.1).
- [x] `CreateAssetInput` and `ListAssetsParams` input types.
- [x] `AssetRepository` interface (`Create`, `GetByID`, `GetByTicker`, `List`, `Delete`).
- [x] `TenantAssetConfig` struct with JSON tags (all fields from ADR §3.2).
- [x] `UpsertTenantAssetConfigInput` input type.
- [x] `TenantAssetConfigRepository` interface (`Upsert`, `GetByAssetID`, `ListByTenant`, `Delete`).
- [x] Sentinel errors: `ErrAssetNotFound`, `ErrAssetConfigNotFound`.
- [x] Unit tests in `internal/domain/asset_test.go` covering validation helpers (if any).

### Out of scope

- Position, PositionSnapshot, PositionIncomeEvent, PortfolioSnapshot (Task 3.9).
- CurrencyConverter interface (Task 3.10).
- Concrete repository implementations (Task 3.7).
- SQLC queries (Task 3.4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                              | Purpose                                        |
| ------ | --------------------------------- | ---------------------------------------------- |
| CREATE | `internal/domain/asset.go`        | Entities, input types, interfaces, errors      |
| CREATE | `internal/domain/asset_test.go`   | Unit tests for domain types / validation       |

### Key types

```go
// AssetType categorises the investment instrument.
type AssetType string

const (
    AssetTypeStock        AssetType = "stock"
    AssetTypeBond         AssetType = "bond"
    AssetTypeFund         AssetType = "fund"
    AssetTypeCrypto       AssetType = "crypto"
    AssetTypeRealEstate   AssetType = "real_estate"
    AssetTypeIncomeSource AssetType = "income_source"
)

// Asset is a global, admin-managed reference record — no tenant_id.
type Asset struct {
    ID        string    `json:"id"`
    Ticker    string    `json:"ticker"`
    ISIN      *string   `json:"isin,omitempty"`
    Name      string    `json:"name"`
    AssetType AssetType `json:"asset_type"`
    Currency  string    `json:"currency"`
    Details   *string   `json:"details,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

// TenantAssetConfig holds a tenant's sparse overrides for a global asset.
// Fields that are nil fall back to the global asset value.
type TenantAssetConfig struct {
    ID        string     `json:"id"`
    TenantID  string     `json:"-"`
    AssetID   string     `json:"asset_id"`
    Name      *string    `json:"name,omitempty"`      // overrides Asset.Name
    Currency  *string    `json:"currency,omitempty"`  // overrides Asset.Currency
    Details   *string    `json:"details,omitempty"`   // overrides Asset.Details
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"-"`
}

// AssetRepository defines persistence operations for the global asset catalogue.
```

### SQL queries (sqlc) — reference only; implemented in Task 3.4

```sql
-- name: CreateAsset :one
-- name: GetAssetByID :one
-- name: GetAssetByTicker :one
-- name: ListAssets :many
-- name: DeleteAsset :exec
-- name: UpsertTenantAssetConfig :one
-- name: GetTenantAssetConfigByAssetID :one
-- name: ListTenantAssetConfigs :many
-- name: SoftDeleteTenantAssetConfig :exec
-- name: GetAssetWithTenantConfig :one   (COALESCE merge — see ADR §2.7)
```

### Error cases to handle

| Scenario              | Sentinel Error              | HTTP Status |
| --------------------- | --------------------------- | ----------- |
| Asset not found       | `domain.ErrAssetNotFound`   | `404`       |
| Config not found      | `domain.ErrAssetConfigNotFound` | `404`   |
| Tenant mismatch       | `domain.ErrForbidden`       | `403`       |
| Duplicate ticker      | unique constraint wrapped   | `409`       |

---

## 5. Acceptance Criteria

- [x] `AssetType` has all six constants matching ADR enum values exactly.
- [x] `Asset` struct has `ISIN *string`, `Details *string` (nullable as per ADR).
- [x] `TenantAssetConfig` has nullable `Name`, `Currency`, `Details` fields.
- [x] `AssetRepository` interface is defined in `internal/domain/`.
- [x] `TenantAssetConfigRepository` interface is defined in `internal/domain/`.
- [x] Sentinel errors are exported from `internal/domain/`.
- [x] Unit tests cover all `AssetType` constant values.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.3 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                 |
| ---------- | ------ | -------------------------------------- |
| 2026-03-13 | —      | Task created; rewritten for ADR v3 (Asset + TenantAssetConfig only) |
| 2026-03-13 | —      | Implemented domain structs, interfaces, tests and passed lint. |
