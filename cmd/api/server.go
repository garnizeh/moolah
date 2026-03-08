package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/garnizeh/moolah/internal/domain"
)

type Server struct {
	addr    string
	handler http.Handler

	// Services passed to routes.go
	authSvc        domain.AuthService
	tenantSvc      domain.TenantService
	accountSvc     domain.AccountService
	categorySvc    domain.CategoryService
	transactionSvc domain.TransactionService
	adminSvc       domain.AdminService
}

func NewServer(
	port string,
	authSvc domain.AuthService,
	tenantSvc domain.TenantService,
	accountSvc domain.AccountService,
	categorySvc domain.CategoryService,
	transactionSvc domain.TransactionService,
	adminSvc domain.AdminService,
) *Server {
	s := &Server{
		addr:           ":" + port,
		authSvc:        authSvc,
		tenantSvc:      tenantSvc,
		accountSvc:     accountSvc,
		categorySvc:    categorySvc,
		transactionSvc: transactionSvc,
		adminSvc:       adminSvc,
	}

	// routes.go will implement the routes method
	s.handler = s.routes()

	return s
}

func (s *Server) ListenAndServe(ctx context.Context, readTimeout, writeTimeout time.Duration) error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	slog.InfoContext(ctx, "starting server", "addr", s.addr)
	return srv.ListenAndServe()
}
