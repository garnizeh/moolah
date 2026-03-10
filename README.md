# Moolah — Household Finance & Investment SaaS

[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/garnizeh/moolah/ci.yml?branch=main&label=build&logo=github)](https://github.com/garnizeh/moolah/actions)
[![codecov](https://codecov.io/gh/garnizeh/moolah/graph/badge.svg?v=3)](https://codecov.io/gh/garnizeh/moolah)
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

2. **Configure environment variables:**

   Copy the sample environment file and set the required variables:

   ```bash
   cp .env.example .env
   ```

   **Required variables for first-run bootstrap:**
   - `SYSADMIN_EMAIL`: The email address of the first system administrator. This is required to break the OTP bootstrap cycle.

3. **Spin up infrastructure:**

   ```bash
   docker-compose up -d
   ```

4. **Run the application:**

   ```bash
   make run
   ```

### Makefile Commands

- `make build`: Build the API binary.
- `make test`: Run unit tests.
- `make lint`: Run the linter.
- `make generate`: Generate code from SQL queries using `sqlc`.
- `make swagger`: Generate Swagger documentation.

## 🧪 Testing & API Exploration

### Bruno Collection

We use [Bruno](https://www.usebruno.com/) for API exploration and manual testing. The collection is located in `docs/bruno/`.

To use the collection:

1. Install Bruno.
2. Open Bruno and select "Open Collection".
3. Navigate to `docs/bruno/` and open the folder.
4. Select the `moolah` environment (usually `local`).
5. Start with the **Auth > Request OTP** request.

### Swagger UI

API documentation is also available via Swagger UI when the server is running:

- **URL:** [http://localhost:8080/swagger/](http://localhost:8080/swagger/)

## 📈 Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for the detailed project roadmap and current status.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
