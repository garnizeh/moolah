# Task 3.1 — ADR: Investment Data Model (Positions, Assets, Snapshots)

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Architecture Decision
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Produce a formal Architecture Decision Record (ADR) that defines the data model for investment portfolio tracking. The ADR must cover how assets, positions, and monthly portfolio snapshots are structured — addressing multi-currency, multi-asset-class concerns — so that every subsequent Phase 3 task has a clear, agreed-upon contract before implementation begins.

---

## 2. Context & Motivation

Phase 2 introduced the "lean record" pattern for credit card installments (Master Purchase → physical rows only at invoice-close time). Phase 3 extends the system into read-heavy investment tracking where accuracy of current position value matters more than write throughput. Before writing a single migration or repository, the team must agree on:

- How assets are catalogued (ticker, ISIN, type: stock/bond/fund/crypto/real-estate).
- How positions are recorded (cost basis, current quantity, last known price).
- Whether price history is stored in-app or fetched live from an external source.
- How portfolio snapshots are generated and retained for charting MoM performance.
- Tenant isolation strategy for all investment tables.

Without this ADR, subsequent tasks risk conflicting assumptions about schema shape and query patterns.

---

## 3. Scope

### In scope

- [ ] ADR document saved to `docs/` (e.g., `docs/ADR-003-investment-data-model.md`).
- [ ] Entity definitions: `Asset`, `Position`, `PortfolioSnapshot`.
- [ ] Decision on price history storage: in-app vs. external feed (Phase 3 MVP = manual / static).
- [ ] Decision on currency conversion: manual rate table for MVP; hook defined in Task 3.10.
- [ ] Diagram (Mermaid ERD) illustrating entity relationships.
- [ ] Tenant isolation approach documented (same `tenant_id` column pattern as Phase 1 & 2).
- [ ] Soft-delete policy for positions and assets.

### Out of scope

- External price feed integration (deferred to Phase 4/5 or a future ADR).
- Real-time WebSocket price streaming (deferred per roadmap).
- Mobile client data model concerns.

---

## 4. Technical Design

### ADR Structure

The ADR follows the format: **Context → Decision → Consequences**.

```
docs/ADR-003-investment-data-model.md
```

### Proposed Entity Map

```
Tenant (1) ──< Position (N)
Asset  (1) ──< Position (N)
Tenant (1) ──< PortfolioSnapshot (N)
```

#### `assets` — global catalogue (no `tenant_id`; read-only reference data)

| Column      | Type         | Notes                                        |
| ----------- | ------------ | -------------------------------------------- |
| `id`        | `VARCHAR(26)` | ULID                                        |
| `ticker`    | `VARCHAR(20)` | e.g. `AAPL`, `BTC`                         |
| `name`      | `VARCHAR(200)` | Human-readable name                        |
| `asset_type`| `asset_type` enum | `stock`, `bond`, `fund`, `crypto`, `real_estate` |
| `currency`  | `CHAR(3)`    | ISO 4217 base currency                       |
| `created_at`| `TIMESTAMPTZ` |                                             |

#### `positions` — per-tenant holdings

| Column              | Type          | Notes                                    |
| ------------------- | ------------- | ---------------------------------------- |
| `id`                | `VARCHAR(26)` | ULID                                     |
| `tenant_id`         | `VARCHAR(26)` | FK → tenants                             |
| `asset_id`          | `VARCHAR(26)` | FK → assets                              |
| `account_id`        | `VARCHAR(26)` | FK → accounts (investment type)          |
| `quantity`          | `NUMERIC(18,8)` | Fractional share support               |
| `avg_cost_cents`    | `BIGINT`      | Average cost per unit in cents           |
| `last_price_cents`  | `BIGINT`      | Last known price in cents (manual entry) |
| `currency`          | `CHAR(3)`     | Position currency (may differ from asset)|
| `purchased_at`      | `TIMESTAMPTZ` | Date of initial purchase                 |
| `created_at`        | `TIMESTAMPTZ` |                                          |
| `updated_at`        | `TIMESTAMPTZ` |                                          |
| `deleted_at`        | `TIMESTAMPTZ` | Soft delete                              |

#### `portfolio_snapshots` — monthly point-in-time value

| Column            | Type          | Notes                                    |
| ----------------- | ------------- | ---------------------------------------- |
| `id`              | `VARCHAR(26)` | ULID                                     |
| `tenant_id`       | `VARCHAR(26)` | FK → tenants                             |
| `snapshot_date`   | `DATE`        | First day of the month snapshotted       |
| `total_value_cents` | `BIGINT`   | Sum of all position values at snapshot date|
| `currency`        | `CHAR(3)`     | Reference currency for total             |
| `details`         | `JSONB`       | Per-asset breakdown for charting         |
| `created_at`      | `TIMESTAMPTZ` |                                          |

### Files to create / modify

| Action | Path                           | Purpose                      |
| ------ | ------------------------------ | ---------------------------- |
| CREATE | `docs/ADR-003-investment-data-model.md` | Formal ADR document |

---

## 5. Acceptance Criteria

- [ ] `docs/ADR-003-investment-data-model.md` is committed and contains all required sections.
- [ ] ERD Mermaid diagram is valid and renders in GitHub Markdown.
- [ ] All three entities (`assets`, `positions`, `portfolio_snapshots`) are fully specified with column names, types, and constraints.
- [ ] Tenant isolation approach is explicitly documented.
- [ ] Soft-delete policy is stated.
- [ ] Currency/price-source decision is documented.
- [ ] `docs/ROADMAP.md` row 3.1 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                   | Type     | Status       |
| -------------------------------------------- | -------- | ------------ |
| Phase 2 fully complete (schema stabilised)   | Upstream | ✅ done      |
| `docs/ARCHITECTURE.md` reviewed for patterns | Upstream | ✅ done      |

---

## 7. Testing Plan

### Unit tests

N/A — this task produces documentation only.

### Integration tests

N/A — no code changes.

---

## 8. Open Questions

| # | Question                                                                                      | Owner | Resolution |
| - | --------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `assets` be global (shared across tenants) or per-tenant?                             | —     | —          |
| 2 | What is the snapshotting trigger — cron job, manual API call, or event from invoice closer?  | —     | —          |
| 3 | Should `positions` link to an `accounts` row of type `investment`, or be standalone?          | —     | —          |
| 4 | Do we need a `price_history` table, or is last-known price sufficient for MVP?               | —     | —          |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
