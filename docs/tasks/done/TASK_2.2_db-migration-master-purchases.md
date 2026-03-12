# Task 2.2 — DB Migration: `master_purchases` Table

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Infrastructure
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Add a Goose migration file that creates the `master_purchases` table in PostgreSQL. The table stores one row per credit card instalment purchase, acting as the "master record" from which the `InvoiceCloser` (Task 2.7) generates concrete `transactions` one instalment at a time.

---

## 2. Context & Motivation

Following the "Ghost Transaction" architecture (see `docs/ARCHITECTURE.md`), creating 12 physical `transactions` rows at purchase time would bloat the DB and make reconciliation complex. Instead, a single `master_purchases` row holds the intent. This migration is the DB prerequisite for Tasks 2.3 (sqlc queries) and 2.4 (repository).

The migration must follow the existing Goose file-naming convention and be embedded via `embed.FS` so it runs automatically on application startup (same pattern as Phase 1 migrations).

---

## 3. Scope

### In scope

- [x] New Goose UP migration: create `master_purchases` table with all required columns.
- [x] New Goose DOWN migration: drop `master_purchases` table.
- [x] `master_purchase_status` enum type (`open`, `closed`).
- [x] Proper FK constraints to `tenants`, `accounts`, `categories`, `users`.
- [x] Index on `(tenant_id, account_id)` for efficient `ListByAccount` queries.
- [x] Index on `(tenant_id, status, closing_day)` for efficient `ListPendingClose` queries.

### Out of scope

- Changes to `transactions` table (already has `master_purchase_id` column from Phase 1).
- sqlc code generation (Task 2.3).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                                        | Purpose                              |
| ------ | --------------------------------------------------------------------------- | ------------------------------------ |
| CREATE | `internal/platform/db/migrations/XXXXXX_create_master_purchases.sql`        | Goose migration (UP + DOWN)          |

> Replace `XXXXXX` with the next sequential Goose version number after the last Phase 1 migration.

### DDL (reference)

```sql
-- +goose Up
CREATE TYPE master_purchase_status AS ENUM ('open', 'closed');

CREATE TABLE master_purchases (
    id                    VARCHAR(26)              NOT NULL,
    tenant_id             VARCHAR(26)              NOT NULL,
    account_id            VARCHAR(26)              NOT NULL,
    category_id           VARCHAR(26)              NOT NULL,
    user_id               VARCHAR(26)              NOT NULL,
    description           TEXT                     NOT NULL,
    status                master_purchase_status   NOT NULL DEFAULT 'open',
    total_amount_cents    BIGINT                   NOT NULL CHECK (total_amount_cents > 0),
    installment_count     SMALLINT                 NOT NULL CHECK (installment_count BETWEEN 2 AND 48),
    paid_installments     SMALLINT                 NOT NULL DEFAULT 0 CHECK (paid_installments >= 0),
    closing_day           SMALLINT                 NOT NULL CHECK (closing_day BETWEEN 1 AND 28),
    first_installment_date DATE                    NOT NULL,
    created_at            TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT pk_master_purchases           PRIMARY KEY (id),
    CONSTRAINT fk_mp_tenant                  FOREIGN KEY (tenant_id)   REFERENCES tenants(id),
    CONSTRAINT fk_mp_account                 FOREIGN KEY (account_id)  REFERENCES accounts(id),
    CONSTRAINT fk_mp_category                FOREIGN KEY (category_id) REFERENCES categories(id),
    CONSTRAINT fk_mp_user                    FOREIGN KEY (user_id)     REFERENCES users(id),
    CONSTRAINT chk_paid_lte_total            CHECK (paid_installments <= installment_count)
);

CREATE INDEX idx_mp_tenant_account ON master_purchases (tenant_id, account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_mp_pending_close  ON master_purchases (tenant_id, status, closing_day) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_mp_pending_close;
DROP INDEX IF EXISTS idx_mp_tenant_account;
DROP TABLE IF EXISTS master_purchases;
DROP TYPE IF EXISTS master_purchase_status;
```

### Error cases to handle

| Scenario                               | pg Error Code | Handling                       |
| -------------------------------------- | ------------- | ------------------------------ |
| Duplicate primary key                  | `23505`       | Should not occur (ULID unique) |
| FK violation (account/category/user)   | `23503`       | → `domain.ErrNotFound`         |
| `CHECK` violation (cents, installments)| `23514`       | → `domain.ErrInvalidInput`     |

---

## 5. Acceptance Criteria

- [x] Migration file follows the existing Goose sequential numbering convention.
- [x] `-- +goose Up` and `-- +goose Down` markers are present and correct.
- [x] `master_purchase_status` enum is created before the table and dropped after.
- [x] All FK constraints reference existing Phase 1 tables.
- [x] `CHECK` constraint ensures `paid_installments <= installment_count`.
- [x] `closing_day` is constrained to 1–28.
- [x] Both indexes are created in UP and dropped in DOWN.
- [x] `goose up` and `goose down` run cleanly against a local Postgres instance (`make db-migrate`).
- [x] `docs/ROADMAP.md` row 2.2 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status       |
| ---------------------------------------------- | -------- | ------------ |
| Task 2.1 — `domain/master_purchase.go`         | Upstream | ✅ done      |
| All Phase 1 migrations applied                 | Upstream | ✅ done      |
| Task 2.3 — sqlc queries (consumer)             | Downstream | 🔵 backlog |
| Task 2.4 — Repository impl (consumer)          | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — migrations are verified via integration tests.

### Integration tests (`//go:build integration`)

- **Covered by:** Task 2.4 repository integration tests (testcontainers-go spins up Postgres and runs all migrations, including this one).
- **Verify:** table exists, indexes exist, FK constraints reject orphaned references.

---

## 8. Open Questions

| # | Question                                                                             | Owner | Resolution |
| - | ------------------------------------------------------------------------------------ | ----- | ---------- |
| 1 | Should `first_installment_date` be `DATE` or `TIMESTAMPTZ`?                          | —     | `DATE` — no time-of-day semantics needed for instalment due dates. |
| 2 | Do we add `updated_at` trigger (like Phase 1 tables), or update manually via sqlc?   | —     | Follow existing pattern: update `updated_at` via `sqlc` UPDATE queries. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
