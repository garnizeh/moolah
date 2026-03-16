package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	t.Parallel()

	t.Run("Handle Panic and Render 500", func(t *testing.T) {
		t.Parallel()

		// Handler that panics
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		recoveryHandler := Recovery(handler)

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rr := httptest.NewRecorder()

		// The test should not crash
		assert.NotPanics(t, func() {
			recoveryHandler.ServeHTTP(rr, req)
		})

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "An unexpected error occurred")
	})

	t.Run("Pass Through Normal Request", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("ok"))
			assert.NoError(t, err)
		})

		recoveryHandler := Recovery(handler)

		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		rr := httptest.NewRecorder()

		recoveryHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "ok", rr.Body.String())
	})

	t.Run("HTMX Panic returns toast", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("htmx panic")
		})

		recoveryHandler := Recovery(handler)

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		req.Header.Set("HX-Request", "true")
		rr := httptest.NewRecorder()

		recoveryHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "#toast-container", rr.Header().Get("HX-Retarget"))
		assert.Contains(t, rr.Body.String(), "An unexpected error occurred")
	})
}
