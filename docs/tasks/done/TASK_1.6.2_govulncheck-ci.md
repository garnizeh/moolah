# Task 1.6.2 — `govulncheck` Passing in CI

> **Roadmap Ref:** Phase 1 — MVP › 1.6 Quality Gate
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-10
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Integrate `govulncheck` (the official Go vulnerability scanner from the Go team) into the CI security stage and the local `make security-check` target. The tool cross-references the project's module dependency graph against the Go vulnerability database (`vuln.go.dev`) and fails the build if any known vulnerability is reachable from the code under analysis.

---

## 2. Context & Motivation

`govulncheck` is Go's first-party, graph-aware vulnerability scanner. Unlike `go list -m -json all | grep` approaches, `govulncheck` identifies only *reachable* vulnerable symbols — reducing noise considerably. It is the recommended first line of supply-chain defense in Go projects. Running it in CI ensures vulnerable dependencies are caught before they reach production. See `docs/ARCHITECTURE.md` for the overall security posture.

---

## 3. Scope

### In scope

- [x] Install and run `govulncheck ./...` in the `security` CI job.
- [x] Add `govulncheck ./...` to the `make security-check` Makefile target.
- [x] Ensure the step fails the CI pipeline on any reachable vulnerability.

### Out of scope

- SBOM generation (deferred to Phase 5 hardening).
- Automatic dependency update PRs (out of scope for this task; consider Dependabot).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                         | Purpose                                   |
| ------ | ---------------------------- | ----------------------------------------- |
| MODIFY | `.github/workflows/ci.yml`   | Add `govulncheck` step in security job    |
| MODIFY | `Makefile`                   | Add `govulncheck` to `security-check`     |

### CI step (security job)

```yaml
- name: Run govulncheck
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...
```

### Makefile target

```makefile
## security-check: Run security scans (govulncheck and gosec)
security-check:
  @echo "🛡️ Running security scans..."
  @go install golang.org/x/vuln/cmd/govulncheck@latest
  @govulncheck ./...
```

### How `govulncheck` works

1. Builds the call graph of the Go program from source.
2. Queries `https://vuln.go.dev` (or a local mirror) for known CVEs against the module versions in `go.mod`.
3. Reports only CVEs where the vulnerable function is *reachable* from `main` or test code — not just "is this version in the graph".

---

## 5. Acceptance Criteria

- [x] `govulncheck ./...` exits 0 on a clean dependency tree.
- [x] CI `security` job fails if a reachable vulnerability is detected.
- [x] `make security-check` runs `govulncheck` as the first security step.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                       | Type     | Status  |
| ------------------------------------------------ | -------- | ------- |
| `golang.org/x/vuln` available on CI runner PATH  | External | ✅ done |
| Go module graph stable (`go.sum` committed)       | Upstream | ✅ done |

---

## 7. Testing Plan

### Validation

- Temporarily add a known-vulnerable module to `go.mod` and verify CI fails.
- Remove it, re-run, and confirm CI passes.
- Run `make security-check` locally to verify parity with CI.

---

## 8. Open Questions

| # | Question                                                         | Owner | Resolution                                             |
| - | ---------------------------------------------------------------- | ----- | ------------------------------------------------------ |
| 1 | Pin `govulncheck` to a specific version to avoid breaking changes? | —  | Currently installed via `@latest`; pin in a future hardening task if flakiness is observed. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-10 | Copilot| Task created; marked done — CI `security` job and Makefile already implement this |
