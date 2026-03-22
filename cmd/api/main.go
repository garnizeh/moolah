package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/garnizeh/moolah/internal/api/middleware"
	"github.com/garnizeh/moolah/internal/platform/log"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("application failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Initialize logging
	log.InitWithWriter(os.Stdout)

	mux := http.NewServeMux()

	// Health check
	type healthResponse struct {
		Status    string `json:"status"`
		RequestID string `json:"request_id"`
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := healthResponse{
			Status:    "ok",
			RequestID: middleware.FromContext(r.Context()),
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("failed to encode health response", slog.Any("error", err))
		}
	})

	// Wrap with middleware
	handler := middleware.RequestID(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Sanitize port for logging to prevent log injection (Gosec G706)
	cleanPort := "0"
	if p, err := strconv.Atoi(port); err == nil {
		cleanPort = strconv.Itoa(p)
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("starting server", slog.String("port", cleanPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Wait for signal or error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-stop:
		slog.Info("shutting down server...", slog.Any("signal", sig))

		// Create context with timeout for shutdown
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				return fmt.Errorf("could not stop server gracefully: %w (close error: %v)", err, closeErr)
			}
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	slog.Info("server exited gracefully")
	return nil
}
