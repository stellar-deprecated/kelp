package backend

import (
	"net/http"

	"github.com/go-chi/chi"
)

// SetRoutes
func SetRoutes(r *chi.Mux, s *APIServer) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", http.HandlerFunc(s.version))
	})
}
