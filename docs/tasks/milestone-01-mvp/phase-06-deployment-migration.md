# Task 1.6.0 — Deployment & Migration

> **Roadmap Ref:** Phase 6 — Deployment & Migration
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-22
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Finalize the production readiness of Moolah and provide tools for migrating legacy data from spreadsheets into the new PostgreSQL engine.

---

## 2. Context & Motivation

The system is only valuable if it contains historical data. A dedicated migration path ensures continuity from 2023 onwards.

---

## 3. Scope

### In scope

- [ ] CLI tool for CSV import/parsing.
- [ ] Production-ready Dockerfile.
- [ ] CI/CD configuration (Linting, Tests).
- [ ] Automated reconciliation check scripts.

---

## 4. Technical Design

| Action   | Path                                      | Purpose                       |
| -------- | ----------------------------------------- | ----------------------------- |
| CREATE   | `cmd/moolah-cli/main.go`                  | Data import entry point       |
| CREATE   | `Dockerfile`                              | Multi-stage build             |
| CREATE   | `.github/workflows/ci.yml`                | Test automation               |

---

## 5. Acceptance Criteria

- [ ] CLI tool can parse the standard project spreadsheet CSV.
- [ ] CI pipeline passes on green code.
- [ ] Docker image is < 50MB (Alpine based).
