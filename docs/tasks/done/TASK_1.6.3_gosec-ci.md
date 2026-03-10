# Task 1.6.3 — `gosec` Passing in CI

> **Roadmap Ref:** Phase 1 — MVP › 1.6 Quality Gate
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-10
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Integrate `gosec` (Go Security Checker) into the CI security stage and the local `make security-check` target. `gosec` performs static analysis of Go source code to detect common security issues such as SQL injection, hardcoded credentials, insecure crypto usage, insufficient integer overflow guards, and misuse of `os/exec`. The CI pipeline must fail if any `gosec` issue is found.

---

## 2. Context & Motivation

`gosec` maps directly to several OWASP Top 10 categories: injection (G201/G202), insecure cryptography (G401–G501), hardcoded secrets (G101), and misconfiguration (G304, G401). Running it in CI enforces the security requirements specified in the project's `operationalSafety` policy. It complements `govulncheck` (dependency-level) with source-level static analysis. See `docs/ARCHITECTURE.md` for security posture notes.

---

## 3. Scope

### In scope

- [x] Install and run `gosec ./...` in the `security` CI job.
- [x] Add `gosec ./...` to the `make security-check` Makefile target.
- [x] Ensure the step fails CI on any reported issue (no suppressions without justification).
- [x] Exclude auto-generated code (`/api/`, `/internal/platform/db/sqlc/`) from gosec scanning via `#nosec` annotations or `.gosec` config where appropriate.

### Out of scope

- Custom `gosec` rule configuration file (use default ruleset for Phase 1).
- SARIF report upload to GitHub Security tab (deferred to Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                         | Purpose                                    |
| ------ | ---------------------------- | ------------------------------------------ |
| MODIFY | `.github/workflows/ci.yml`   | Add `gosec` step in security job           |
| MODIFY | `Makefile`                   | Add `gosec` to `security-check` target     |

### CI step (security job)

```yaml
- name: Run gosec
  run: |
    go install github.com/securego/gosec/v2/cmd/gosec@v2.24.7
    "$(go env GOPATH)/bin/gosec" ./...
```

> **Note:** Pin `gosec` to a specific version (e.g. `v2.24.7`) to prevent non-reproducible CI failures from upstream breaking changes.

### Makefile target

```makefile
## security-check: Run security scans (govulncheck and gosec)
security-check:
  @echo "🛡️ Running security scans..."
  @go install golang.org/x/vuln/cmd/govulncheck@latest
  @govulncheck ./...
  @go install github.com/securego/gosec/v2/cmd/gosec@latest
  @gosec ./...
```

### Common issues caught by gosec in this codebase

| Rule   | Description                                  | Location pattern                         |
| ------ | -------------------------------------------- | ---------------------------------------- |
| G115   | Integer overflow risk (`int` → `int64`)      | `strconv.Atoi` results used as `int64`   |
| G401   | Weak crypto (MD5/SHA1)                       | Hash functions in OTP or token logic     |
| G501   | Import of `crypto/md5` or `crypto/sha1`      | Any import of weak hash packages         |
| G304   | File path provided as taint input            | `os.Open` / `os.ReadFile` with user data |
| G101   | Hardcoded credentials detected               | Any string literal resembling a secret   |

---

## 5. Acceptance Criteria

- [x] `gosec ./...` exits 0 on a clean codebase.
- [x] CI `security` job fails if any gosec issue is detected.
- [x] `make security-check` runs `gosec` after `govulncheck`.
- [x] No unresolved `#nosec` suppressions without an explanatory comment.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                         | Type     | Status  |
| -------------------------------------------------- | -------- | ------- |
| `github.com/securego/gosec/v2` available on runner | External | ✅ done |
| All Phase 1 source files free of critical findings | Upstream | ✅ done |

---

## 7. Testing Plan

### Validation

- Introduce a deliberate `G101` (hardcoded credential) in a test file protected by `//go:build ignore`, verify CI fails, then remove it.
- Run `make security-check` locally and confirm output matches CI.

### Suppression policy

If a `gosec` finding is a known false positive, it must be suppressed with `// #nosec GXX -- <reason>` inline, and documented in this task's change log. Do not suppress without justification.

---

## 8. Open Questions

| # | Question                                                           | Owner | Resolution                                                 |
| - | ------------------------------------------------------------------ | ----- | ---------------------------------------------------------- |
| 1 | Should we upload SARIF results to GitHub Security Advisories tab?  | —     | Deferred to Phase 5 observability work.                    |
| 2 | Should `gosec` be pinned or use `@latest`?                         | —     | Pin to `v2.24.7` in CI; `@latest` in Makefile for local dev. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-10 | Copilot| Task created; marked done — CI `security` job and Makefile already implement this |
