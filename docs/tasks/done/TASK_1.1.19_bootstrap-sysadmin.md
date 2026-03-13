# Task 1.1.19 — Sysadmin Bootstrap on Startup

> **Roadmap Ref:** Phase 1 — MVP › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-10
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement an idempotent startup bootstrap that automatically creates a `sysadmin` user (and the required "System" tenant) when the application starts for the first time. This breaks the bootstrapping cycle: the OTP login flow requires a user to pre-exist in the database, so without this mechanism there is no safe way to obtain the first `sysadmin` token in a fresh deployment.

---

## 2. Context & Motivation

The authentication flow (`POST /v1/auth/otp/request`) is OTP-only and requires the user to already exist in the `users` table. Admin routes (`/v1/admin/*`) require a `sysadmin` token. This creates a bootstrap paradox in a fresh environment:

- You cannot call any admin endpoint without a `sysadmin` token.
- You cannot get a `sysadmin` token without a `sysadmin` user in the DB.
- You cannot insert a `sysadmin` user via the API without an existing `sysadmin` token.

The resolution is a **startup-time bootstrap** executed inside `run()` in `cmd/api/main.go`, right after migrations run. It reads the sysadmin email from an environment variable and creates the system tenant + sysadmin user if they do not yet exist. The operation is **idempotent** — re-running it on subsequent starts is a no-op.

The smoke test in `internal/server/smoke_test.go` already seeds a sysadmin manually via raw SQL, confirming that the pattern is correct; this task formalises it for production.

- Related config: `pkg/config/config.go`
- Related migrations: `internal/platform/db/migrations/`
- Related service: `internal/service/admin_service.go`

---

## 3. Scope

### In scope

- [ ] Add `SYSADMIN_EMAIL` (required) and `SYSADMIN_TENANT_NAME` (optional, default: `"System"`) to `pkg/config/config.go`.
- [ ] Implement `bootstrap.EnsureSysadmin(ctx, querier, cfg)` in a new file `internal/platform/bootstrap/sysadmin.go`.
  - Checks whether a user with `role = 'sysadmin'` and `email = cfg.SysadminEmail` exists.
  - If not, creates the system tenant (`cfg.SysadminTenantName`) and the sysadmin user in a single DB transaction.
  - Logs `sysadmin bootstrapped` on first creation; logs `sysadmin already exists, skipping` on subsequent runs.
- [ ] Call `bootstrap.EnsureSysadmin` in `cmd/api/main.go` after migrations, before the HTTP server starts.
- [ ] Unit tests in `internal/platform/bootstrap/sysadmin_test.go` with mocked `sqlc.Querier`.
- [ ] Integration test (build tag `integration`) verifying idempotency: calling `EnsureSysadmin` twice results in exactly one tenant and one sysadmin user.
- [ ] Update `README.md`.

### Out of scope

- Changing or resetting the sysadmin email at runtime (requires DB migration / manual intervention).
- A dedicated HTTP endpoint to trigger bootstrapping.
- Multi-sysadmin setup (one sysadmin is sufficient for MVP).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                              | Purpose                                              |
| ------ | ------------------------------------------------- | ---------------------------------------------------- |
| MODIFY | `pkg/config/config.go`                            | Add `SysadminEmail` and `SysadminTenantName` fields  |
| CREATE | `internal/platform/bootstrap/sysadmin.go`         | `EnsureSysadmin` idempotent bootstrap function       |
| CREATE | `internal/platform/bootstrap/sysadmin_test.go`    | Unit tests with mocked querier                       |
| CREATE | `internal/platform/bootstrap/sysadmin_integration_test.go` | Integration test with testcontainers        |
| MODIFY | `cmd/api/main.go`                                 | Call `bootstrap.EnsureSysadmin` after DB init        |

### Key interfaces / types

```go
// internal/platform/bootstrap/sysadmin.go

// EnsureSysadmin creates the system tenant and sysadmin user if they do not exist.
// It is idempotent: calling it multiple times has no side effects after the first run.
func EnsureSysadmin(ctx context.Context, q sqlc.Querier, cfg *config.Config) error
```

