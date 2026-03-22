package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garnizeh/moolah/internal/api/middleware"
	"github.com/garnizeh/moolah/internal/platform/log"
)

func main() {
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

	go func() {
		slog.Info("starting server", slog.String("port", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for signal
	<-stop

	slog.Info("shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", slog.Any("error", err))
	} else {
		slog.Info("server exited gracefully")
	}
}
