// Package handler provides HTTP handlers for the application.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/garnizeh/moolah/internal/domain"
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

func handleError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, r, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrForbidden):
		respondError(w, r, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrConflict):
		respondError(w, r, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrInvalidInput):
		respondError(w, r, err.Error(), http.StatusUnprocessableEntity)
	case errors.Is(err, domain.ErrInvalidOTP):
		respondError(w, r, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrOTPRateLimited):
		respondError(w, r, err.Error(), http.StatusTooManyRequests)
	case errors.Is(err, domain.ErrTokenExpired), errors.Is(err, domain.ErrUnauthorized):
		respondError(w, r, "invalid or expired token", http.StatusUnauthorized)
	default:
		slog.ErrorContext(r.Context(), msg, "error", err)
		respondError(w, r, "internal error", http.StatusInternalServerError)
	}
}
