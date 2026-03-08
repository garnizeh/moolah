package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"aidanwoods.dev/go-paseto"
	"github.com/garnizeh/moolah/internal/platform/db/migrations"
	"github.com/garnizeh/moolah/internal/platform/db/sqlc"
	"github.com/garnizeh/moolah/internal/platform/idempotency"
	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/server"
	"github.com/garnizeh/moolah/internal/service"
	"github.com/garnizeh/moolah/pkg/config"
	"github.com/garnizeh/moolah/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// 1. Load Config
	cfg := config.Load()

	// 2. Init Logger
	l := logger.New(nil, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(l)

	// Connect DB, run migrations, and create sqlc querier
	querier, dbPool, err := db.Querier(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// 5. Connect Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// 6. Wire Repositories
	querier := sqlc.New(dbPool)

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

	// 7. Wire Services
	authSvc := service.NewAuthService(authRepo, userRepo, auditRepo, smtpMailer, pasetoKey)
	tenantSvc := service.NewTenantService(tenantRepo, userRepo, auditRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo, auditRepo)
	categorySvc := service.NewCategoryService(categoryRepo, auditRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, accountRepo, categoryRepo, auditRepo)
	adminSvc := service.NewAdminService(adminTenantRepo, adminUserRepo, adminAuditRepo, auditRepo, l)

	// Avoid unused variable warnings until 1.5.2/1.5.3 are implemented
	_ = authSvc
	_ = tenantSvc
	_ = accountSvc
	_ = categorySvc
	_ = transactionSvc
	_ = adminSvc
	_ = idempotencyStore

	// 8. Create Server
	srv := server.New(
		cfg.HTTPPort,
		authSvc,
		tenantSvc,
		accountSvc,
		categorySvc,
		transactionSvc,
		adminSvc,
	)

	// 9. Graceful Shutdown
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

	// 10. Start Server
	slog.Info("starting server", "port", cfg.HTTPPort)
	if err := srv.ListenAndServe(ctx, cfg.ReadTimeout, cfg.WriteTimeout); err != http.ErrServerClosed {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}

	<-idleConnsClosed
	slog.Info("server stopped")
}

func runMigrations(dbPool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(dbPool)
	defer db.Close()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run goose migrations: %w", err)
	}

	return nil
}
