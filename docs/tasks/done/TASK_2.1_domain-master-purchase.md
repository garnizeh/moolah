# Task 2.1 — Domain: MasterPurchase Entity + Repository Interface

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the `MasterPurchase` domain entity and the `MasterPurchaseRepository` interface in `internal/domain/master_purchase.go`. This is the foundational "Ghost Transaction" model: a single record that stores the original credit card purchase intent and generates concrete installment `transactions` only at invoice-close time, keeping the database lean.

---

## 2. Context & Motivation

The "Master Purchase" pattern (see `docs/ARCHITECTURE.md` and `docs/ROADMAP.md#2.1`) avoids materialising N rows immediately when a user registers a 12× instalment purchase. Instead, a single `master_purchases` row holds the intent. The `InvoiceCloser` service (Task 2.7) reads pending master purchases each billing cycle and writes only the instalment `transaction` for that cycle.

This task delivers **only** the domain types and repository interface. Concrete implementation follows in Task 2.4 (repository) and Task 2.5 (service).

---

## 3. Scope

### In scope

- [x] `MasterPurchase` struct with all required fields (see §4).
- [x] `MasterPurchaseStatus` type + constants (`open`, `closed`).
- [x] `CreateMasterPurchaseInput` value object with `go-playground/validator` tags.
- [x] `UpdateMasterPurchaseInput` value object (partial update).
- [x] `ProjectedInstalment` value object — runtime-computed, never persisted.
- [x] `MasterPurchaseRepository` interface including `ListPendingClose`.
- [x] `MasterPurchaseService` interface stub (consumed by handler via DI).

### Out of scope

- Concrete repository implementation (Task 2.4).
- Business logic / instalment projection (Task 2.5).
- Goose migration (Task 2.2).
- sqlc queries (Task 2.3).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                    | Purpose                                      |
| ------ | --------------------------------------- | -------------------------------------------- |
| CREATE | `internal/domain/master_purchase.go`    | Entity, value objects, repository interface  |

### Key interfaces / types

