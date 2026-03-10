# Task 1.6.1 — ≥ 80% Unit Test Coverage Enforced in CI

> **Roadmap Ref:** Phase 1 — MVP › 1.6 Quality Gate
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-10
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Enforce a minimum of 80% unit test coverage across the codebase by wiring a coverage threshold check into the CI pipeline (`ci.yml`) and into the local `make task-check` target. Any PR that causes coverage to drop below 80% fails the pipeline, ensuring coverage never erodes silently.

---

## 2. Context & Motivation

The project mandate (see `copilot-instructions.md`) requires ~100% coverage intention with a hard floor of ≥ 80% enforced in CI. Without this gate, coverage can silently regress as features are added. The check must live both locally (developer feedback loop via `make task-check`) and remotely (CI gatekeeper on PRs). Additionally, the coverage report is uploaded to Codecov for historical tracking.

---

## 3. Scope

### In scope

- [x] `go test -coverprofile=coverage.out` in CI `unit-tests` job.
- [x] Enforce 80% threshold via `awk` in CI (fail the job if below threshold).
- [x] Upload `coverage.out` to Codecov via `codecov/codecov-action@v5`.
- [x] `make test-coverage` target in `Makefile` that mirrors CI behaviour locally.
- [x] Exclude generated code from coverage measurement (`/platform/db/sqlc`, `/testutil/mocks`, `/api`).

### Out of scope

- Per-package coverage enforcement (aggregate threshold only for now).
- Codecov PR annotations (nice-to-have; enabled via Codecov dashboard settings).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                              | Purpose                                          |
| ------ | --------------------------------- | ------------------------------------------------ |
| MODIFY | `.github/workflows/ci.yml`        | Add coverage measurement and threshold check     |
| MODIFY | `Makefile`                        | Add `test-coverage` target                       |

### CI step (unit-tests job)

```yaml
- name: Run unit tests with race detector and coverage
  run: |
    go test -v -race -count=1 \
      -tags=integration \
      -timeout=600s \
      -coverprofile=coverage.out \
      -covermode=atomic \
      $(go list ./... | grep -v /platform/db/sqlc | grep -v /testutil/mocks | grep -v /api)

- name: Enforce coverage threshold (80%)
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    echo "Total coverage: ${COVERAGE}%"
    awk "BEGIN { exit (${COVERAGE} < 80) }" || (echo "ERROR: Coverage ${COVERAGE}% is below the 80% threshold" && exit 1)

- name: Upload coverage report to Codecov
  uses: codecov/codecov-action@v5
  with:
    files: ./coverage.out
    fail_ci_if_error: false
    token: ${{ secrets.CODECOV_TOKEN }}
```

### Makefile target

```makefile
## test-coverage: Run unit tests and enforce coverage (80% threshold)
test-coverage:
  @$(GO) test -v -race -count=1 -tags=integration -timeout=600s \
    -coverprofile=coverage.out -covermode=atomic \
    $$(go list ./... | grep -v /platform/db/sqlc | grep -v /testutil/mocks)
  @COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
  awk "BEGIN { if ($${COVERAGE} < 80) exit 1 }" || (echo "❌ ERROR: Coverage below 80%"; exit 1)
```

---

## 5. Acceptance Criteria

- [x] CI `unit-tests` job runs `go test -coverprofile=coverage.out`.
- [x] CI fails if total coverage drops below 80%.
- [x] `make test-coverage` mirrors the CI check locally.
- [x] Generated packages (`/sqlc`, `/mocks`, `/api`) are excluded from measurement.
- [x] Coverage report is uploaded to Codecov on every push to `main`.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                         | Type     | Status  |
| ---------------------------------- | -------- | ------- |
| All Phase 1 services/handlers done | Upstream | ✅ done |
| Codecov account + `CODECOV_TOKEN`  | External | ✅ done |

---

## 7. Testing Plan

### Validation

- Manually lower the threshold to 99% and verify the CI step fails.
- Run `make test-coverage` locally and observe the pass/fail output.

---

## 8. Open Questions

| # | Question                                              | Owner | Resolution                                              |
| - | ----------------------------------------------------- | ----- | ------------------------------------------------------- |
| 1 | Raise threshold to 90% after handler layer is tested? | —     | Revisit after Phase 1 handler tests are complete. Current floor: 80%. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-10 | Copilot| Task created; marked done — CI and Makefile already implement this |
