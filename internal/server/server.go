// Package server implements the HTTP server for the Moolah API. It defines the Server struct,
// which holds references to all services and the HTTP handler. The server is responsible for
// starting and shutting down the HTTP server, as well as defining the routes in a separate file (routes.go).
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
	"github.com/garnizeh/moolah/internal/handler"
	"github.com/garnizeh/moolah/internal/platform/middleware"
)

// Server is the main HTTP server for the API. It holds references to all services and the HTTP handler.
type Server struct {
	// Handler and service interfaces grouped for optimal field alignment
	handler http.Handler

	// Handlers
	authHandler *handler.AuthHandler

	// Services passed to routes.go
	authSvc        domain.AuthService
	tenantSvc      domain.TenantService
	accountSvc     domain.AccountService
	categorySvc    domain.CategoryService
	transactionSvc domain.TransactionService
	adminSvc       domain.AdminService

	// Middleware dependencies
	idempotencyStore domain.IdempotencyStore
	rateLimiterStore *middleware.RateLimiterStore
	tokenParser      middleware.TokenParser

	addr string
}

// NewServer is a factory function that creates and configures a new Server instance.
func New(
	port string,
	authSvc domain.AuthService,
	tenantSvc domain.TenantService,
	accountSvc domain.AccountService,
	categorySvc domain.CategoryService,
	transactionSvc domain.TransactionService,
	adminSvc domain.AdminService,
	idempotencyStore domain.IdempotencyStore,
	rateLimiterStore *middleware.RateLimiterStore,
	tokenParser middleware.TokenParser,
) *Server {
	s := &Server{
		addr:             ":" + port,
		authHandler:      handler.NewAuthHandler(authSvc, slog.Default()),
		authSvc:          authSvc,
		tenantSvc:        tenantSvc,
		accountSvc:       accountSvc,
		categorySvc:      categorySvc,
		transactionSvc:   transactionSvc,
		adminSvc:         adminSvc,
		idempotencyStore: idempotencyStore,
		rateLimiterStore: rateLimiterStore,
		tokenParser:      tokenParser,
	}

	// routes.go will implement the routes method
	// apply global logger middleware using the default slog logger
	s.handler = middleware.RequestLogger(slog.Default())(s.routes())

	return s
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

	slog.InfoContext(ctx, "starting server", "addr", s.addr)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.ErrorContext(ctx, "server failed", "err", err)
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP server, allowing in-flight requests to complete within the given context timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	// Mirror the timeouts used when serving to ensure server struct has safe defaults
	const (
		defaultReadHeaderTimeout = 5 * time.Second
		defaultIdleTimeout       = 120 * time.Second
	)

	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.handler,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}

	slog.InfoContext(ctx, "shutting down server", "addr", s.addr)
	err := srv.Shutdown(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "server shutdown failed", "err", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
