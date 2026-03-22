# Moolah 💸

**Moolah** is a high-performance personal finance management system designed to replace manual spreadsheets with a robust, API-driven architecture. It focuses on the evolution from retrospective logging to proactive wealth engineering.

## 🚀 Vision: Simple to Sophisticated
Transition from fragmented spreadsheets to a centralized system that models multi-entity family finances, long-term obligations, and automated investment rebalancing.

### Core Principles
- **Always Cents**: Monetary values are stored as 64-bit integers in their smallest unit (cents) to maintain absolute precision.
- **Extensible Schema**: Powered by PostgreSQL 17 `JSONB` to handle evolving financial metadata without constant schema migrations.
- **Clean Architecture**: A Go-based backend providing a pure REST API and HTMX-driven hypermedia fragments.

## 🛠️ Tech Stack
- **Backend**: Go 1.22+ (Standard Library, `slog`, `sqlc`, `goose`)
- **Database**: PostgreSQL 17 (Relational + JSONB)
- **Cache & Jobs**: Redis 7 + `asynq`
- **Frontend**: HTMX, Tailwind CSS, Alpine.js
- **Observability**: OpenTelemetry (OTel) + ULID Request Tracing

## 📖 Documentation
Detailed project documentation can be found in the `docs/` directory:

- **[Product RFC (003)](docs/design/003-moolah-product.md)**: The core product vision and architectural blueprint.
- **[Initial Design (001)](docs/design/001-moolah-initial-design-doc.md)**: Deep dive into the legacy spreadsheet logic and transition goals.
- **[Architecture (002)](docs/design/002-moolah-initial-architecture.md)**: Technical engineering standards.
- **[MVP Roadmap](docs/tasks/roadmap.md)**: High-level 6-phase implementation plan.
- **[Milestone Tasks](docs/tasks/milestone-01-mvp/)**: Granular task definitions following the project's task template.

## 🚦 Getting Started (Prerequisites)
To contribute or run Moolah locally, you will need:
- [Go 1.22+](https://golang.org/dl/)
- [Docker & Docker Compose](https://www.docker.com/products/docker-desktop)
- [sqlc](https://sqlc.dev/) (for code generation)

```bash
# Clone the repository
git clone https://github.com/garnizeh/moolah.git
cd moolah

# Start local infrastructure
docker-compose up -d

# Run migrations (once implemented)
# goose -dir migrations postgres "user=... dbname=moolah" up

# Run the API
# go run cmd/api/main.go
```

## 📜 License
Distributable under the [MIT License](LICENSE).