```go
// MasterPurchaseStatus represents the lifecycle state of a master purchase.
type MasterPurchaseStatus string

const (
    MasterPurchaseStatusOpen   MasterPurchaseStatus = "open"   // instalments still pending
    MasterPurchaseStatusClosed MasterPurchaseStatus = "closed" // all instalments materialised
)

// MasterPurchase is a credit-card purchase that generates instalment transactions
// one at a time at each invoice-close cycle, rather than all at once.
// All monetary values are stored in cents (int64).
type MasterPurchase struct {
    FirstInstallmentDate time.Time            `json:"first_installment_date"`
    CreatedAt            time.Time            `json:"created_at"`
    UpdatedAt            time.Time            `json:"updated_at"`
    DeletedAt            *time.Time           `json:"deleted_at,omitempty"`
    ID                   string               `json:"id"`
    TenantID             string               `json:"tenant_id"`
    AccountID            string               `json:"account_id"`   // must be AccountTypeCreditCard
    CategoryID           string               `json:"category_id"`
    UserID               string               `json:"user_id"`
    Description          string               `json:"description"`
    Status               MasterPurchaseStatus `json:"status"`
    TotalAmountCents     int64                `json:"total_amount_cents"`
    InstallmentCount     int32                `json:"installment_count"`
    PaidInstallments     int32                `json:"paid_installments"` // incremented each close cycle
    ClosingDay           int32                `json:"closing_day"`       // day-of-month invoice closes (1–28)
}

// CreateMasterPurchaseInput is the value object for creating a new master purchase.
type CreateMasterPurchaseInput struct {
    FirstInstallmentDate time.Time `validate:"required"`
    AccountID            string    `validate:"required"`
    CategoryID           string    `validate:"required"`
    UserID               string    `validate:"required"`
    Description          string    `validate:"required,min=1,max=255"`
    TotalAmountCents     int64     `validate:"required,gt=0"`
    InstallmentCount     int32     `validate:"required,min=2,max=48"`
    ClosingDay           int32     `validate:"required,min=1,max=28"`
}

// UpdateMasterPurchaseInput allows patching description or category only.
type UpdateMasterPurchaseInput struct {
    CategoryID  *string `validate:"omitempty"`
    Description *string `validate:"omitempty,min=1,max=255"`
}

// ProjectedInstallment is a runtime-computed instalment — never stored in the DB directly.
type ProjectedInstallment struct {
    InstallmentNumber int32     `json:"installment_number"`
    DueDate           time.Time `json:"due_date"`
    AmountCents       int64     `json:"amount_cents"` // last instalment absorbs remainder (Task 2.9)
}

// MasterPurchaseRepository defines persistence operations for master purchases.
// All queries enforce tenant isolation via tenant_id.
type MasterPurchaseRepository interface {
    Create(ctx context.Context, tenantID string, input CreateMasterPurchaseInput) (*MasterPurchase, error)
    GetByID(ctx context.Context, tenantID, id string) (*MasterPurchase, error)
    ListByTenant(ctx context.Context, tenantID string) ([]MasterPurchase, error)
    ListByAccount(ctx context.Context, tenantID, accountID string) ([]MasterPurchase, error)
    // ListPendingClose returns open master purchases whose next instalment falls on or before cutoffDate.
    ListPendingClose(ctx context.Context, tenantID string, cutoffDate time.Time) ([]MasterPurchase, error)
    Update(ctx context.Context, tenantID, id string, input UpdateMasterPurchaseInput) (*MasterPurchase, error)
    // IncrementPaidInstallments atomically advances PaidInstallments and, if all paid, sets status=closed.
    IncrementPaidInstallments(ctx context.Context, tenantID, id string) error
    Delete(ctx context.Context, tenantID, id string) error
}

// MasterPurchaseService defines the business-logic contract for master purchases.
type MasterPurchaseService interface {
    Create(ctx context.Context, tenantID string, input CreateMasterPurchaseInput) (*MasterPurchase, error)
    GetByID(ctx context.Context, tenantID, id string) (*MasterPurchase, error)
    ListByTenant(ctx context.Context, tenantID string) ([]MasterPurchase, error)
    ListByAccount(ctx context.Context, tenantID, accountID string) ([]MasterPurchase, error)
    // ProjectInstallments computes the full instalment schedule without persisting anything.
    ProjectInstallments(mp *MasterPurchase) []ProjectedInstallment
    Update(ctx context.Context, tenantID, id string, input UpdateMasterPurchaseInput) (*MasterPurchase, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### Error cases to handle

| Scenario                              | Sentinel Error              | HTTP Status |
| ------------------------------------- | --------------------------- | ----------- |
| Master purchase not found             | `domain.ErrNotFound`        | `404`       |
| Account is not credit_card type       | `domain.ErrInvalidInput`    | `422`       |
| Attempt to delete a closed purchase   | `domain.ErrForbidden`       | `403`       |
| Tenant mismatch / unauthorized        | `domain.ErrForbidden`       | `403`       |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] `MasterPurchaseRepository` and `MasterPurchaseService` interfaces are defined in `internal/domain/`.
- [ ] `ProjectedInstallment` is a value object (no DB column) with correct `json` tags.
- [ ] `ClosingDay` is validated as `min=1,max=28` (avoids month-end ambiguity).
- [ ] No float types used for monetary values.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.1 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                              | Type       | Status       |
| --------------------------------------- | ---------- | ------------ |
| Phase 1 fully complete (quality gate)   | Upstream   | ✅ done      |
| `domain/errors.go` — sentinel errors    | Upstream   | ✅ done      |
| Task 2.2 — Goose migration (consumer)   | Downstream | 🔵 backlog   |
| Task 2.3 — sqlc queries (consumer)      | Downstream | 🔵 backlog   |
| Task 2.4 — Repository impl (consumer)   | Downstream | 🔵 backlog   |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/domain/master_purchase_test.go`
- **Cases:**
  - Validate `CreateMasterPurchaseInput` struct tags reject invalid `ClosingDay` (0, 29).
  - Validate `InstallmentCount` rejects values < 2 and > 48.
  - Validate `TotalAmountCents` rejects zero and negative.

### Integration tests (`//go:build integration`)

N/A — domain layer has no DB dependency. Integration coverage starts at Task 2.4.

---

## 8. Open Questions

| # | Question                                                                        | Owner | Resolution |
| - | ------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `ClosingDay=28` be the safe ceiling, or allow 29/30/31 with adjustments? | —     | Cap at 28 (avoids February edge-case) — confirmed in roadmap note. |
| 2 | Should `MasterPurchase` support a `currency` field or always inherit account?   | —     | Inherit from `Account.Currency` (no duplication). |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
