// Package server implements the HTTP server for the Moolah API. It defines the Server struct,
// which holds references to all services and the HTTP handler. The server is responsible for
// starting and shutting down the HTTP server, as well as defining the routes in a separate file (routes.go).
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/handler"
	"github.com/garnizeh/moolah/internal/platform/middleware"
)

// Server is the main HTTP server for the API. It holds references to all services and the HTTP handler.
type Server struct {
	// Handler and service interfaces grouped for optimal field alignment
	handler http.Handler

	// Internal http server for lifecycle management
	httpServer *http.Server

	// Handlers
	authHandler           *handler.AuthHandler
	tenantHandler         *handler.TenantHandler
	accountHandler        *handler.AccountHandler
	categoryHandler       *handler.CategoryHandler
	transactionHandler    *handler.TransactionHandler
	masterPurchaseHandler *handler.MasterPurchaseHandler
	adminHandler          *handler.AdminHandler

	// Services passed to routes.go
	authSvc           domain.AuthService
	tenantSvc         domain.TenantService
	accountSvc        domain.AccountService
	categorySvc       domain.CategoryService
	transactionSvc    domain.TransactionService
	masterPurchaseSvc domain.MasterPurchaseService
	invoiceCloser     domain.InvoiceCloser
	adminSvc          domain.AdminService

	// Additional providers for middleware
	idempotencyStore domain.IdempotencyStore
	rateLimiterStore *middleware.RateLimiterStore
	tokenParser      middleware.TokenParser

	addr string
	mu   sync.Mutex
}

// NewServer is a factory function that creates and configures a new Server instance.
func New(
	port string,
	authSvc domain.AuthService,
	tenantSvc domain.TenantService,
	accountSvc domain.AccountService,
	categorySvc domain.CategoryService,
	transactionSvc domain.TransactionService,
	masterPurchaseSvc domain.MasterPurchaseService,
	invoiceCloser domain.InvoiceCloser,
	adminSvc domain.AdminService,
	idempotencyStore domain.IdempotencyStore,
	rateLimiterStore *middleware.RateLimiterStore,
	tokenParser middleware.TokenParser,
) *Server {
	s := &Server{
		addr:                  ":" + port,
		authHandler:           handler.NewAuthHandler(authSvc),
		tenantHandler:         handler.NewTenantHandler(tenantSvc),
		accountHandler:        handler.NewAccountHandler(accountSvc, invoiceCloser),
		categoryHandler:       handler.NewCategoryHandler(categorySvc),
		transactionHandler:    handler.NewTransactionHandler(transactionSvc),
		masterPurchaseHandler: handler.NewMasterPurchaseHandler(masterPurchaseSvc),
		adminHandler:          handler.NewAdminHandler(adminSvc),
		authSvc:               authSvc,
		tenantSvc:             tenantSvc,
		accountSvc:            accountSvc,
		categorySvc:           categorySvc,
		transactionSvc:        transactionSvc,
		masterPurchaseSvc:     masterPurchaseSvc,
		invoiceCloser:         invoiceCloser,
		adminSvc:              adminSvc,
		idempotencyStore:      idempotencyStore,
		rateLimiterStore:      rateLimiterStore,
		tokenParser:           tokenParser,
	}

	s.handler = middleware.RequestLogger()(s.routes())

	return s
}

// Handler returns the underlying http.Handler for testing purposes.
func (s *Server) Handler() http.Handler {
	return s.handler
}

// ListenAndServe starts the HTTP server with the configured address and handler. It also sets the read and write timeouts.
func (s *Server) ListenAndServe(ctx context.Context, readTimeout, writeTimeout time.Duration) error {
	// Security: set ReadHeaderTimeout and IdleTimeout to mitigate Slowloris (gosec G112)
	const (
		defaultReadHeaderTimeout = 5 * time.Second
		defaultIdleTimeout       = 120 * time.Second
	)

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.handler,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	s.mu.Lock()
	s.httpServer = srv
	s.mu.Unlock()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server, allowing in-flight requests to complete within the given context timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	// Close middleware stores if they have a Close method
	if s.rateLimiterStore != nil {
		s.rateLimiterStore.Close()
	}

	s.mu.Lock()
	srv := s.httpServer
	// avoid leaving a stale pointer; set to nil to indicate shutdown in progress/completed
	s.httpServer = nil
	s.mu.Unlock()

	if srv == nil {
		return nil
	}

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
