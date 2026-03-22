# Technical Architecture & Engineering Standards: Moolah

## 1. Core Engineering Philosophy
* **Standard Library First:** External dependencies are a liability. The `stdlib` (`net/http`, `context`, `encoding/json`) will be the default choice for all core application plumbing. Third-party packages are only introduced when they provide significant, undeniable leverage (e.g., `sqlc` for type-safe SQL, OTel for tracing).
* **Decoupling of API and UI:** The backend must remain a pure data and presentation server. The REST API will serve strictly structured data, while a separate presentation layer (or dedicated handlers) will serve the HTML fragments for the UI. This ensures the core financial engine can eventually power mobile apps or third-party integrations without UI entanglement.

## 2. The Backend Stack (Go)
* **Language:** Go (Golang) 1.22+.
* **Routing:** **[Suggestion]** Utilize the enhanced Go 1.22 `net/http` ServeMux, which natively supports method and wildcard routing (e.g., `GET /api/v1/expenses/{id}`), eliminating the need for external routers like Chi or Gorilla.
* **Logging:** `log/slog` for structured, leveled JSON logging.
* **Identifiers:** `ULID` (Universally Unique Lexicographically Sortable Identifier) for all database primary keys and correlation IDs. This provides the uniqueness of a UUID while allowing database indexes to sort sequentially by time, reducing page fragmentation in Postgres.
* **Observability:** Native integration with OpenTelemetry (OTel). All incoming requests will initiate a trace span, which will be propagated through the `context.Context` to the database and cache layers. This prepares the system for full metric, trace, and log export to platforms like Datadog, Jaeger, or Grafana.



## 3. Data Persistence & Caching
* **Primary Database:** PostgreSQL. Chosen for its robust transactional integrity, essential for financial ledgers.
* **Database Access:** `sqlc`. We will write pure SQL queries, and `sqlc` will generate idiomatic, type-safe Go code. No ORMs (like GORM) will be used, ensuring maximum performance and query transparency.
* **Caching Layer:** Redis. Used for caching frequently accessed, slow-changing data (e.g., categorized expense lists, user sessions) and rate-limiting.
* **Schema Migrations:** **[Suggestion]** Use `pressly/goose`. It integrates perfectly with the standard `database/sql` package and allows migrations to be written in pure `.sql` files or Go code.

## 4. The Frontend Stack (Hypermedia-Driven)
* **DOM Manipulation & AJAX:** HTMX. The UI will be driven by hypermedia exchanges. The Go server will render and return HTML fragments over the wire, drastically reducing frontend state management complexity.
* **Styling:** Tailwind CSS. Utility-first styling for rapid, consistent UI development directly within Go HTML templates.
* **Client-Side Interactivity:** Alpine.js. Used strictly for ephemeral UI state that doesn't need a round-trip to the server (e.g., toggling modals, dropdowns, or mobile menus).

## 5. Architectural Patterns & Flow
### 5.1 Request ID & Error Handling
Every incoming HTTP request will be assigned a ULID upon hitting the first middleware. This Request ID will be:
1. Injected into the request `context.Context`.
2. Attached to all `slog` entries associated with that request.
3. Appended to the OpenTelemetry trace.
4. Returned in the JSON payload or HTML response of any HTTP 4xx/5xx error, allowing the user to report the exact ID for debugging.

### 5.2 Middleware Chain
Middleware will be strictly layered:
* **Global Middleware:** Applied to all routes. Includes Panic Recovery, OTel Tracing Initialization, Request ID Generation, and generic Request Logging.
* **Resource/Endpoint Middleware:** Applied to specific routes or groups. Includes Authentication (JWT/Session validation), Authorization (RBAC), Rate Limiting (via Redis), and specific Context value injection.

### 5.3 Background Processing
* **Task Queue:** **[Suggestion]** Use `hibiken/asynq`. Since Redis is already in the stack, `asynq` provides a robust, production-ready background worker queue. This is essential for offloading heavy tasks like the Monte Carlo retirement simulations, parsing large banking CSVs, or triggering the automated variance/pending checks without blocking the HTTP response.

### 5.4 Testing Strategy
* **Integration Testing:** **[Suggestion]** Utilize `testcontainers-go`. This allows the test suite to programmatically spin up ephemeral, isolated Docker containers for PostgreSQL and Redis during `go test` runs, ensuring the `sqlc` queries and caching logic are tested against real infrastructure, not mocks.