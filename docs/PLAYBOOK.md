# Moolah API Playbook

This playbook documents how to use the Moolah API (v1) with the currently implemented functionality. It contains quick start steps, common use-cases, example curl requests, and operational notes (bootstrap, env, headers, error handling).

**Audience:** developers, QA, DevOps who want to exercise the API or automate integration tests.

---

## 1. Quick prerequisites

- Run local infrastructure: `docker-compose up -d` (Postgres, Redis, MailHog).
- Copy environment file and set the required variables (see `.env`):

  - `DATABASE_URL` — postgres connection
  - `REDIS_ADDR` — redis address
  - `PASETO_SECRET_KEY` — 32-byte hex key
  - `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `EMAIL_FROM`
  - `SYSADMIN_EMAIL` — required for first-run bootstrap (the email used to create the initial `sysadmin` user)

- Start the API: `make run` (or `go run ./cmd/api`). On first startup the app will run migrations and automatically bootstrap a `sysadmin` user with the `SYSADMIN_EMAIL` value.

---

## 2. Authentication (OTP-only)

All users authenticate with an email-only OTP flow.

1. Request OTP

- Endpoint: `POST /v1/auth/otp/request`
- Body: `{ "email": "user@example.com" }`
- Response: 200 on success. The OTP is delivered by email (MailHog in local dev).
- Rate-limited by middleware (token-bucket). Expect 429 if abused.

Example curl:

```bash
curl -X POST "http://localhost:8080/v1/auth/otp/request" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@moolah.local"}'
```

1. Verify OTP (get token)

- Endpoint: `POST /v1/auth/otp/verify`
- Body: `{ "email": "user@example.com", "code": "123456" }`
- Response: 200 with a JSON containing the PASETO token.

Example curl:

```bash
curl -X POST "http://localhost:8080/v1/auth/otp/verify" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@moolah.local","code":"123456"}'
```

After verify you receive a bearer token. Add it to requests:

- Header: `Authorization: Bearer <token>`

---

## 3. Sysadmin bootstrap

- On first run, the application creates a `System` tenant and a `sysadmin` user using `SYSADMIN_EMAIL` env var. This user can request OTP and obtain a sysadmin token to call admin endpoints.
- If you change `SYSADMIN_EMAIL` after bootstrap, you will need to manually create or change the sysadmin in DB.

---

## 4. Common workflows and endpoints

Notes: every tenant-scoped query requires a `tenant_id` context; handlers extract tenant info from the token. Monetary fields use cents (int64). Soft deletes are used (queries filter `deleted_at IS NULL`).

A. Create a tenant (sysadmin)

- Endpoint: `POST /v1/admin/tenants` (admin-only)
- Auth: `sysadmin` token required
- Body: `{ "name": "Household Name" }`

B. Invite user to tenant (tenant admin)

- Endpoint: `POST /v1/tenants/{tenant_id}/users` or service `InviteUser` (check handlers)
- Role: `member` or `admin` (must be enforced)
- Body: `{ "email": "member@example.com", "name": "Member" }`

C. Create Account

- Endpoint: `POST /v1/accounts`
- Auth: Bearer
- Headers: `Idempotency-Key` for POSTs that must be idempotent
- Body example:

  ```json
  {
    "name": "Savings Account",
    "type": "CHECKING",
    "initial_balance": 100000
  }
  ```

- Note: `initial_balance` is in cents.

D. Create Category

- Endpoint: `POST /v1/categories`
- Body example:

  ```json
  { "name": "Groceries", "type": "EXPENSE" }
  ```

E. Create Transaction

- Endpoint: `POST /v1/transactions`
- Body example:

  ```json
  {
    "account_id": "01F...",
    "amount_cents": 2500,
    "type": "DEBIT",
    "category_id": "01F...",
    "description": "Grocery at Market"
  }
  ```

F. Listing endpoints

- `GET /v1/categories` — list categories for tenant
- `GET /v1/accounts` — list accounts for tenant
- `GET /v1/transactions` — list transactions (supports tenant-scoped filters)

G. Admin-only endpoints

- `GET /v1/admin/tenants` — list tenants
- `GET /v1/admin/users` — list all users across tenants
- `DELETE /v1/admin/users/{id}` — hard-delete a user
- `GET /v1/admin/audit-logs` — global audit logs

---

## 5. Example end-to-end sequence (local dev)

1. Ensure `.env` has `SYSADMIN_EMAIL=admin@moolah.local` and run `docker-compose up -d`.
2. Start the API: `make run`. Confirm logs show "sysadmin bootstrapped" on first run.
3. Request OTP for the sysadmin (MailHog will capture email):

```bash
curl -X POST http://localhost:8080/v1/auth/otp/request \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@moolah.local"}'
```

1. Retrieve the code from MailHog (<http://localhost:8025>) and verify:

```bash
curl -X POST http://localhost:8080/v1/auth/otp/verify \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@moolah.local","code":"<code-from-mailhog>"}'
```

1. Use returned token to call admin endpoints. Create a tenant or perform admin actions.
2. Create a tenant user (tenant admin flow) and then create accounts/categories/transactions as that user.

---

## 6. Headers and patterns

- `Authorization: Bearer <token>` — required for protected endpoints.
- `Idempotency-Key: <uuid>` — recommended on `POST` endpoints that must be safe to retry.
- Content-type: `application/json` for JSON bodies.

---

## 7. Operational notes & constraints

- Multi-tenancy: All tenant-scoped queries include `tenant_id` filters. Never query tenant-scoped data without tenant context.
- Monetary values: use integer cents (`int64`) to avoid floating point rounding.
- Soft deletes: records are soft-deleted using `deleted_at`; list endpoints filter these out.
- OTP rate-limiting: enforced by middleware (throttle frequent requests).
- Audit: actions are logged to audit logs (admin can query global audit logs).
- ULIDs: All primary keys are ULIDs (26-char strings). Use `pkg/ulid` helper when generating client-side fixtures.

---

## 8. Helpful files

- Bruno collection (requests): `docs/bruno/` — open the folder in Bruno to explore requests.
- Roadmap & task details: `docs/ROADMAP.md` and `docs/tasks/`.
- Bootstrap implementation: `internal/platform/bootstrap/sysadmin.go`.

---

If you want, I can:

- Add concrete curl examples for every endpoint (with full request/response examples).
- Export the curl requests into the Bruno collection or a Postman collection.
- Translate this playbook to Portuguese.

Tell me which of the above you prefer next.

## 9. Full curl examples

The following are ready-to-run curl examples (requests + sample responses) for the currently implemented endpoints. Replace placeholder values (`<...>`) before running.

Note: all examples use `http://localhost:8080` and assume the API is running locally.

