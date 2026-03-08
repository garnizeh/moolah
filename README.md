# Moolah — Household Finance & Investment SaaS

[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/garnizeh/moolah/ci.yml?branch=main&label=build&logo=github)](https://github.com/garnizeh/moolah/actions)
[![Coverage](https://codecov.io/gh/garnizeh/moolah/branch/main/graph/badge.svg)](https://codecov.io/gh/garnizeh/moolah)
[![Go Report Card](https://goreportcard.com/badge/github.com/garnizeh/moolah)](https://goreportcard.com/report/github.com/garnizeh/moolah)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Moolah is a multi-tenant personal finance application built with **Go**, focusing on simplicity, scalability, and financial integrity. It helps households manage their accounts payable, cash flow, credit card installments, and investment portfolios.

## 🚀 Technical Stack

- **Backend:** Go 1.26.1 (Standard library `net/http` for routing).
- **Database:** PostgreSQL (Querying via `sqlc`).
- **Identity:** ULID for all primary keys.
- **Security:** PASETO v4 tokens and OTP-only authentication.
- **Infrastructure:** Docker & Docker Compose.
- **Linter:** `golangci-lint`.

## 🏗️ Project Structure

Following a Pragmatic DDD approach:

- `cmd/api/`: Application entry point and DI.
- `internal/domain/`: Business logic, entities, and repository interfaces.
- `internal/platform/`: Infrastructure implementations (DB, Mailer, Middleware).
- `internal/service/`: Domain orchestration and business rules.
- `pkg/`: Generic utilities.
- `docs/`: Architecture Decision Records (ADRs) and Roadmap.

## 🛠️ Development

### Prerequisites

- Go 1.26.1+
- Docker & Docker Compose
- `golangci-lint`
- `sqlc`

### Getting Started

1. **Clone the repository:**

   ```bash
   git clone https://github.com/garnizeh/moolah.git
   cd moolah
   ```

2. **Spin up infrastructure:**

   ```bash
   docker-compose up -d
   ```

3. **Run the application:**

   ```bash
   make run
   ```

### Makefile Commands

- `make build`: Build the API binary.
- `make test`: Run unit tests.
- `make lint`: Run the linter.
- `make generate`: Generate code from SQL queries using `sqlc`.

## 🧪 API Testing & Collections

Moolah uses **[Bruno](https://usebruno.com/)** for API exploration and testing. You can find the collection in `docs/bruno/moolah`.

To use:

1. Download and install Bruno.
2. "Open Collection" and select the `docs/bruno/moolah` folder.
3. Configure your environment variables (e.g., `base_url`).

## 📈 Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for the detailed project roadmap and current status.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