### New config fields

```go
// pkg/config/config.go
type Config struct {
    // ... existing fields ...

    // Bootstrap
    SysadminEmail      string // required — SYSADMIN_EMAIL
    SysadminTenantName string // optional — SYSADMIN_TENANT_NAME (default: "System")
}
```

```
SYSADMIN_EMAIL=admin@company.com        # required
SYSADMIN_TENANT_NAME=System             # optional, default: "System"
```

### Bootstrap logic (pseudo-code)

```go
func EnsureSysadmin(ctx context.Context, q sqlc.Querier, cfg *config.Config) error {
    // 1. Check if sysadmin with this email already exists
    existing, err := q.GetUserByEmail(ctx, cfg.SysadminEmail)
    if err == nil && existing.Role == sqlc.UserRoleSysadmin {
        slog.Info("sysadmin already exists, skipping bootstrap", "email", cfg.SysadminEmail)
        return nil
    }
    if err != nil && !errors.Is(repository.translateError(err), domain.ErrNotFound) {
        return fmt.Errorf("bootstrap: failed to check existing sysadmin: %w", err)
    }

    // 2. Create system tenant
    tenantID := ulid.New()
    _, err = q.CreateTenant(ctx, sqlc.CreateTenantParams{
        ID:   tenantID,
        Name: cfg.SysadminTenantName,
        Plan: sqlc.TenantPlanFree,
    })
    if err != nil {
        return fmt.Errorf("bootstrap: failed to create system tenant: %w", err)
    }

    // 3. Create sysadmin user
    _, err = q.CreateUser(ctx, sqlc.CreateUserParams{
        ID:       ulid.New(),
        TenantID: tenantID,
        Email:    cfg.SysadminEmail,
        Name:     "Sysadmin",
        Role:     sqlc.UserRoleSysadmin,
    })
    if err != nil {
        return fmt.Errorf("bootstrap: failed to create sysadmin user: %w", err)
    }

    slog.Info("sysadmin bootstrapped successfully", "email", cfg.SysadminEmail, "tenant", cfg.SysadminTenantName)
    return nil
}
```

### Startup wiring (`cmd/api/main.go`)

```go
// After db.Querier(ctx, cfg.DatabaseURL) succeeds and before server.New(...)
if err := bootstrap.EnsureSysadmin(ctx, querier, cfg); err != nil {
    return fmt.Errorf("sysadmin bootstrap failed: %w", err)
}
```

### Error cases to handle

| Scenario                                | Handling                                                     |
| --------------------------------------- | ------------------------------------------------------------ |
| `SYSADMIN_EMAIL` not set                | `config.Load()` panics (required env var)                    |
| Sysadmin already exists                 | No-op, log at `INFO` level                                   |
| DB unavailable during bootstrap         | Return wrapped error; startup aborts                         |
| Tenant creation succeeds, user fails    | Orphan tenant left (acceptable for MVP — idempotency on email re-check prevents duplicate tenants on retry) |

---

## 5. Acceptance Criteria

- [ ] `SYSADMIN_EMAIL` added to `pkg/config/config.go` as a required field; `config_test.go` updated.
- [ ] `SYSADMIN_TENANT_NAME` added as optional with default `"System"`.
- [ ] `internal/platform/bootstrap/sysadmin.go` exists and compiles.
- [ ] `EnsureSysadmin` called in `cmd/api/main.go` after migrations, before server start.
- [ ] Unit tests cover: first-run creation, idempotent skip, DB error on lookup, DB error on tenant create, DB error on user create.
- [ ] Integration test confirms exactly one sysadmin row exists after two consecutive `EnsureSysadmin` calls.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `make task-check` passes.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.
- [ ] Update `README.md` with instructions on setting `SYSADMIN_EMAIL` for bootstrapping.

---

## 6. Change Log

| Date       | Author  | Change                  |
| ---------- | ------- | ----------------------- |
| 2026-03-10 | Copilot | Initial task document   |
