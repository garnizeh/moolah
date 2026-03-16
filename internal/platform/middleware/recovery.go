package middleware

import (
	"log/slog"
	"net/http"

	"github.com/garnizeh/moolah/internal/ui/pages/errors"
)

// Recovery recovers from panics in handlers and renders a 500 internal server error page.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(ctx, "panic recovered",
					"path", r.URL.Path,
					"method", r.Method,
					"panic", rec,
				)

				// Ensure we haven't already written headers
				// If headers are already written, there's not much we can do
				// besides logging but we attempt to render if possible.
				errors.RenderError(w, r, http.StatusInternalServerError, errors.BasePropsFromRequest(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
