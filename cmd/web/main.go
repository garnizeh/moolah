package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garnizeh/moolah/internal/config"
	"github.com/garnizeh/moolah/internal/platform/db"
	"github.com/garnizeh/moolah/internal/platform/mailer"
	"github.com/garnizeh/moolah/internal/platform/middleware"
	"github.com/garnizeh/moolah/internal/platform/repository"
	"github.com/garnizeh/moolah/internal/platform/ws"
	"github.com/garnizeh/moolah/internal/service"
	uimiddleware "github.com/garnizeh/moolah/internal/ui/middleware"
	"github.com/garnizeh/moolah/internal/ui/pages/auth"
	"github.com/garnizeh/moolah/internal/ui/pages/errors"
	"github.com/garnizeh/moolah/pkg/logger"
	"github.com/garnizeh/moolah/pkg/paseto"
	"github.com/garnizeh/moolah/web"
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
	l := logger.New(nil, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(l)

	if err := run(ctx, cfg, l, *showConfig); err != nil {
		slog.ErrorContext(ctx, "web server error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, _ *slog.Logger, showConfig bool) error {
	slog.InfoContext(ctx, "starting web server",
		"version", tagVersion,
		"buildTime", buildTime,
		"commitHash", commitHash,
		"goVersion", goVersion,
		"addr", ":"+cfg.WebPort,
	)

	if showConfig {
		cfg.Log(ctx)
		return nil
	}

	// Connect DB and create sqlc querier
	pool, querier, err := db.Querier(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}
	defer pool.Close()

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
	userRepo := repository.NewUserRepository(querier)
	auditRepo := repository.NewAuditRepository(querier)

	// Wire Services
	authSvc := service.NewAuthService(authRepo, userRepo, auditRepo, smtpMailer, pasetoKey)
	tokenParser := paseto.NewTokenParser(pasetoKey)

	// Wire Handlers
	authHandler := auth.NewAuthHandler(authSvc, cfg.IsDevelopment())

	// Initialize WebSocket Hub
	hub := ws.NewHub(10) // Max 10 connections per tenant
	go hub.Run(ctx)

	mux := buildMux(cfg, authHandler, tokenParser, hub)

	srv := &http.Server{
		Addr:         ":" + cfg.WebPort,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	serverErr := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "web server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("web server listen error: %w", err)
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.InfoContext(ctx, "web server context canceled", "err", ctx.Err())
	case <-quit:
		slog.InfoContext(ctx, "web server shutting down")
	}

	shutdownBase := context.WithoutCancel(ctx)
	shutdownCtx, cancel := context.WithTimeout(shutdownBase, cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("web server shutdown error: %w", err)
	}

	slog.InfoContext(ctx, "web server stopped")
	return nil
}

// buildMux constructs and returns the HTTP mux for the web UI server.
// Routes are registered in this function; subsequent tasks (4.3–4.9) will
// add page handlers here.
func buildMux(_ *config.Config, authHandler *auth.AuthHandler, tokenParser func(string) (*paseto.Claims, error), hub *ws.Hub) http.Handler {
	mux := http.NewServeMux()

	// Static assets — served directly from the embedded FS.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(web.StaticFS)))

	// Health check — used by load balancers and Docker health checks.
	mux.HandleFunc("GET /healthz", handleHealthz)

	// --- Public Routes ---
	redirectIfAuth := uimiddleware.RedirectIfAuthenticated(tokenParser, "/dashboard")

	mux.Handle("GET /web/login", redirectIfAuth(http.HandlerFunc(authHandler.Login)))
	mux.Handle("POST /web/auth/otp/request", http.HandlerFunc(authHandler.RequestOTP))
	mux.Handle("POST /web/auth/otp/verify", http.HandlerFunc(authHandler.VerifyOTP))

	// --- Protected Routes ---
	sessionAuth := uimiddleware.SessionAuth(tokenParser, "/web/login")

	mux.Handle("GET /dashboard", sessionAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// For dashboard, we'll implement a proper page later.
		if _, err := fmt.Fprint(w, "<h1>Dashboard</h1><p>Welcome!</p><form method='post' action='/web/auth/logout' hx-post='/web/auth/logout'><button type='submit'>Logout</button></form>"); err != nil {
			slog.ErrorContext(r.Context(), "failed to write dashboard response", "error", err)
		}
	})))

	mux.Handle("POST /web/auth/logout", sessionAuth(http.HandlerFunc(authHandler.Logout)))

	// WebSocket endpoint
	mux.Handle("GET /ws", sessionAuth(ws.UpgradeHandler(hub)))

	// Root redirect
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		errors.RenderError(w, r, http.StatusNotFound, errors.BasePropsFromRequest(r))
	}))

	return middleware.Recovery(mux)
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.Error("failed to write healthz response", "err", err)
	}
}
