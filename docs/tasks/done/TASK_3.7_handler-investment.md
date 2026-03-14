# Task 3.7 — HTTP Handlers: Asset Catalogue + Tenant Asset Configs

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › HTTP Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-14
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement the `AssetHandler` struct and register all asset-related HTTP endpoints in `internal/server/routes.go`. Covers the global asset catalogue (admin-only write, authenticated read) and the per-tenant asset config endpoints (tenant read/write). Thin handler layer: decode → validate → delegate to `InvestmentService` → respond.

---

## 2. Context & Motivation

The global `assets` table is admin-managed. Tenants can read asset details and maintain their own configuration overrides via `tenant_asset_configs` (ADR §2.7). These endpoints are logically separate from position management (covered in Task 3.14) and are simpler to implement first.

**Reference:** ADR-003 §2.1, §2.7; handler pattern in `internal/handler/account_handler.go`.

---

## 3. Scope

### In scope

- [x] `internal/handler/asset_handler.go` — `AssetHandler` struct + all handlers.
- [x] `internal/handler/asset_handler_test.go` — unit tests with mocked `InvestmentService`.
- [x] Route registration for all asset endpoints in `internal/server/routes.go`.
- [x] Swaggo annotations on every handler function.
- [x] `RequireRole("admin")` middleware on asset write endpoints.

### Out of scope

- Position endpoints (Task 3.14).
- Income event endpoints (Task 3.14).
- Summary endpoint (Task 3.14).
- DI wiring in `cmd/api/main.go` (final wiring task deferred to Task 3.15 or server.go updates).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                        | Purpose                                        |
| ------ | ------------------------------------------- | ---------------------------------------------- |
| CREATE | `internal/handler/asset_handler.go`         | HTTP handler implementation                    |
| CREATE | `internal/handler/asset_handler_test.go`    | Unit tests with mocked service                 |
| MODIFY | `internal/server/routes.go`                 | Register asset + tenant config routes          |
| MODIFY | `internal/server/server.go`                 | Add `assetHandler` field + constructor update  |

### API Endpoints

| Method | Path                                        | Auth             | Description                                    |
| ------ | ------------------------------------------- | ---------------- | ---------------------------------------------- |
| GET    | `/v1/assets`                                | ✅ Bearer         | List all global assets (all tenants)           |
| GET    | `/v1/assets/{id}`                           | ✅ Bearer         | Get single asset (COALESCE with tenant config) |
| POST   | `/v1/assets`                                | ✅ Bearer + Admin | Create asset (admin only)                      |
| DELETE | `/v1/assets/{id}`                           | ✅ Bearer + Admin | Delete asset (admin only)                      |
| GET    | `/v1/me/asset-configs`                      | ✅ Bearer         | List tenant's asset config overrides           |
| PUT    | `/v1/me/asset-configs/{asset_id}`           | ✅ Bearer         | Upsert tenant asset config                     |
| DELETE | `/v1/me/asset-configs/{asset_id}`           | ✅ Bearer         | Remove tenant asset config (restore defaults)  |

### Request / response types (excerpt)

```go
type CreateAssetRequest struct {
    Ticker    string  `json:"ticker"    validate:"required,max=20"`
    ISIN      *string `json:"isin"      validate:"omitempty,len=12"`
    Name      string  `json:"name"      validate:"required,max=200"`
    AssetType string  `json:"asset_type" validate:"required,oneof=stock bond fund crypto real_estate income_source"`
    Currency  string  `json:"currency"  validate:"required,len=3"`
    Details   *string `json:"details"`
}

type UpsertTenantAssetConfigRequest struct {
    Name     *string `json:"name"     validate:"omitempty,max=200"`
    Currency *string `json:"currency" validate:"omitempty,len=3"`
    Details  *string `json:"details"`
}
```

### Error cases to handle

| Scenario                  | Sentinel Error             | HTTP Status |
| ------------------------- | -------------------------- | ----------- |
| Asset not found           | `domain.ErrAssetNotFound`  | `404`       |
| Non-admin creating asset  | middleware / role check     | `403`       |
| Invalid asset_type value  | validation error           | `422`       |

---

## 5. Acceptance Criteria

- [x] `GET /v1/assets/{id}` returns the global asset with tenant overrides applied (calls `GetAssetWithTenantConfig`).
- [x] `POST /v1/assets` returns `403` for non-admin callers.
- [x] `PUT /v1/me/asset-configs/{asset_id}` creates or updates the config (upsert semantics).
- [x] `DELETE /v1/me/asset-configs/{asset_id}` soft-deletes and returns `204`.
- [x] Swaggo annotations exist on every handler function.
- [x] Unit tests achieve ≥ 80% coverage for the handler package additions.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.7 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                            |
| ---------- | ------ | ------------------------------------------------- |
| 2026-03-14 | Copilot| Task completed: handlers implemented, routes registered, and tests passing with 100% coverage. |
| 2026-03-13 | —      | Task created; rewritten for ADR v3 (asset + tenant_asset_config endpoints only; position endpoints moved to Task 3.14) |
