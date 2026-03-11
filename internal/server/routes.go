package server

import (
	"fmt"
	"net/http"

	_ "github.com/garnizeh/moolah/api"
	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// 1. Auth & Public Routes
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.Handle("GET /swagger/", httpSwagger.Handler())

	// Inject middleware helpers
	requireAuth := middleware.RequireAuth(s.tokenParser)
	rateLimit := s.rateLimiterStore.OTPRateLimiter()
	idempotency := middleware.Idempotency(s.idempotencyStore)

	// Auth routes (Task 1.5.4)
	mux.Handle("POST /v1/auth/otp/request", rateLimit(idempotency(http.HandlerFunc(s.authHandler.RequestOTP))))
	mux.Handle("POST /v1/auth/otp/verify", rateLimit(idempotency(http.HandlerFunc(s.authHandler.VerifyOTP))))
	mux.Handle("POST /v1/auth/token/refresh", requireAuth(http.HandlerFunc(s.authHandler.RefreshToken)))

	// 2. Tenant Routes (Task 1.5.5)
	mux.Handle("GET /v1/tenants/me", requireAuth(http.HandlerFunc(s.handleGetTenantMe)))
	mux.Handle("PATCH /v1/tenants/me", requireAuth(idempotency(http.HandlerFunc(s.handleUpdateTenantMe))))
	mux.Handle("POST /v1/tenants/me/invite", requireAuth(idempotency(http.HandlerFunc(s.handleInviteUser))))

	// 3. Account Routes (Task 1.5.6)
	mux.Handle("GET /v1/accounts", requireAuth(http.HandlerFunc(s.handleListAccounts)))
	mux.Handle("POST /v1/accounts", requireAuth(idempotency(http.HandlerFunc(s.handleCreateAccount))))
	mux.Handle("GET /v1/accounts/{id}", requireAuth(http.HandlerFunc(s.handleGetAccountByID)))
	mux.Handle("PATCH /v1/accounts/{id}", requireAuth(idempotency(http.HandlerFunc(s.handleUpdateAccount))))
	mux.Handle("DELETE /v1/accounts/{id}", requireAuth(http.HandlerFunc(s.handleDeleteAccount)))

	// 4. Category Routes (Task 1.5.7)
	mux.Handle("GET /v1/categories", requireAuth(http.HandlerFunc(s.handleListCategories)))
	mux.Handle("POST /v1/categories", requireAuth(idempotency(http.HandlerFunc(s.handleCreateCategory))))
	mux.Handle("GET /v1/categories/{id}", requireAuth(http.HandlerFunc(s.handleGetCategoryByID)))
	mux.Handle("PATCH /v1/categories/{id}", requireAuth(idempotency(http.HandlerFunc(s.handleUpdateCategory))))
	mux.Handle("DELETE /v1/categories/{id}", requireAuth(http.HandlerFunc(s.handleDeleteCategory)))

	// 5. Transaction Routes (Task 1.5.8)
	mux.Handle("GET /v1/transactions", requireAuth(http.HandlerFunc(s.handleListTransactions)))
	mux.Handle("POST /v1/transactions", requireAuth(idempotency(http.HandlerFunc(s.handleCreateTransaction))))
	mux.Handle("GET /v1/transactions/{id}", requireAuth(http.HandlerFunc(s.handleGetTransactionByID)))
	mux.Handle("PATCH /v1/transactions/{id}", requireAuth(idempotency(http.HandlerFunc(s.handleUpdateTransaction))))
	mux.Handle("DELETE /v1/transactions/{id}", requireAuth(http.HandlerFunc(s.handleDeleteTransaction)))

	// 6. Admin Routes (Task 1.5.9)
	sysadminOnly := middleware.RequireRole(domain.RoleSysadmin)
	mux.Handle("GET /v1/admin/tenants", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminListTenants))))
	mux.Handle("GET /v1/admin/tenants/{id}", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminGetTenant))))
	mux.Handle("PATCH /v1/admin/tenants/{id}/plan", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminUpdatePlan))))
	mux.Handle("POST /v1/admin/tenants/{id}/suspend", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminSuspendTenant))))
	mux.Handle("POST /v1/admin/tenants/{id}/restore", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminRestoreTenant))))
	mux.Handle("DELETE /v1/admin/tenants/{id}", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminHardDeleteTenant))))
	mux.Handle("GET /v1/admin/users", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminListUsers))))
	mux.Handle("DELETE /v1/admin/users/{id}", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminForceDeleteUser))))
	mux.Handle("GET /v1/admin/audit-logs", requireAuth(sysadminOnly(http.HandlerFunc(s.handleAdminListAuditLogs))))

	return mux
}

func (s *Server) handleGetTenantMe(w http.ResponseWriter, r *http.Request) {
	s.tenantHandler.GetMe(w, r)
}

func (s *Server) handleUpdateTenantMe(w http.ResponseWriter, r *http.Request) {
	s.tenantHandler.UpdateMe(w, r)
}

func (s *Server) handleInviteUser(w http.ResponseWriter, r *http.Request) {
	s.tenantHandler.InviteUser(w, r)
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	s.accountHandler.List(w, r)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	s.accountHandler.Create(w, r)
}

func (s *Server) handleGetAccountByID(w http.ResponseWriter, r *http.Request) {
	s.accountHandler.GetByID(w, r)
}

func (s *Server) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	s.accountHandler.Update(w, r)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	s.accountHandler.Delete(w, r)
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	s.categoryHandler.List(w, r)
}

func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	s.categoryHandler.Create(w, r)
}

func (s *Server) handleGetCategoryByID(w http.ResponseWriter, r *http.Request) {
	s.categoryHandler.GetByID(w, r)
}

func (s *Server) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	s.categoryHandler.Update(w, r)
}

func (s *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	s.categoryHandler.Delete(w, r)
}

func (s *Server) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	s.transactionHandler.List(w, r)
}

func (s *Server) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	s.transactionHandler.Create(w, r)
}

func (s *Server) handleGetTransactionByID(w http.ResponseWriter, r *http.Request) {
	s.transactionHandler.GetByID(w, r)
}

func (s *Server) handleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	s.transactionHandler.Update(w, r)
}

func (s *Server) handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	s.transactionHandler.Delete(w, r)
}

func (s *Server) handleAdminListTenants(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.ListTenants(w, r)
}

func (s *Server) handleAdminGetTenant(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.GetTenant(w, r)
}

func (s *Server) handleAdminUpdatePlan(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.UpdateTenantPlan(w, r)
}

func (s *Server) handleAdminSuspendTenant(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.SuspendTenant(w, r)
}

func (s *Server) handleAdminRestoreTenant(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.RestoreTenant(w, r)
}

func (s *Server) handleAdminHardDeleteTenant(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.HardDeleteTenant(w, r)
}

func (s *Server) handleAdminListUsers(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.ListUsers(w, r)
}

func (s *Server) handleAdminForceDeleteUser(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.ForceDeleteUser(w, r)
}

func (s *Server) handleAdminListAuditLogs(w http.ResponseWriter, r *http.Request) {
	s.adminHandler.ListAuditLogs(w, r)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
