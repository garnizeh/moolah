package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/garnizeh/moolah/internal/platform/db"
	"github.com/garnizeh/moolah/internal/platform/idempotency"
	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/platform/redis"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/server"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/pkg/config"
	"github.com/garnizeh/moolah/pkg/logger"
	"github.com/garnizeh/moolah/pkg/paseto"
)

func main() {
	ctx := context.Background()

	// Load Config
	cfg := config.Load()

	// Init Logger
	l := logger.New(nil, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(l)

	// Connect DB, run migrations, and create sqlc querier
	pool, querier, err := db.Querier(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Connect Redis
	rdb, err := redis.NewClient(ctx, cfg.RedisAddr, cfg.RedisPassword, 0)
	if err != nil {
		slog.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// Wire Repositories
	authRepo := repository.NewAuthRepository(querier)
	tenantRepo := repository.NewTenantRepository(querier)
	userRepo := repository.NewUserRepository(querier)
	accountRepo := repository.NewAccountRepository(querier)
	categoryRepo := repository.NewCategoryRepository(querier)
	transactionRepo := repository.NewTransactionRepository(querier)
	auditRepo := repository.NewAuditRepository(querier)

	adminTenantRepo := repository.NewAdminTenantRepository(querier)
	adminUserRepo := repository.NewAdminUserRepository(querier)
	adminAuditRepo := repository.NewAdminAuditRepository(querier)

	idempotencyStore := idempotency.NewRedisStore(rdb)

	smtpMailer, err := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.EmailFrom)
	if err != nil {
		slog.Error("failed to initialize mailer", "err", err)
		os.Exit(1)
	}

	pasetoKey, err := paseto.V4SymmetricKeyFromHex(cfg.PasetoSecretKey)
	if err != nil {
		slog.Error("failed to parse paseto secret key", "err", err)
		os.Exit(1)
	}

	// Wire Services
	authSvc := service.NewAuthService(authRepo, userRepo, auditRepo, smtpMailer, pasetoKey)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, auditRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo, auditRepo)
	categorySvc := service.NewCategoryService(categoryRepo, auditRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo, auditRepo)
	adminSvc := service.NewAdminService(adminTenantRepo, adminUserRepo, adminAuditRepo, auditRepo, l)

	rateLimiterStore := middleware.NewRateLimiterStore(l)
	tokenParser := paseto.NewTokenParser(pasetoKey)

	// Create Server
	srv := server.New(
		cfg.HTTPPort,
		authSvc,
		tenantSvc,
		accountSvc,
		categorySvc,
		transactionSvc,
		adminSvc,
		idempotencyStore,
		rateLimiterStore,
		tokenParser,
	)

	// Graceful Shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		slog.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown failed", "err", err)
		}
		close(idleConnsClosed)
	}()

	// Start Server
	slog.Info("starting server", "port", cfg.HTTPPort)
	if err := srv.ListenAndServe(ctx, cfg.ReadTimeout, cfg.WriteTimeout); err != http.ErrServerClosed {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}

	<-idleConnsClosed
	slog.Info("server stopped")
}
