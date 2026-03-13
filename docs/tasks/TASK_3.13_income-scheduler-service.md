# Task 3.13 — Income Scheduler Service (Background Goroutine — ADR §9)

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Background Jobs
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement `IncomeSchedulerJob` — a continuously running background goroutine that polls `positions.next_income_at` and automatically creates `position_income_events` rows (status `pending`) when income becomes due. After creating the event, it advances `next_income_at` by `income_interval_days`. See ADR-003 §9.

---

## 2. Context & Motivation

Fixed-income positions (rent, salary, coupons, dividends) need periodic income events created without tenant interaction. The income scheduler is the automated engine behind the receivables ledger: it produces the `pending` rows that tenants later mark as `received`.

This job is distinct from the Portfolio Snapshot Job (Task 3.8): the snapshot job runs monthly on a cron schedule; the income scheduler runs on a tighter polling interval (default: every hour) checking `next_income_at <= NOW()`.

**Reference:** ADR-003 §9 (Fixed-Income & Yield Model).

---

## 3. Scope

### In scope

- [ ] `internal/service/income_scheduler_job.go` — `IncomeSchedulerJob` struct + `Run(ctx context.Context)`.
- [ ] `internal/service/income_scheduler_job_test.go` — unit tests.
- [ ] Polling interval configurable via `INCOME_SCHEDULER_INTERVAL` ENV VAR (default: `1h`).
- [ ] For each due position (`next_income_at <= NOW()` and `income_type != 'none'`):
  - Compute `amount_cents`: use `income_amount_cents` if fixed, or `ROUND(last_price_cents * quantity * income_rate_bps / 10000)` if rate-based, or sum of both if hybrid (ADR §9).
  - Create `position_income_events` row with `status = 'pending'`.
  - Advance `next_income_at += income_interval_days`.
- [ ] Graceful shutdown via `ctx.Done()`.
- [ ] Structured logging per position processed; errors non-fatal.
- [ ] Wiring in `cmd/api/main.go`.

### Out of scope

- Marking income as `received` (that is the tenant's action via the API — Task 3.14).
- Creating a `transactions` row when income is received (delegated to `InvestmentService.MarkIncomeReceived` — Task 3.6).
- Portfolio snapshot creation (Task 3.8).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                              | Purpose                                           |
| ------ | ------------------------------------------------- | ------------------------------------------------- |
| CREATE | `internal/service/income_scheduler_job.go`        | Polling goroutine                                 |
| CREATE | `internal/service/income_scheduler_job_test.go`   | Unit tests (mocked repos)                         |
| MODIFY | `cmd/api/main.go`                                 | Start goroutine at server startup                 |
| MODIFY | `internal/config/config.go`                       | Add `IncomeSchedulerInterval time.Duration` field |

### Struct

```go
type IncomeSchedulerJob struct {
    positionRepo  domain.PositionRepository
    incomeRepo    domain.PositionIncomeEventRepository
    logger        *slog.Logger
    interval      time.Duration // default: 1h
}

func NewIncomeSchedulerJob(
    positionRepo  domain.PositionRepository,
    incomeRepo    domain.PositionIncomeEventRepository,
    logger        *slog.Logger,
    interval      time.Duration,
) *IncomeSchedulerJob { ... }

// Run polls for due income events. Blocks until ctx is cancelled.
func (j *IncomeSchedulerJob) Run(ctx context.Context) error { ... }
```

### Amount calculation

```go
// computeIncomeAmount calculates the income amount in cents for one event.
// Fixed:     income_amount_cents (directly)
// Rate:      ROUND(last_price_cents * quantity * income_rate_bps / 10000)
// Hybrid:    fixed + rate (both non-nil)
func computeIncomeAmount(p domain.Position) int64 { ... }
```

### Error cases to handle

| Scenario                              | Action                                                  |
| ------------------------------------- | ------------------------------------------------------- |
| Income event creation fails           | Log error with `position_id`; continue to next position |
| `next_income_at` update fails         | Log error; the position will be processed again on next poll (idempotent by design) |
| `ctx` cancelled                       | Exit `Run` cleanly                                      |

---

## 5. Acceptance Criteria

- [ ] `IncomeSchedulerJob.Run` polls `PositionRepository.ListDueIncome` on the configured interval.
- [ ] Creates one `PositionIncomeEvent` with `status = 'pending'` per due position per poll.
- [ ] Advances `next_income_at` by exactly `income_interval_days` days after creating the event.
- [ ] Fixed, rate-based, and hybrid amount calculation all produce correct `int64` cents.
- [ ] A creation failure for one position does not stop processing of other positions.
- [ ] `Run` exits when `ctx` is cancelled.
- [ ] Unit tests cover: fixed income, rate income, hybrid income, context cancellation, error isolation.
- [ ] `make task-check` passes.
- [ ] `docs/ROADMAP.md` row 3.13 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-13 | —      | Task created (new) |
