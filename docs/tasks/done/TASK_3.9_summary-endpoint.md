# TASK 3.9: Position Domain Refactor & Summary Infrastructure

## Status
- **Status:** ✅ done
- **Last Updated:** 2024-03-24

## Description
Refactor the Investment/Position domain to support decimal quantities, consolidated income events, and robust portfolio snapshots. Ensure all repository mocks are updated to match the new interface signatures.

## Acceptance Criteria
- [x] Refactor `Position` to use `decimal.Decimal` for `Quantity`.
- [x] Consolidate `PositionIncomeEvent` interfaces within `PositionRepository` where applicable.
- [x] Implement manual mocks in `internal/testutil/mocks/domain.go` for all core repositories.
- [x] Ensure all service tests in `internal/service/` pass.
- [x] `make task-check` passes successfully.

## Technical Details
- Used `github.com/shopspring/decimal` for finance-accurate quantities.
- Updated `InvestmentService`, `AdminService`, `AuthService`, and others to match new mock signatures.
- Maintained interface assertions `var _ domain.XXXRepository = (*XXXRepository)(nil)`.

## Change Log
- **2024-03-24:** Initial refactor and mock reconstruction. Fixed widespread test breakage.
- **2024-03-24:** Finalized all repository mocks (Admin, Auth, Tenant, User, etc.) to unblock full service layer testing.
