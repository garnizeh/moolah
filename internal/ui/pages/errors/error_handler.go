package errors

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/garnizeh/moolah/internal/ui/components"
	"github.com/garnizeh/moolah/internal/ui/layout"
)

// RenderError writes the appropriate error response, respecting HTMX partial requests.
func RenderError(w http.ResponseWriter, r *http.Request, status int, props layout.BaseProps) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	isHTMX := r.Header.Get("HX-Request") == "true"

	if isHTMX {
		// Return a toast fragment instead of a full error page
		w.Header().Set("HX-Retarget", "#toast-container")
		w.Header().Set("HX-Reswap", "beforeend")
		w.WriteHeader(status)

		toastProps := components.ToastProps{
			Variant: components.ToastError,
			Title:   http.StatusText(status),
			Message: getErrorMessage(status),
		}
		if err := components.Toast(toastProps).Render(r.Context(), w); err != nil {
			// If rendering toast fails, we log it but we can't do much else
			// because we already wrote headers.
			return
		}
		return
	}

	w.WriteHeader(status)
	var content templ.Component
	switch status {
	case http.StatusNotFound:
		content = NotFound(props)
	case http.StatusForbidden:
		content = Forbidden(props)
	default:
		content = InternalError(props)
	}

	// Wrap in layout if needed
	// For errors, we generally want the base layout if authenticated
	props.Content = content
	if props.Title == "" {
		props.Title = http.StatusText(status)
	}

	// We use the same layout as the rest of the app
	// If the user is unauthenticated, the layout might need to be different,
	// but for now we follow the task's base layout requirement.
	if err := layout.Base(props).Render(r.Context(), w); err != nil {
		// Fallback for extreme failure during rendering
		http.Error(w, http.StatusText(status), status)
	}
}

func getErrorMessage(status int) string {
	switch status {
	case http.StatusNotFound:
		return "The page you're looking for doesn't exist."
	case http.StatusForbidden:
		return "You don't have permission to perform this action."
	default:
		return "An unexpected error occurred. Please try again."
	}
}

// Helper to get base props from context or create empty ones
func BasePropsFromRequest(r *http.Request) layout.BaseProps {
	// In a real app, we'd extract user/tenant from context if present
	return layout.BaseProps{
		CurrentPath: r.URL.Path,
	}
}
