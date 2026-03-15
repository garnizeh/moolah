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
	sysadminOnly := middleware.RequireRole(domain.RoleSysadmin)
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
	mux.Handle("POST /v1/accounts/{id}/close-invoice", requireAuth(idempotency(http.HandlerFunc(s.handleCloseInvoice))))

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

	// 6. Master Purchase Routes (Phase 2)
	mux.Handle("GET /v1/master-purchases", requireAuth(http.HandlerFunc(s.handleListMasterPurchases)))
	mux.Handle("POST /v1/master-purchases", requireAuth(idempotency(http.HandlerFunc(s.handleCreateMasterPurchase))))
	mux.Handle("GET /v1/master-purchases/{id}", requireAuth(http.HandlerFunc(s.handleGetMasterPurchaseByID)))
	mux.Handle("GET /v1/master-purchases/{id}/projection", requireAuth(http.HandlerFunc(s.handleProjectMasterPurchase)))
	mux.Handle("PATCH /v1/master-purchases/{id}", requireAuth(idempotency(http.HandlerFunc(s.handleUpdateMasterPurchase))))
	mux.Handle("DELETE /v1/master-purchases/{id}", requireAuth(http.HandlerFunc(s.handleDeleteMasterPurchase)))

	// 7. Investment & Asset Routes (Task 3.7)
	mux.Handle("GET /v1/assets", requireAuth(http.HandlerFunc(s.handleListAssets)))
	mux.Handle("GET /v1/assets/{id}", requireAuth(http.HandlerFunc(s.handleGetAssetByID)))
	mux.Handle("POST /v1/assets", requireAuth(sysadminOnly(idempotency(http.HandlerFunc(s.handleCreateAsset)))))
	mux.Handle("DELETE /v1/assets/{id}", requireAuth(sysadminOnly(http.HandlerFunc(s.handleDeleteAsset))))
	mux.Handle("GET /v1/me/asset-configs", requireAuth(http.HandlerFunc(s.handleListMyAssetConfigs)))
	mux.Handle("PUT /v1/me/asset-configs/{asset_id}", requireAuth(idempotency(http.HandlerFunc(s.handleUpsertMyAssetConfig))))
	mux.Handle("DELETE /v1/me/asset-configs/{asset_id}", requireAuth(http.HandlerFunc(s.handleDeleteMyAssetConfig)))

	// 8. Position & Portfolio Routes (Task 3.14)
	mux.Handle("GET /v1/positions", requireAuth(http.HandlerFunc(s.handleListPositions)))
	mux.Handle("POST /v1/positions", requireAuth(idempotency(http.HandlerFunc(s.handleCreatePosition))))
	mux.Handle("GET /v1/positions/{id}", requireAuth(http.HandlerFunc(s.handleGetPositionByID)))
	mux.Handle("PATCH /v1/positions/{id}", requireAuth(idempotency(http.HandlerFunc(s.handleUpdatePosition))))
	mux.Handle("DELETE /v1/positions/{id}", requireAuth(http.HandlerFunc(s.handleDeletePosition)))
	mux.Handle("GET /v1/accounts/{id}/positions", requireAuth(http.HandlerFunc(s.handleListPositionsByAccount)))
	mux.Handle("GET /v1/income-events", requireAuth(http.HandlerFunc(s.handleListIncomeEvents)))
	mux.Handle("GET /v1/income-events/pending", requireAuth(http.HandlerFunc(s.handleListPendingIncomeEvents)))
	mux.Handle("PATCH /v1/income-events/{id}/receive", requireAuth(idempotency(http.HandlerFunc(s.handleMarkIncomeReceived))))
	mux.Handle("PATCH /v1/income-events/{id}/cancel", requireAuth(idempotency(http.HandlerFunc(s.handleCancelIncome))))
	mux.Handle("POST /v1/portfolio/snapshot", requireAuth(idempotency(http.HandlerFunc(s.handleTriggerSnapshot))))
	mux.Handle("GET /v1/investments/summary", requireAuth(http.HandlerFunc(s.handleGetSummary)))

	// 9. Admin Routes (Task 1.5.9)
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

func (s *Server) handleCloseInvoice(w http.ResponseWriter, r *http.Request) {
	s.accountHandler.CloseInvoice(w, r)
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

func (s *Server) handleListAssets(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.ListAssets(w, r)
}

func (s *Server) handleGetAssetByID(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.GetAsset(w, r)
}

func (s *Server) handleCreateAsset(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.CreateAsset(w, r)
}

func (s *Server) handleDeleteAsset(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.DeleteAsset(w, r)
}

func (s *Server) handleListMyAssetConfigs(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.ListMyConfigs(w, r)
}

func (s *Server) handleUpsertMyAssetConfig(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.UpsertMyConfig(w, r)
}

func (s *Server) handleDeleteMyAssetConfig(w http.ResponseWriter, r *http.Request) {
	s.assetHandler.DeleteMyConfig(w, r)
}

func (s *Server) handleListPositions(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.List(w, r)
}

func (s *Server) handleCreatePosition(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.Create(w, r)
}

func (s *Server) handleGetPositionByID(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.GetByID(w, r)
}

func (s *Server) handleUpdatePosition(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.Update(w, r)
}

func (s *Server) handleDeletePosition(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.Delete(w, r)
}

func (s *Server) handleListPositionsByAccount(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.ListByAccount(w, r)
}

func (s *Server) handleListIncomeEvents(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.ListIncomeEvents(w, r)
}

func (s *Server) handleListPendingIncomeEvents(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.ListPendingIncomeEvents(w, r)
}

func (s *Server) handleMarkIncomeReceived(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.MarkIncomeReceived(w, r)
}

func (s *Server) handleCancelIncome(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.CancelIncome(w, r)
}

func (s *Server) handleTriggerSnapshot(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.TriggerSnapshot(w, r)
}

func (s *Server) handleGetSummary(w http.ResponseWriter, r *http.Request) {
	s.positionHandler.GetSummary(w, r)
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

func (s *Server) handleListMasterPurchases(w http.ResponseWriter, r *http.Request) {
	s.masterPurchaseHandler.ListByTenant(w, r)
}

func (s *Server) handleCreateMasterPurchase(w http.ResponseWriter, r *http.Request) {
	s.masterPurchaseHandler.Create(w, r)
}

func (s *Server) handleGetMasterPurchaseByID(w http.ResponseWriter, r *http.Request) {
	s.masterPurchaseHandler.GetByID(w, r)
}

func (s *Server) handleProjectMasterPurchase(w http.ResponseWriter, r *http.Request) {
	s.masterPurchaseHandler.Project(w, r)
}

func (s *Server) handleUpdateMasterPurchase(w http.ResponseWriter, r *http.Request) {
	s.masterPurchaseHandler.Update(w, r)
}

func (s *Server) handleDeleteMasterPurchase(w http.ResponseWriter, r *http.Request) {
	s.masterPurchaseHandler.Delete(w, r)
}
