package server

import (
	"fmt"
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Placeholder routes until Task 1.5.4+ handlers are built
	mux.HandleFunc("/healthz", s.handleHealthz)

	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
