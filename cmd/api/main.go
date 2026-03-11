package main

// @title           Moolah Financial API
// @version         1.0
// @description     Moolah Household Finance & Investment SaaS API.
// @contact.name    API Support
// @host            localhost:8080
// @BasePath        /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/garnizeh/moolah/internal/platform/bootstrap"
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

var (
	tagVersion = "v0.0.0-dev"
	buildTime  = "development"
	commitHash = "development"
	goVersion  = "development"
)

func main() {
	ctx := context.Background()

	// CLI flags
	showConfig := flag.Bool("show-config", false, "print loaded config and exit")
	flag.Parse()

	// Load Config
	cfg := config.Load()

	// Init Logger
	log := logger.New(nil, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	err := run(ctx, cfg, *showConfig)
	if err != nil {
		slog.ErrorContext(ctx, "application error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, showConfig bool) error {
	slog.InfoContext(ctx, "starting application",
		"version", tagVersion,
		"buildTime", buildTime,
		"commitHash", commitHash,
		"goVersion", goVersion,
	)

	if showConfig {
		cfg.Log(ctx)
	}

	// Connect DB, run migrations, and create sqlc querier
	pool, querier, err := db.Querier(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}
	defer pool.Close()

	// Bootstrap Sysadmin
	if err := bootstrap.EnsureSysadmin(ctx, querier, cfg); err != nil {
		return fmt.Errorf("sysadmin bootstrap failed: %w", err)
	}

	// Connect Redis
	rdb, err := redis.NewClient(ctx, cfg.RedisAddr, cfg.RedisPassword, 0)
	if err != nil {
		return fmt.Errorf("redis initialization failed: %w", err)
	}
	defer rdb.Close()

	// Initialize Mailer
	smtpMailer, err := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.EmailFrom)
	if err != nil {
		return fmt.Errorf("mailer initialization failed: %w", err)
	}

	// Parse PASETO key
	pasetoKey, err := paseto.V4SymmetricKeyFromHex(cfg.PasetoSecretKey)
	if err != nil {
		return fmt.Errorf("failed to parse paseto secret key: %w", err)
	}

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

	// Wire Services
	authSvc := service.NewAuthService(authRepo, userRepo, auditRepo, smtpMailer, pasetoKey)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, auditRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo, auditRepo)
	categorySvc := service.NewCategoryService(categoryRepo, auditRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo, auditRepo)
	adminSvc := service.NewAdminService(adminTenantRepo, adminUserRepo, adminAuditRepo, auditRepo)

	rateLimiterStore := middleware.NewRateLimiterStore()
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

		ctxwt, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		slog.InfoContext(ctxwt, "shutting down server")
		if err := srv.Shutdown(ctxwt); err != nil {
			slog.ErrorContext(ctxwt, "server shutdown failed", "err", err)
		}
		close(idleConnsClosed)
	}()

	// Start Server
	slog.InfoContext(ctx, "starting server", "port", cfg.HTTPPort)
	if err := srv.ListenAndServe(ctx, cfg.ReadTimeout, cfg.WriteTimeout); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	<-idleConnsClosed
	slog.InfoContext(ctx, "server stopped")

	return nil
}
