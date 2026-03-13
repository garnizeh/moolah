//go:build integration

// Package server_test contains integration-level smoke tests for the Moolah
// HTTP API. The TestSmoke_Phase1HappyPath test covers the full Phase 1 happy
// path: auth → tenant → accounts → categories → transactions → admin.
package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/handler"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/platform/idempotency"
	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/server"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/internal/testutil/containers"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/garnizeh/moolah/pkg/ulid"
)

// idempotencyHeader is the header name used by the idempotency middleware.
const idempotencyHeader = "Idempotency-Key"

// testPasetoHexKey is a fixed 32-byte (64 hex chars) symmetric key used only in tests.
const testPasetoHexKey = "707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f"

// TestSmoke_Phase1HappyPath exercises the full Phase 1 API journey end-to-end
// against real Postgres and Redis containers. It verifies that all Phase 1
// routes are correctly wired and return expected HTTP status codes for the
// happy path.
//
// Journey:
//
//	GET  /healthz
//	POST /v1/auth/otp/request      (admin user)
//	POST /v1/auth/otp/verify       → access token
//	GET  /v1/tenants/me
//	PATCH /v1/tenants/me
//	POST /v1/accounts              → accountID
//	GET  /v1/accounts
//	GET  /v1/accounts/{id}
//	PATCH /v1/accounts/{id}
//	POST /v1/categories            → categoryID
//	GET  /v1/categories
//	GET  /v1/categories/{id}
//	PATCH /v1/categories/{id}
//	POST /v1/transactions          → txID
//	GET  /v1/transactions
//	GET  /v1/transactions/{id}
//	PATCH /v1/transactions/{id}
//	DELETE /v1/transactions/{id}
//	DELETE /v1/categories/{id}
//	DELETE /v1/accounts/{id}
//	POST /v1/auth/otp/request      (sysadmin user)
//	POST /v1/auth/otp/verify       → sysadmin token
//	GET  /v1/admin/tenants
//	GET  /v1/admin/tenants/{id}
//	GET  /v1/admin/users
//	GET  /v1/admin/audit-logs
func TestSmoke_Phase1HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// ── 1. Infrastructure containers ────────────────────────────────────────
	pgDB := containers.NewPostgresDB(t)
	rdb := containers.NewRedisClient(t)

	// ── 2. Crypto & mailer ──────────────────────────────────────────────────
	pasetoKey, err := paseto.V4SymmetricKeyFromHex(testPasetoHexKey)
	require.NoError(t, err, "failed to parse test PASETO key")

	capMailer := mailer.NewCapturingMailer()

	// ── 3. Seed: tenant + admin user + sysadmin user ─────────────────────────
	tenantID := ulid.New()
	_, err = pgDB.Queries.CreateTenant(ctx, sqlc.CreateTenantParams{
		ID:   tenantID,
		Name: "Smoke Household",
		Plan: sqlc.TenantPlanFree,
	})
	require.NoError(t, err, "failed to seed tenant")

	adminEmail := fmt.Sprintf("admin-smoke-%s@example.com", tenantID)
	adminID := ulid.New()
	_, err = pgDB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       adminID,
		TenantID: tenantID,
		Email:    adminEmail,
		Name:     "Smoke Admin",
		Role:     sqlc.UserRoleAdmin,
	})
	require.NoError(t, err, "failed to seed admin user")

	sysadminEmail := fmt.Sprintf("sysadmin-smoke-%s@example.com", tenantID)
	sysadminID := ulid.New()
	_, err = pgDB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       sysadminID,
		TenantID: tenantID,
		Email:    sysadminEmail,
		Name:     "Smoke Sysadmin",
		Role:     sqlc.UserRoleSysadmin,
	})
	require.NoError(t, err, "failed to seed sysadmin user")

	// ── 4. Wire repos → services ────────────────────────────────────────────
	authRepo := repository.NewAuthRepository(pgDB.Queries)
	tenantRepo := repository.NewTenantRepository(pgDB.Queries)
	userRepo := repository.NewUserRepository(pgDB.Queries)
	accountRepo := repository.NewAccountRepository(pgDB.Queries)
	categoryRepo := repository.NewCategoryRepository(pgDB.Queries)
	transactionRepo := repository.NewTransactionRepository(pgDB.Queries)
	masterPurchaseRepo := repository.NewMasterPurchaseRepository(pgDB.Queries)
	auditRepo := repository.NewAuditRepository(pgDB.Queries)
	adminTenantRepo := repository.NewAdminTenantRepository(pgDB.Queries)
	adminUserRepo := repository.NewAdminUserRepository(pgDB.Queries)
	adminAuditRepo := repository.NewAdminAuditRepository(pgDB.Queries)

	idempotencyStore := idempotency.NewRedisStore(rdb)
	rateLimiterStore := middleware.NewRateLimiterStore()
	t.Cleanup(rateLimiterStore.Close)

	authSvc := service.NewAuthService(authRepo, userRepo, auditRepo, capMailer, pasetoKey)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, auditRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo, auditRepo)
	categorySvc := service.NewCategoryService(categoryRepo, auditRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo, auditRepo)
	masterPurchaseSvc := service.NewMasterPurchaseService(masterPurchaseRepo, accountRepo, categoryRepo)
	invoiceCloser := service.NewInvoiceCloser(masterPurchaseRepo, transactionRepo, auditRepo, accountRepo, masterPurchaseSvc, pgDB.Pool)
	adminSvc := service.NewAdminService(adminTenantRepo, adminUserRepo, adminAuditRepo, auditRepo)

	tokenParser := paseto.NewTokenParser(pasetoKey)

	// ── 5. Build server & test HTTP server ───────────────────────────────────
	srv := server.New(
		"0",
		authSvc,
		tenantSvc,
		accountSvc,
		categorySvc,
		transactionSvc,
		masterPurchaseSvc,
		invoiceCloser,
		adminSvc,
		idempotencyStore,
		rateLimiterStore,
		tokenParser,
	)

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	client := ts.Client()
	base := ts.URL

	// ── Helpers ──────────────────────────────────────────────────────────────

	// do sends an authenticated (or anonymous) HTTP request and returns the response.
	// The caller is responsible for closing the response body.
	// idempotencyKey is required for POST endpoints wrapped by the idempotency middleware;
	// pass an empty string for non-POST requests or routes without idempotency middleware.
	do := func(tb testing.TB, method, path string, body any, token, idempotencyKey string) *http.Response {
		tb.Helper()
		var buf bytes.Buffer
		if body != nil {
			b, encErr := json.Marshal(body)
			require.NoError(tb, encErr, "failed to marshal request body")
			buf = *bytes.NewBuffer(b)
		}
		req, reqErr := http.NewRequestWithContext(ctx, method, base+path, &buf)
		require.NoError(tb, reqErr, "failed to build request")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if idempotencyKey != "" {
			req.Header.Set(idempotencyHeader, idempotencyKey)
		}
		resp, doErr := client.Do(req)
		require.NoError(tb, doErr, "HTTP request failed")
		return resp
	}

	// decodeJSON decodes the response body into target.
	// The caller remains responsible for closing resp.Body.
	decodeJSON := func(tb testing.TB, resp *http.Response, target any) {
		tb.Helper()
		require.NoError(tb, json.NewDecoder(resp.Body).Decode(target), "failed to decode response body")
	}

	// ── Journey Variables ────────────────────────────────────────────────────
	var (
		accessToken   string
		refreshToken  string
		sysadminToken string
		accountID     string
		categoryID    string
		transactionID string
	)

	// idempotency keys — fixed per resource so we can replay them
	accountKey := ulid.New()
	categoryKey := ulid.New()
	transactionKey := ulid.New()
	tenantPatchKey := ulid.New()
	accountPatchKey := ulid.New()
	categoryPatchKey := ulid.New()
	transactionPatchKey := ulid.New()
	otpRequestKey := ulid.New()
	otpVerifyKey := ulid.New()
	sysOTPRequestKey := ulid.New()
	sysOTPVerifyKey := ulid.New()

	// ══════════════════════════════════════════════════════════════════════════
	// HEALTH CHECK
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("00_healthz", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/healthz", nil, "", "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// AUTH — OTP FLOW (admin user)
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("01_auth_otp_request", func(t *testing.T) {
		resp := do(t, http.MethodPost, "/v1/auth/otp/request",
			map[string]string{"email": adminEmail}, "", otpRequestKey)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("02_auth_otp_verify", func(t *testing.T) {
		code := capMailer.OTPFor(adminEmail)
		require.NotEmpty(t, code, "admin OTP should have been captured by mailer")

		var tokenResp handler.TokenResponse
		resp := do(t, http.MethodPost, "/v1/auth/otp/verify",
			map[string]string{"email": adminEmail, "code": code}, "", otpVerifyKey)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tokenResp)

		accessToken = tokenResp.AccessToken
		refreshToken = tokenResp.RefreshToken
		require.NotEmpty(t, accessToken, "access token must be non-empty after OTP verify")
		require.NotEmpty(t, refreshToken, "refresh token must be non-empty after OTP verify")
	})

	// ══════════════════════════════════════════════════════════════════════════
	// TENANT
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("03_get_tenant_me", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/tenants/me", nil, accessToken, "")
		defer resp.Body.Close()
		var tenant domain.Tenant
		decodeJSON(t, resp, &tenant)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, tenantID, tenant.ID)
	})

	t.Run("04_patch_tenant_me", func(t *testing.T) {
		newName := "Updated Smoke Household"
		resp := do(t, http.MethodPatch, "/v1/tenants/me",
			map[string]any{"name": &newName}, accessToken, tenantPatchKey)
		defer resp.Body.Close()
		var tenant domain.Tenant
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tenant)
		assert.Equal(t, newName, tenant.Name)
	})

	t.Run("04b_token_refresh", func(t *testing.T) {
		// POST /v1/auth/token/refresh does NOT use idempotency middleware;
		// pass the refresh token as the Bearer credential.
		resp := do(t, http.MethodPost, "/v1/auth/token/refresh", nil, refreshToken, "")
		defer resp.Body.Close()
		var tokenResp handler.TokenResponse
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tokenResp)
		assert.NotEmpty(t, tokenResp.AccessToken, "refreshed access token must be non-empty")
		assert.NotEmpty(t, tokenResp.RefreshToken, "refreshed refresh token must be non-empty")
	})

	// ══════════════════════════════════════════════════════════════════════════
	// ACCOUNTS
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("05_create_account", func(t *testing.T) {
		var acc domain.Account
		resp := do(t, http.MethodPost, "/v1/accounts", map[string]any{
			"name":          "Smoke Checking",
			"type":          "checking",
			"currency":      "USD",
			"initial_cents": 100000,
		}, accessToken, accountKey)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &acc)
		accountID = acc.ID
		require.NotEmpty(t, accountID, "account ID must be set after creation")
	})

	t.Run("05b_idempotency_replay_account", func(t *testing.T) {
		// Replay the same request with the same idempotency key — must return
		// the cached 201 response with the same account ID.
		var acc domain.Account
		resp := do(t, http.MethodPost, "/v1/accounts", map[string]any{
			"name":          "Smoke Checking",
			"type":          "checking",
			"currency":      "USD",
			"initial_cents": 100000,
		}, accessToken, accountKey)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &acc)
		assert.Equal(t, accountID, acc.ID, "idempotency replay must return the original account ID")
		assert.Equal(t, "HIT", resp.Header.Get("X-Cache"), "idempotency replay must set X-Cache: HIT")
	})

	t.Run("06_list_accounts", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/accounts", nil, accessToken, "")
		defer resp.Body.Close()
		var accs []domain.Account
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &accs)
		assert.NotEmpty(t, accs, "account list should contain at least one entry")
	})

	t.Run("07_get_account_by_id", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/accounts/"+accountID, nil, accessToken, "")
		defer resp.Body.Close()
		var acc domain.Account
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &acc)
		assert.Equal(t, accountID, acc.ID)
	})

	t.Run("08_update_account", func(t *testing.T) {
		newName := "Smoke Updated Checking"
		resp := do(t, http.MethodPatch, "/v1/accounts/"+accountID,
			map[string]any{"name": &newName}, accessToken, accountPatchKey)
		defer resp.Body.Close()
		var acc domain.Account
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &acc)
		assert.Equal(t, newName, acc.Name)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// CATEGORIES
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("09_create_category", func(t *testing.T) {
		var cat domain.Category
		resp := do(t, http.MethodPost, "/v1/categories", map[string]any{
			"name":  "Smoke Groceries",
			"type":  "expense",
			"icon":  "🛒",
			"color": "#FF5733",
		}, accessToken, categoryKey)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &cat)
		categoryID = cat.ID
		require.NotEmpty(t, categoryID, "category ID must be set after creation")
	})

	t.Run("10_list_categories", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/categories", nil, accessToken, "")
		defer resp.Body.Close()
		var cats []domain.Category
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &cats)
		assert.NotEmpty(t, cats, "category list should contain at least one entry")
	})

	t.Run("11_get_category_by_id", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/categories/"+categoryID, nil, accessToken, "")
		defer resp.Body.Close()
		var cat domain.Category
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &cat)
		assert.Equal(t, categoryID, cat.ID)
	})

	t.Run("12_update_category", func(t *testing.T) {
		newName := "Smoke Updated Groceries"
		resp := do(t, http.MethodPatch, "/v1/categories/"+categoryID,
			map[string]any{"name": &newName}, accessToken, categoryPatchKey)
		defer resp.Body.Close()
		var cat domain.Category
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &cat)
		assert.Equal(t, newName, cat.Name)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// TRANSACTIONS
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("13_create_transaction", func(t *testing.T) {
		var tx domain.Transaction
		resp := do(t, http.MethodPost, "/v1/transactions", map[string]any{
			"account_id":   accountID,
			"category_id":  categoryID,
			"description":  "Smoke grocery run",
			"type":         "expense",
			"amount_cents": 5000,
			"occurred_at":  time.Now().UTC().Format(time.RFC3339),
		}, accessToken, transactionKey)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &tx)
		transactionID = tx.ID
		require.NotEmpty(t, transactionID, "transaction ID must be set after creation")
	})

	t.Run("14_list_transactions", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/transactions", nil, accessToken, "")
		defer resp.Body.Close()
		var res struct {
			Data []domain.Transaction `json:"data"`
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &res)
		assert.NotEmpty(t, res.Data, "transaction list should contain at least one entry")
	})

	t.Run("15_get_transaction_by_id", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/transactions/"+transactionID, nil, accessToken, "")
		defer resp.Body.Close()
		var tx domain.Transaction
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tx)
		assert.Equal(t, transactionID, tx.ID)
	})

	t.Run("16_update_transaction", func(t *testing.T) {
		newDesc := "Smoke updated grocery run"
		resp := do(t, http.MethodPatch, "/v1/transactions/"+transactionID,
			map[string]any{"description": &newDesc}, accessToken, transactionPatchKey)
		defer resp.Body.Close()
		var tx domain.Transaction
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tx)
		assert.Equal(t, newDesc, tx.Description)
	})

	// ── Teardown of user-created resources ───────────────────────────────────

	t.Run("17_delete_transaction", func(t *testing.T) {
		resp := do(t, http.MethodDelete, "/v1/transactions/"+transactionID, nil, accessToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("18_delete_category", func(t *testing.T) {
		resp := do(t, http.MethodDelete, "/v1/categories/"+categoryID, nil, accessToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("19_delete_account", func(t *testing.T) {
		resp := do(t, http.MethodDelete, "/v1/accounts/"+accountID, nil, accessToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// ADMIN ROUTES (sysadmin token)
	// ══════════════════════════════════════════════════════════════════════════

	t.Run("20_admin_otp_request", func(t *testing.T) {
		resp := do(t, http.MethodPost, "/v1/auth/otp/request",
			map[string]string{"email": sysadminEmail}, "", sysOTPRequestKey)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("21_admin_otp_verify", func(t *testing.T) {
		code := capMailer.OTPFor(sysadminEmail)
		require.NotEmpty(t, code, "sysadmin OTP should have been captured by mailer")

		var tokenResp handler.TokenResponse
		resp := do(t, http.MethodPost, "/v1/auth/otp/verify",
			map[string]string{"email": sysadminEmail, "code": code}, "", sysOTPVerifyKey)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tokenResp)

		sysadminToken = tokenResp.AccessToken
		require.NotEmpty(t, sysadminToken, "sysadmin access token must be non-empty")
	})

	t.Run("22_admin_list_tenants", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/admin/tenants", nil, sysadminToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("23_admin_get_tenant", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/admin/tenants/"+tenantID, nil, sysadminToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("24_admin_list_users", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/admin/users", nil, sysadminToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("25_admin_list_audit_logs", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/admin/audit-logs", nil, sysadminToken, "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestSmoke_Phase2HappyPath exercises the full Phase 2 API journey end-to-end.
// It covers Credit Card accounts, Master Purchases (installments), and Invoice Closing.
//
// Journey:
//
//	00: Health Check
//	01-02: Auth (OTP)
//	03: Create Credit Card Account
//	04: Create Category
//	05: Create Master Purchase (3 installments) -> capture mpID, assert projection
//	05b: Idempotency Replay (Master Purchase)
//	06: List Master Purchases
//	07: Get Master Purchase by ID -> verify projected_schedule
//	08: Get Master Purchase Projection (separate endpoint)
//	09: Update Master Purchase
//	10: Close Invoice -> processed_count=1
//	11: List Transactions -> assert materialized installment exists
//	12: List Audit Logs -> assert SYSTEM actor entry
func TestSmoke_Phase2HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// ── 1. Infrastructure containers ────────────────────────────────────────
	pgDB := containers.NewPostgresDB(t)
	rdb := containers.NewRedisClient(t)

	// ── 2. Crypto & mailer ──────────────────────────────────────────────────
	pasetoKey, err := paseto.V4SymmetricKeyFromHex(testPasetoHexKey)
	require.NoError(t, err, "failed to parse test PASETO key")

	capMailer := mailer.NewCapturingMailer()

	// ── 3. Seed: tenant + admin user ────────────────────────────────────────
	tenantID := ulid.New()
	_, err = pgDB.Queries.CreateTenant(ctx, sqlc.CreateTenantParams{
		ID:   tenantID,
		Name: "Phase 2 Smoke Household",
		Plan: sqlc.TenantPlanFree,
	})
	require.NoError(t, err, "failed to seed tenant")

	adminEmail := fmt.Sprintf("admin-smoke-p2-%s@example.com", tenantID)
	adminID := ulid.New()
	_, err = pgDB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:       adminID,
		TenantID: tenantID,
		Email:    adminEmail,
		Name:     "P2 Smoke Admin",
		Role:     sqlc.UserRoleAdmin,
	})
	require.NoError(t, err, "failed to seed admin user")

	// ── 4. Wire repos → services ────────────────────────────────────────────
	authRepo := repository.NewAuthRepository(pgDB.Queries)
	tenantRepo := repository.NewTenantRepository(pgDB.Queries)
	userRepo := repository.NewUserRepository(pgDB.Queries)
	accountRepo := repository.NewAccountRepository(pgDB.Queries)
	categoryRepo := repository.NewCategoryRepository(pgDB.Queries)
	transactionRepo := repository.NewTransactionRepository(pgDB.Queries)
	masterPurchaseRepo := repository.NewMasterPurchaseRepository(pgDB.Queries)
	auditRepo := repository.NewAuditRepository(pgDB.Queries)
	adminTenantRepo := repository.NewAdminTenantRepository(pgDB.Queries)
	adminUserRepo := repository.NewAdminUserRepository(pgDB.Queries)
	adminAuditRepo := repository.NewAdminAuditRepository(pgDB.Queries)

	idempotencyStore := idempotency.NewRedisStore(rdb)
	rateLimiterStore := middleware.NewRateLimiterStore()
	t.Cleanup(rateLimiterStore.Close)

	authSvc := service.NewAuthService(authRepo, userRepo, auditRepo, capMailer, pasetoKey)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, auditRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo, auditRepo)
	categorySvc := service.NewCategoryService(categoryRepo, auditRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo, auditRepo)
	masterPurchaseSvc := service.NewMasterPurchaseService(masterPurchaseRepo, accountRepo, categoryRepo)
	invoiceCloser := service.NewInvoiceCloser(masterPurchaseRepo, transactionRepo, auditRepo, accountRepo, masterPurchaseSvc, pgDB.Pool)
	adminSvc := service.NewAdminService(adminTenantRepo, adminUserRepo, adminAuditRepo, auditRepo)

	tokenParser := paseto.NewTokenParser(pasetoKey)

	// ── 5. Build server & test HTTP server ───────────────────────────────────
	srv := server.New(
		"0",
		authSvc,
		tenantSvc,
		accountSvc,
		categorySvc,
		transactionSvc,
		masterPurchaseSvc,
		invoiceCloser,
		adminSvc,
		idempotencyStore,
		rateLimiterStore,
		tokenParser,
	)

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	client := ts.Client()
	base := ts.URL

	// ── Helpers ──────────────────────────────────────────────────────────────
	do := func(tb testing.TB, method, path string, body any, token, idempotencyKey string) *http.Response {
		tb.Helper()
		var buf bytes.Buffer
		if body != nil {
			b, encErr := json.Marshal(body)
			require.NoError(tb, encErr, "failed to marshal request body")
			buf = *bytes.NewBuffer(b)
		}
		req, reqErr := http.NewRequestWithContext(ctx, method, base+path, &buf)
		require.NoError(tb, reqErr, "failed to build request")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if idempotencyKey != "" {
			req.Header.Set(idempotencyHeader, idempotencyKey)
		}
		resp, doErr := client.Do(req)
		require.NoError(tb, doErr, "HTTP request failed")
		return resp
	}

	decodeJSON := func(tb testing.TB, resp *http.Response, target any) {
		tb.Helper()
		require.NoError(tb, json.NewDecoder(resp.Body).Decode(target), "failed to decode response body")
	}

	// ── Journey Variables ────────────────────────────────────────────────────
	var (
		accessToken      string
		accountID        string
		categoryID       string
		masterPurchaseID string
	)

	otpRequestKey := ulid.New()
	otpVerifyKey := ulid.New()
	mpKey := ulid.New()

	// ══════════════════════════════════════════════════════════════════════════
	// 00. HEALTH CHECK
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("00_healthz", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/healthz", nil, "", "")
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 01-02. AUTH
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("01_auth_otp_request", func(t *testing.T) {
		resp := do(t, http.MethodPost, "/v1/auth/otp/request", map[string]string{"email": adminEmail}, "", otpRequestKey)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("02_auth_otp_verify", func(t *testing.T) {
		code := capMailer.OTPFor(adminEmail)
		require.NotEmpty(t, code)
		var tokenResp handler.TokenResponse
		resp := do(t, http.MethodPost, "/v1/auth/otp/verify", map[string]string{"email": adminEmail, "code": code}, "", otpVerifyKey)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &tokenResp)
		accessToken = tokenResp.AccessToken
		require.NotEmpty(t, accessToken)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 03. CREATE CREDIT CARD ACCOUNT
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("03_create_cc_account", func(t *testing.T) {
		var acc domain.Account
		resp := do(t, http.MethodPost, "/v1/accounts", map[string]any{
			"name":          "Smoke Visa",
			"type":          "credit_card",
			"currency":      "BRL",
			"initial_cents": int64(1), // must be > 0 to pass validator:'required'
		}, accessToken, ulid.New())
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &acc)
		accountID = acc.ID
		require.NotEmpty(t, accountID)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 04. CREATE CATEGORY
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("04_create_category", func(t *testing.T) {
		var cat domain.Category
		resp := do(t, http.MethodPost, "/v1/categories", map[string]any{
			"name": "Electronics",
			"type": "expense",
		}, accessToken, ulid.New())
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &cat)
		categoryID = cat.ID
		require.NotEmpty(t, categoryID)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 05. CREATE MASTER PURCHASE
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("05_create_master_purchase", func(t *testing.T) {
		var mp domain.MasterPurchase
		resp := do(t, http.MethodPost, "/v1/master-purchases", map[string]any{
			"account_id":             accountID,
			"category_id":            categoryID,
			"description":            "New iPhone",
			"total_amount_cents":     int64(120000), // 1200.00
			"installment_count":      int32(3),
			"first_installment_date": time.Now().UTC().Format(time.RFC3339),
			"closing_day":            int32(28),
		}, accessToken, mpKey)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		decodeJSON(t, resp, &mp)

		masterPurchaseID = mp.ID
		require.NotEmpty(t, masterPurchaseID)
		assert.Equal(t, int64(120000), mp.TotalAmountCents)
	})

	t.Run("05b_idempotency_replay_mp", func(t *testing.T) {
		resp := do(t, http.MethodPost, "/v1/master-purchases", map[string]any{
			"account_id":             accountID,
			"category_id":            categoryID,
			"description":            "New iPhone",
			"total_amount_cents":     int64(120000),
			"installment_count":      int32(3),
			"first_installment_date": time.Now().UTC().Format(time.RFC3339),
			"closing_day":            int32(28),
		}, accessToken, mpKey)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "HIT", resp.Header.Get("X-Cache"))
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 06-09. MASTER PURCHASE QUERIES & UPDATES
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("06_list_master_purchases", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/master-purchases", nil, accessToken, "")
		defer resp.Body.Close()
		var list []domain.MasterPurchase
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &list)
		assert.NotEmpty(t, list)
	})

	t.Run("07_get_master_purchase_by_id", func(t *testing.T) {
		var mp domain.MasterPurchase
		resp := do(t, http.MethodGet, "/v1/master-purchases/"+masterPurchaseID, nil, accessToken, "")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &mp)
		assert.Equal(t, masterPurchaseID, mp.ID)
	})

	t.Run("08_get_master_purchase_projection", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/master-purchases/"+masterPurchaseID+"/projection", nil, accessToken, "")
		defer resp.Body.Close()
		var schedule []domain.ProjectedInstallment
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &schedule)
		assert.Len(t, schedule, 3)
	})

	t.Run("09_update_master_purchase", func(t *testing.T) {
		newDesc := "iPhone 16 Pro"
		resp := do(t, http.MethodPatch, "/v1/master-purchases/"+masterPurchaseID, map[string]any{
			"description": newDesc,
		}, accessToken, ulid.New())
		defer resp.Body.Close()

		var mp domain.MasterPurchase
		require.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &mp)
		assert.Equal(t, newDesc, mp.Description)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 10. CLOSE INVOICE
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("10_close_invoice", func(t *testing.T) {
		var closeResp struct {
			Errors         []string `json:"errors"`
			ProcessedCount int      `json:"processed_count"`
		}
		resp := do(t, http.MethodPost, "/v1/accounts/"+accountID+"/close-invoice", nil, accessToken, ulid.New())
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &closeResp)
		assert.Equal(t, 1, closeResp.ProcessedCount)
		assert.Empty(t, closeResp.Errors)
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 11. LIST TRANSACTIONS (assert materialized)
	// ══════════════════════════════════════════════════════════════════════════
	t.Run("11_assert_materialized_tx", func(t *testing.T) {
		resp := do(t, http.MethodGet, "/v1/transactions", nil, accessToken, "")
		defer resp.Body.Close()
		var res struct {
			Data []domain.Transaction `json:"data"`
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &res)

		var found bool
		for _, tx := range res.Data {
			if tx.MasterPurchaseID == masterPurchaseID {
				found = true
				assert.Equal(t, int64(40000), tx.AmountCents) // 120000 / 3
				assert.Equal(t, domain.TransactionTypeExpense, tx.Type)
			}
		}
		assert.True(t, found, "materialized installment transaction not found in list")
	})

	// ══════════════════════════════════════════════════════════════════════════
	// 12. AUDIT LOG (assert SYSTEM actor)
	// ══════════════════════════════════════════════════════════════════════════
	// Note: We need a sysadmin token to view audit logs via the API.
	t.Run("12_assert_system_audit_log", func(t *testing.T) {
		// Create sysadmin user
		sysEmail := "sys@moolah.com"
		_, err = pgDB.Queries.CreateUser(ctx, sqlc.CreateUserParams{
			ID:       ulid.New(),
			TenantID: tenantID,
			Email:    sysEmail,
			Name:     "System Admin",
			Role:     sqlc.UserRoleSysadmin,
		})
		require.NoError(t, err)

		// Get sys token
		reqResp := do(t, http.MethodPost, "/v1/auth/otp/request", map[string]string{"email": sysEmail}, "", ulid.New())
		require.NoError(t, reqResp.Body.Close())
		code := capMailer.OTPFor(sysEmail)
		var tResp handler.TokenResponse
		vResp := do(t, http.MethodPost, "/v1/auth/otp/verify", map[string]string{"email": sysEmail, "code": code}, "", ulid.New())
		decodeJSON(t, vResp, &tResp)
		vResp.Body.Close()

		// List audit logs
		resp := do(t, http.MethodGet, "/v1/admin/audit-logs", nil, tResp.AccessToken, "")
		defer resp.Body.Close()
		var res struct {
			Data []domain.AuditLog `json:"data"`
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		decodeJSON(t, resp, &res)

		var foundSystem bool
		for _, alog := range res.Data {
			if alog.ActorID == domain.ActorSystem {
				foundSystem = true
				assert.Equal(t, domain.AuditActionCreate, alog.Action)
				assert.Equal(t, "transaction", alog.EntityType)
			}
		}
		assert.True(t, foundSystem, "audit log item with SYSTEM actor not found")
	})
}
