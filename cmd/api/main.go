package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/garnizeh/moolah/internal/api/middleware"
	"github.com/garnizeh/moolah/internal/platform/log"
)

func main() {
	// Initialize logging
	log.Init()

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "request_id": "%s"}`, middleware.FromContext(r.Context()))
	})

	// Wrap with middleware
	handler := middleware.RequestID(mux)

	port := "8080"
	slog.Info("starting server", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		slog.Error("server failed", slog.Any("error", err))
	}
}
