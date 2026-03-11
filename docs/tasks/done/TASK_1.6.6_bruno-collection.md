# Task 1.6.6 — Generate Bruno Collection

> **Roadmap Ref:** Phase 1 — MVP › 1.6 Quality Gate
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-10
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Create a Bruno collection for the Phase 1 API. This provides a git-friendly, collaborative environment for testing and demonstrating API endpoints.

---

## 2. Context & Motivation

Bruno is an open-source IDE for exploring and testing APIs that saves collections locally in the filesystem as Markup files, making it ideal for version control. It replaces the need for shared Postman collections.

---

## 3. Scope

### In scope

- [x] Initialize Bruno collection in `docs/bruno/`.
- [x] Define environment variables (Base URL, Auth Token).
- [x] Create requests for implemented endpoints:
  - Auth: Request OTP, Verify OTP, Refresh Token.
  - Health: Check status.
  - Accounts: List, Create.
  - Categories: List, Create.
  - Transactions: List, Create.
  - Admin: Tenants, Audit Logs.
- [x] Document how to use the collection in `README.md`.

---

## 5. Acceptance Criteria

- [x] Bruno collection files exist in `docs/bruno/`.
- [x] Requests are functional and use environment-based variables.
- [x] Collection is properly structured by domain (Auth, Accounts, etc.).
- [x] Pre-request scripts generate idempotency keys where needed.
- [x] Post-response scripts update token environment variables.

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-08 | —      | Task created from user request |
| 2026-03-10 | Copilot| Created collection, folders, requests and scripts |
| 2026-03-10 | Copilot| Marked as done |
