# Task 3.8 — Monthly Portfolio Snapshot Job (`portfolio_snapshots`)

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Background Jobs
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement a background job — `SnapshotJob` — that automatically generates a `PortfolioSnapshot` for every active tenant at the start of each month. The job uses the `InvestmentService.TakeSnapshot` method (Task 3.6) and is wired into the application's startup lifecycle via a dedicated goroutine. For Phase 3, the scheduler is a simple in-process ticker; a more robust distributed cron (e.g., Redis-backed) is deferred to Phase 5.

---

## 2. Context & Motivation

Monthly snapshots provide the historical data required for portfolio performance charting. Without automated snapshotting, tenants could only see their current position value — not how it evolved over time. The Phase 2 pattern for background work (`InvoiceCloser`) used a synchronous endpoint trigger; portfolio snapshotting is better suited to an async job because it affects all tenants simultaneously and doesn't require a per-tenant manual action.

The job must be safe to run concurrently with normal API traffic and must be idempotent — running it twice in the same month for the same tenant must produce at most one snapshot (enforced by the `UNIQUE(tenant_id, snapshot_date)` constraint in the DB).

---

## 3. Scope

### In scope

- [ ] `internal/service/snapshot_job.go` — `SnapshotJob` struct with `Run(ctx context.Context)` method.
- [ ] `internal/service/snapshot_job_test.go` — unit tests with mocked service and tenant list.
- [ ] Wiring in `cmd/api/main.go`: start the job goroutine at server startup.
- [ ] Graceful shutdown: the job listens on `ctx.Done()` for clean exit.
- [ ] Structured logging on each tenant processed, with errors logged but not fatal.
- [ ] Idempotency: if snapshot already exists for the month (unique constraint), log and skip.

### Out of scope

- Distributed lock / Redis-backed cron (deferred to Phase 5).
- Per-tenant configurable snapshot frequency.
- Backfilling historical snapshots for existing tenants (separate migration/script).
- Manual trigger endpoint (available via `POST /v1/accounts/{id}/snapshot` in Task 3.7).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                       | Purpose                                                 |
| ------ | ------------------------------------------ | ------------------------------------------------------- |
| CREATE | `internal/service/snapshot_job.go`         | Background job struct + ticker loop                     |
| CREATE | `internal/service/snapshot_job_test.go`    | Unit tests (mock service, mock tenant list)             |
| MODIFY | `cmd/api/main.go`                          | Start job goroutine; pass cancel context from server    |

### Struct and Run loop

```go
// SnapshotJob generates monthly portfolio snapshots for all active tenants.
type SnapshotJob struct {
    investmentSvc domain.InvestmentService
    tenantRepo    domain.TenantRepository
    logger        *slog.Logger
    interval      time.Duration // default: 24h; checked daily, fires on month boundary
}

func NewSnapshotJob(
    investmentSvc domain.InvestmentService,
    tenantRepo    domain.TenantRepository,
    logger        *slog.Logger,
) *SnapshotJob {
    return &SnapshotJob{
        investmentSvc: investmentSvc,
        tenantRepo:    tenantRepo,
        logger:        logger,
        interval:      24 * time.Hour,
    }
}

// Run starts the daily ticker loop. It returns when ctx is cancelled.
func (j *SnapshotJob) Run(ctx context.Context) {
    ticker := time.NewTicker(j.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            j.logger.InfoContext(ctx, "snapshot_job: shutting down")
            return
        case t := <-ticker.C:
            if t.Day() == 1 { // first day of the month
                j.runOnce(ctx)
            }
        }
    }
}

func (j *SnapshotJob) runOnce(ctx context.Context) {
    tenants, err := j.tenantRepo.ListAll(ctx)
    if err != nil {
        j.logger.ErrorContext(ctx, "snapshot_job: failed to list tenants", "error", err)
        return
    }
    for _, tenant := range tenants {
        if _, err := j.investmentSvc.TakeSnapshot(ctx, tenant.ID); err != nil {
            j.logger.WarnContext(ctx, "snapshot_job: failed to take snapshot",
                "tenant_id", tenant.ID, "error", err)
        }
    }
}
```

### Wiring in `main.go`

```go
snapshotJob := service.NewSnapshotJob(investmentSvc, tenantRepo, logger)
go snapshotJob.Run(ctx) // ctx is the application-level context tied to OS signal
```

---

## 5. Acceptance Criteria

- [ ] `SnapshotJob.Run` starts without blocking the HTTP server.
- [ ] Job fires once on the first day of the month only (not every 24 h on other days).
- [ ] Errors in individual tenant snapshots are logged but do not abort the full run.
- [ ] `ctx.Done()` causes the job to exit cleanly within one tick period.
- [ ] Idempotency: running `runOnce` twice in the same month for the same tenant produces only one snapshot (DB uniqueness enforced; error logged and skipped).
- [ ] Unit tests cover: tick fires on day 1 → all tenants processed; tick fires on day 15 → no-op; tenant snapshot error → loop continues.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 3.8 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                                    | Type     | Status     |
| ------------------------------------------------------------- | -------- | ---------- |
| Task 3.6 — `InvestmentService.TakeSnapshot` implemented       | Upstream | 🔵 backlog |
| `domain.TenantRepository.ListAll` method (may need adding)    | Upstream | 🔵 backlog |
| Application-level `context.Context` with cancel in `main.go`  | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/service/snapshot_job_test.go`
- **Cases:**
  - `runOnce` calls `TakeSnapshot` for each tenant returned by `ListAll`.
  - `runOnce` logs warn and continues when one tenant fails.
  - `Run` exits when context is cancelled.
  - `Run` does not call `runOnce` when ticker fires on day != 1.

### Integration tests

N/A for the scheduler itself. End-to-end flow (snapshot appears in DB after job fires) can be added in a dedicated integration test if needed.

---

## 8. Open Questions

| # | Question                                                                                      | Owner | Resolution |
| - | --------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `TenantRepository` expose a `ListAll` method, or should the job use `AdminService`?   | —     | Prefer `tenantRepo.ListAll(ctx)` to avoid coupling to admin layer |
| 2 | Should a missed month (e.g., server was down on day 1) trigger a backfill on next startup?   | —     | Not for MVP; document as known limitation |
| 3 | Should the interval be configurable via env var for testing?                                  | —     | Yes — add `WithInterval(d time.Duration)` functional option |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
