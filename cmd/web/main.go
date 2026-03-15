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
	"github.com/garnizeh/moolah/pkg/logger"
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

	mux := buildMux(cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.WebPort,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

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

	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
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
func buildMux(_ *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	// Static assets — served directly from the embedded FS.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(web.StaticFS)))

	// Health check — used by load balancers and Docker health checks.
	mux.HandleFunc("GET /healthz", handleHealthz)

	return mux
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.Error("failed to write healthz response", "err", err)
	}
}
