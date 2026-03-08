# Task 1.4.2 — `service/tenant_service.go` + unit tests

> **Roadmap Ref:** Phase 1 — MVP › 1.4 Service Layer
> **Status:** � `in-progress`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Implement the `TenantService` which orchestrates CRUD operations for tenants and user invitation. It wraps `TenantRepository` and `UserRepository` with business rules: plan enforcement, name uniqueness messaging, and audit trail writes. All operations are fully unit-tested with mocked dependencies.

---

## 2. Context & Motivation

A `Tenant` is the root entity of the multi-tenancy hierarchy — it represents a household. Tenant management is a sysadmin concern (create/delete/update plans), while self-service operations (read own tenant, invite a user) are available to `admin` role users within a tenant. See `docs/ARCHITECTURE.md` and roadmap item 1.4.2.

---

## 3. Scope

### In scope

- [ ] `internal/service/tenant_service.go` — concrete `TenantService` struct with constructor.
- [ ] `internal/domain/tenant.go` — add `TenantService` interface definition.
- [ ] `GetByID(ctx, id string) (*Tenant, error)` — retrieve tenant by ID.
- [ ] `List(ctx) ([]Tenant, error)` — list all tenants (sysadmin scope).
- [ ] `Create(ctx, input CreateTenantInput) (*Tenant, error)` — create new tenant; audit log `create`.
- [ ] `Update(ctx, id string, input UpdateTenantInput) (*Tenant, error)` — update name/plan; audit log `update`.
- [ ] `Delete(ctx, id string) error` — soft-delete tenant; audit log `soft_delete`.
- [ ] `InviteUser(ctx, tenantID string, input CreateUserInput) (*User, error)` — create a user within the tenant (role must be `member` or `admin`); audit log `create`.
- [ ] Full unit tests in `internal/service/tenant_service_test.go`.

### Out of scope

- HTTP handler (`handler/tenant_handler.go`) — Task 1.5.5.
- Hard delete — available only via `AdminService` (Task 1.4.6).
- Plan quota enforcement — Task 4.2.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                       | Purpose                                         |
| ------ | ------------------------------------------ | ----------------------------------------------- |
| MODIFY | `internal/domain/tenant.go`                | Add `TenantService` interface                   |
| CREATE | `internal/service/tenant_service.go`       | Concrete service implementation                 |
| CREATE | `internal/service/tenant_service_test.go`  | Unit tests with mocked deps                     |

### Key interfaces / types

```go
// TenantService defines the business-logic contract for tenant management.
type TenantService interface {
    Create(ctx context.Context, input CreateTenantInput) (*Tenant, error)
    GetByID(ctx context.Context, id string) (*Tenant, error)
    List(ctx context.Context) ([]Tenant, error)
    Update(ctx context.Context, id string, input UpdateTenantInput) (*Tenant, error)
    Delete(ctx context.Context, id string) error
    InviteUser(ctx context.Context, tenantID string, input CreateUserInput) (*User, error)
}
```

### Business rules

1. `Create`: delegate to `TenantRepository.Create`; on `ErrConflict` surface it; write audit log `create`.
2. `Update`: fetch first via `GetByID`; apply partial changes; write audit log `update` with `OldValues`/`NewValues`.
3. `Delete`: write audit log `soft_delete` before deleting.
4. `InviteUser`: validate `input.TenantID == tenantID`; delegate to `UserRepository.Create`; write audit log `create` (entity: `user`).

### Error cases

| Scenario                   | Sentinel Error         | Notes                        |
| -------------------------- | ---------------------- | ---------------------------- |
| Tenant not found           | `domain.ErrNotFound`   | `GetByID`, `Update`, `Delete`|
| Duplicate tenant name      | `domain.ErrConflict`   | `Create`                     |
| Invalid input              | validation error       | all mutating methods         |

---

## 5. Acceptance Criteria

- [ ] `TenantService` interface defined in `internal/domain/tenant.go`.
- [ ] `NewTenantService` constructor accepts `TenantRepository`, `UserRepository`, and `AuditRepository`.
- [ ] `Create` returns `ErrConflict` on duplicate name and writes no audit log.
- [ ] `Update` writes audit log with old and new values.
- [ ] `Delete` writes audit log before delegating.
- [ ] `InviteUser` delegates to `UserRepository.Create` and writes audit log.
- [ ] Unit tests cover all happy paths and all error branches (≥ 80% coverage).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status  |
| -------------------------------- | -------- | ------- |
| Task 1.3.1 — `tenant_repo.go`    | Upstream | ✅ done |
| Task 1.3.2 — `user_repo.go`      | Upstream | ✅ done |
| Task 1.3.7 — `audit_repo.go`     | Upstream | ✅ done |
| Task 1.1.17 — `testutil/mocks`   | Upstream | ✅ done |

---

## 7. Testing Plan

Unit tests only. Use `testutil/mocks` for all repository dependencies.

Key scenarios:

- `Create` → success → audit log written.
- `Create` → duplicate name → `ErrConflict`, no audit log.
- `GetByID` → not found → `ErrNotFound`.
- `Update` → success → audit log with old/new values.
- `Delete` → success → audit log `soft_delete`.
- `InviteUser` → success → user created, audit log written.

---

## 8. Open Questions

| # | Question                                                                     | Owner | Resolution |
| - | ---------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `InviteUser` send a welcome email (triggering OTP on first login)?    | —     | No — first login triggers OTP naturally via `RequestOTP`. |
| 2 | Should `List` support pagination in Phase 1?                                 | —     | No pagination for now; list is sysadmin-only and bounded. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
