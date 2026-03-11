# Task 1.5.6 — `handler/account_handler.go` — full CRUD

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** ✅ `done`
> **Last Updated:** 2025-03-20
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the account HTTP handler in `internal/handler/account_handler.go`. It exposes full CRUD endpoints for financial accounts (checking, savings, cash, investment) scoped to the authenticated user's tenant. The handler delegates to `AccountService` and translates domain errors to HTTP status codes.

---

## 2. Context & Motivation

The `AccountService` is fully implemented (Task 1.4.3) but has no HTTP entry point. Accounts are the foundation of the cash flow ledger and must be accessible via the API. See `docs/ARCHITECTURE.md` and roadmap item 1.5.6.

---

## 3. Scope

### In scope

- [x] `internal/handler/account_handler.go` — `AccountHandler` struct + 5 HTTP handler methods.
- [x] `GET /v1/accounts` — list all accounts for the tenant.
- [x] `POST /v1/accounts` — create a new account.
- [x] `GET /v1/accounts/{id}` — get a single account by ID.
- [x] `PATCH /v1/accounts/{id}` — partial update (name, description, currency).
- [x] `DELETE /v1/accounts/{id}` — soft delete.
- [x] Unit tests in `internal/handler/account_handler_test.go`.

### Out of scope

- Balance recalculation endpoint — handled internally by `AccountService`.
- Admin cross-tenant account access — Task 1.5.9.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                       | Purpose                              |
| ------ | ------------------------------------------ | ------------------------------------ |
| CREATE | `internal/handler/account_handler.go`      | HTTP handler for account endpoints   |
| CREATE | `internal/handler/account_handler_test.go` | Unit tests with mocked AccountService|

### Request / Response types

```go
type CreateAccountRequest struct {
    Name        string            `json:"name"        validate:"required,min=1,max=100"`
    Type        domain.AccountType `json:"type"        validate:"required"`
    Currency    string            `json:"currency"    validate:"required,len=3"`
    Description *string           `json:"description" validate:"omitempty,max=255"`
    Color       *string           `json:"color"       validate:"omitempty"`
}

type UpdateAccountRequest struct {
    Name        *string `json:"name"        validate:"omitempty,min=1,max=100"`
    Description *string `json:"description" validate:"omitempty,max=255"`
    Color       *string `json:"color"       validate:"omitempty"`
}
```

### API endpoints

| Method | Path                  | Auth Required | Description              |
| ------ | --------------------- | ------------- | ------------------------ |
| GET    | `/v1/accounts`        | ✅ Bearer     | List all accounts        |
| POST   | `/v1/accounts`        | ✅ Bearer     | Create an account        |
| GET    | `/v1/accounts/{id}`   | ✅ Bearer     | Get account by ID        |
| PATCH  | `/v1/accounts/{id}`   | ✅ Bearer     | Update account           |
| DELETE | `/v1/accounts/{id}`   | ✅ Bearer     | Soft-delete account      |

### Error cases to handle

| Scenario                  | Sentinel Error           | HTTP Status |
| ------------------------- | ------------------------ | ----------- |
| Not found                 | `domain.ErrNotFound`     | `404`       |
| Validation failure        | —                        | `422`       |
| Conflict (duplicate name) | `domain.ErrConflict`     | `409`       |
| Forbidden                 | `domain.ErrForbidden`    | `403`       |

---

## 5. Acceptance Criteria

- [x] All 5 endpoints decode, validate, and call the service correctly.
- [x] `tenant_id` is always sourced from context (never from the request body).
- [x] All domain error sentinels map to the correct HTTP status codes.
- [x] `POST` and `PATCH` responses include the created/updated resource.
- [x] Unit tests cover all happy paths and error cases.
- [x] Test coverage for handler ≥ 80%.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                              | Type     | Status  |
| --------------------------------------- | -------- | ------- |
| Task 1.4.3 — `service/account_service`  | Upstream | ✅ done |
| Task 1.1.9 — Auth middleware            | Upstream | ✅ done |
| `domain.AccountService` interface       | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/account_handler_test.go`
- **Cases:**
  - `List`: returns accounts array → `200 OK`.
  - `Create`: valid body → `201 Created` with resource.
  - `Create`: invalid body → `422`.
  - `Create`: duplicate name → `409`.
  - `GetByID`: found → `200 OK`.
  - `GetByID`: not found → `404`.
  - `Update`: valid partial update → `200 OK`.
  - `Delete`: success → `204 No Content`.
  - `Delete`: not found → `404`.

---

## 8. Open Questions

| # | Question                                         | Owner | Resolution |
| - | ------------------------------------------------ | ----- | ---------- |
| 1 | Should balance be exposed in the list response?  | —     | Yes — include `balance_cents` in the response struct. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
