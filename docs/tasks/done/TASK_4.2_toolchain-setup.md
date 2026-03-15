# Task 4.2 — Toolchain Setup: Templ + Tailwind CLI + Web Binary

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Toolchain
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Bootstrap the entire front-end build pipeline inside the existing Go monorepo: add `a-h/templ` to `go.mod`, create the `cmd/web/` server entry point, embed static assets via `embed.FS`, integrate `templ generate` and the Tailwind CSS CLI into the `Makefile`, and wire the generation steps into the GitHub Actions CI pipeline. After this task, `make web` must produce a working (empty-shell) web binary and all subsequent UI tasks can build on this foundation.

---

## 2. Context & Motivation

The API binary lives in `cmd/api/`. The web (UI) binary will live in `cmd/web/` and serve HTML pages to browsers. Keeping them as separate Go binaries allows independent deployment and clean import boundaries. Templ compiles `.templ` files to `.go` files; those generated files must be tracked in the CI pipeline (like `sqlc generate`). Tailwind CLI processes `web/static/css/app.css` and outputs `web/static/css/app.min.css` — the only CSS file served to browsers.

**Depends on:** Task 4.1 (ADR approved, directory layout decided).

---

## 3. Scope

### In scope

- [x] Add `github.com/a-h/templ` to `go.mod` and `vendor/`.
- [x] Create `cmd/web/main.go` — web server entry point (HTTP, graceful shutdown, config reuse).
- [x] Create `internal/ui/` directory structure (empty placeholder `.templ` files as stubs).
- [x] Create `web/static/css/app.css` — Tailwind CSS v4 entrypoint file.
- [x] Create `web/static/js/` — vendor HTMX 2 and Alpine.js 3 minified bundles.
- [x] Create `web/embed.go` — `//go:embed` directive for `web/static/` subtree.
- [x] Update `Makefile`:
  - `make templ` — runs `templ generate ./...`
  - `make tailwind` — runs `tailwindcss -i web/static/css/app.css -o web/static/css/app.min.css --minify`
  - `make web-build` — runs `templ + tailwind + go build ./cmd/web/`
  - `make web` — runs the web binary locally (dev mode with `--watch` flag awareness)
  - `make generate` — must call both `sqlc generate` and `templ generate`
- [x] Update `.github/workflows/ci.yml`:
  - Install `templ` CLI tool in the generate step.
  - Install Tailwind CLI binary in the generate step.
  - Add `make web-build` to the build job.
- [x] Create `cmd/web/main_test.go` — smoke test that the server starts and serves a 200 on `/`.
- [x] Update `docker-compose.yml` — add a `web` service or document how to run it separately.

### Out of scope

- Actual page content (Tasks 4.3–4.9).
- Authentication wiring to the web server (Task 4.6).
- Database connection in the web binary (web server calls the API; it does not access the DB directly).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                               | Purpose                                              |
| ------ | ---------------------------------- | ---------------------------------------------------- |
| MODIFY | `go.mod`                           | Add `github.com/a-h/templ`                           |
| CREATE | `cmd/web/main.go`                  | Web server entry point                               |
| CREATE | `cmd/web/main_test.go`             | Startup smoke test                                   |
| CREATE | `internal/ui/layout/.gitkeep`      | Placeholder for layout templates                     |
| CREATE | `internal/ui/components/.gitkeep`  | Placeholder for component templates                  |
| CREATE | `internal/ui/pages/.gitkeep`       | Placeholder for page templates                       |
| CREATE | `web/static/css/app.css`           | Tailwind v4 entrypoint (imports + custom tokens)     |
| CREATE | `web/static/js/htmx.min.js`        | Vendored HTMX 2 bundle                               |
| CREATE | `web/static/js/alpine.min.js`      | Vendored Alpine.js 3 bundle                          |
| CREATE | `web/embed.go`                     | `//go:embed static/**` for production builds         |
| MODIFY | `Makefile`                         | Add `templ`, `tailwind`, `web-build`, `web` targets  |
| MODIFY | `.github/workflows/ci.yml`         | Add templ + tailwind CLI install + web-build job     |
| MODIFY | `docker-compose.yml`               | Document / add web service                           |

### `cmd/web/main.go` skeleton

```go
package main

import (
    "context"
    "embed"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/garnizeh/moolah/internal/config"
    "github.com/garnizeh/moolah/web"
)

func main() {
    cfg := config.Load()

    mux := http.NewServeMux()

    // Static assets
    mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(web.StaticFS))))

    // Health
    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    srv := &http.Server{
        Addr:         ":" + cfg.WebPort,   // default: 8081
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        slog.Info("Web server starting", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("Web server error", "error", err)
            os.Exit(1)
        }
    }()

    <-quit
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("Web server shutdown error", "error", err)
    }
    slog.Info("Web server stopped")
}
```

### `web/embed.go`

```go
package web

import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed static
var staticFiles embed.FS

// StaticFS exposes the embedded static file system.
var StaticFS, _ = fs.Sub(staticFiles, "static")
```

### Makefile targets

```makefile
.PHONY: templ tailwind web-build web

# Run templ code generation
templ:
 @echo "==> Running templ generate..."
 templ generate ./...

# Build Tailwind CSS
tailwind:
 @echo "==> Building Tailwind CSS..."
 tailwindcss -i web/static/css/app.css -o web/static/css/app.min.css --minify

# Build web binary
web-build: templ tailwind
 @echo "==> Building web binary..."
 go build -o bin/moolah-web ./cmd/web/

# Run web server locally
web: web-build
 ./bin/moolah-web

# Update generate to include templ
generate: templ
 sqlc generate
```

### Config additions

Add `WEB_PORT` (default `8081`) to `internal/config/config.go`.  
The web server shares `DB_*` and `PASETO_*` env vars from the existing config for reading auth cookies.

---

## 5. Acceptance Criteria

- [x] `go mod vendor` runs cleanly with `a-h/templ` included.
- [x] `make templ` runs without errors (even with no `.templ` files yet).
- [x] `make tailwind` produces `web/static/css/app.min.css`.
- [x] `make web-build` produces `bin/moolah-web` binary.
- [x] `bin/moolah-web` starts and serves `GET /healthz` → `200 OK`.
- [x] `GET /static/js/htmx.min.js` and `GET /static/js/alpine.min.js` return `200 OK` (static embed works).
- [x] CI pipeline installs `templ` CLI and `tailwindcss` CLI and runs `make web-build`.
- [ ] `golangci-lint run ./cmd/web/...` passes.
- [ ] `docs/ROADMAP.md` row 4.2 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-15 | —      | Task created (new) |
