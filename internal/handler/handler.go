// Package handler provides HTTP handlers for the application.
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func respondJSON(w http.ResponseWriter, r *http.Request, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
	}
}

func respondError(w http.ResponseWriter, r *http.Request, message string, status int) {
	respondJSON(w, r, map[string]string{"error": message}, status)
}