9.1 Auth — Request OTP

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/auth/otp/request" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@moolah.local"}'
```

Sample response: 200 OK (empty body)

9.2 Auth — Verify OTP (exchange for token)

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/auth/otp/verify" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@moolah.local","code":"123456"}'
```

Sample response (200):

```json
{
  "token": "v4.local.<pase to token>"
}
```

9.3 Admin — List Tenants

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/admin/tenants" \
  -H "Authorization: Bearer <SYSADMIN_TOKEN>"
```

Sample response (200):

```json
[
  { "id": "01F...", "name": "System", "plan": "free" }
]
```

9.4 Admin — Get Tenant by ID

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/admin/tenants/<TENANT_ID>" \
  -H "Authorization: Bearer <SYSADMIN_TOKEN>"
```

Sample response (200):

```json
{ "id": "01F...", "name": "Household Name", "plan": "free" }
```

9.5 Admin — Create Tenant

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/admin/tenants" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <SYSADMIN_TOKEN>" \
  -d '{"name":"New Household"}'
```

Sample response (201):

```json
{ "id": "01F...", "name": "New Household", "plan": "free" }
```

9.6 Admin — List Users (cross-tenant)

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/admin/users" \
  -H "Authorization: Bearer <SYSADMIN_TOKEN>"
```

Sample response (200):

```json
[
  { "id": "01G...", "email": "admin@moolah.local", "role": "sysadmin", "tenant_id": "01F..." }
]
```

9.7 Admin — List Audit Logs

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/admin/audit-logs" \
  -H "Authorization: Bearer <SYSADMIN_TOKEN>"
```

Sample response (200):

```json
[
  { "id": "01A...", "action": "create", "actor_id": "01G...", "entity": "tenant", "entity_id": "01F...", "metadata": {} }
]
```

9.8 Invite User to Tenant (Tenant admin)

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/tenants/<TENANT_ID>/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TENANT_ADMIN_TOKEN>" \
  -d '{"email":"member@example.com","name":"Member","role":"member"}'
```

Sample response (201):

```json
{ "id": "01H...", "email": "member@example.com", "name": "Member", "role": "member", "tenant_id": "<TENANT_ID>" }
```

9.9 Create Account

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/accounts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"name":"Savings","type":"CHECKING","initial_balance":100000}'
```

Sample response (201):

```json
{ "id": "01C...", "name":"Savings","type":"CHECKING","balance_cents":100000 }
```

9.10 List Accounts

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/accounts" \
  -H "Authorization: Bearer <TOKEN>"
```

Sample response (200):

```json
[
  { "id":"01C...","name":"Savings","type":"CHECKING","balance_cents":100000 }
]
```

9.11 Create Category

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/categories" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"name":"Groceries","type":"EXPENSE"}'
```

Sample response (201):

```json
{ "id":"01K...","name":"Groceries","type":"EXPENSE" }
```

9.12 List Categories

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/categories" \
  -H "Authorization: Bearer <TOKEN>"
```

Sample response (200):

```json
[
  { "id":"01K...","name":"Groceries","type":"EXPENSE" }
]
```

9.13 Create Transaction

Request:

```bash
curl -s -X POST "http://localhost:8080/v1/transactions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"account_id":"<ACCOUNT_ID>","amount_cents":2500,"type":"DEBIT","category_id":"<CATEGORY_ID>","description":"Grocery"}'
```

Sample response (201):

```json
{ "id":"01T...","account_id":"<ACCOUNT_ID>","amount_cents":2500,"type":"DEBIT","category_id":"<CATEGORY_ID>","description":"Grocery" }
```

9.14 List Transactions

Request:

```bash
curl -s -X GET "http://localhost:8080/v1/transactions" \
  -H "Authorization: Bearer <TOKEN>"
```

Sample response (200):

```json
[
  { "id":"01T...","account_id":"<ACCOUNT_ID>","amount_cents":2500,"type":"DEBIT","description":"Grocery" }
]
```

9.15 Admin — Force Delete User

Request:

```bash
curl -s -X DELETE "http://localhost:8080/v1/admin/users/<USER_ID>" \
  -H "Authorization: Bearer <SYSADMIN_TOKEN>"
```

Sample response: 204 No Content

---

If you'd like, I can also:

- Export these curl examples into the Bruno collection in `docs/bruno/` as executable requests.
- Generate a small script (`scripts/e2e.sh`) to run the happy-path sequence against a local environment.

Which follow-up would you like?
