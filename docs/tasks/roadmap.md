# Moolah MVP Roadmap: Spreadsheet Replacement Phase

This roadmap outlines the tasks required to transition from manual spreadsheets to the Moolah platform, as defined in [RFC 003](file:///home/userone/Code/github/garnizeh/moolah/docs/design/003-moolah-product.md).

## Phase 1: Technical Foundation & Standards
The goal is to establish the engineering high-ground before building features.
- [x] **Project Scaffolding**: Setup Go modules, `docker-compose` for local PostgreSQL, and basic folder structure.
- [x] **Observability**: Implement `log/slog` structured logging and Request ID (ULID) middleware.
- [x] **OpenTelemetry**: Basic trace instrumentation for HTTP handlers.
- [x] **Database Plumbing**: Setup `pressly/goose` for migrations and `sqlc` for type-safe query generation.
- [x] **Shared Types**: Internal representation of monetary values (integers in cents).

## Phase 2: Domain Modeling & Core Registry
Establishing the entities that govern the system's extensibility.
- [ ] **Currency Engine**: Implement CURRENCY table and logic for handling multiple decimal precisions (BRL=2, BTC=8, etc.).
- [ ] **Entity Registry**: Implement ENTITY table for family members and cost centers with JSONB support.
- [ ] **Account Management**: CRUD for Accounts/Wallets/Credit Cards linked to entities.

## Phase 3: The Ledger Engine (Transactions)
The source of truth for all cash flows.
- [ ] **Transactional Schema**: Implement TRANSACTION table with `expected_value` and `actual_paid` fields.
- [ ] **Category Hierarchy**: Multi-level categorization for expenses and income.
- [ ] **Status Calculation**: Logic for "PENDING", "PARTIAL", "PAID", and "OVERDUE" states.
- [ ] **Metadata Support**: Ensure the `transaction.metadata` (JSONB) can store the original legacy parsing strings.

## Phase 4: Obligations & Installments
Modeling long-term debt as a contract rather than fragmented entries.
- [ ] **Long-Term Obligation Model**: Implement the `LONG_TERM_OBLIGATION` contract table.
- [ ] **Installment Generator**: Logic to auto-populate future months based on a contract (e.g., 36 installments).

## Phase 5: Hypermedia UI (HTMX/Tailwind)
Building the "WOW" interface for internal testing and data entry.
- [ ] **Component Library**: Setup Tailwind CSS with a premium dark-mode theme.
- [ ] **Navigation & Layout**: Persistent sidebar/shell using HTMX for page transitions.
- [ ] **Global Dashboard**: Net Worth summary and multi-account balance cards.
- [ ] **Expenditure Entry Flow**: Quick-entry forms for manual logs with real-time validation.
- [ ] **Variance Analysis View**: A dedicated dashboard screen to identify the "Pendente" gap between planned and paid values for the month.

## Phase 6: Deployment & Migration
- [ ] **Local CLI**: Tooling to import legacy CSV data into the new schema.
- [ ] **Production Readiness**: Dockerfile and basic CI/CD pipeline (GitHub Actions).
