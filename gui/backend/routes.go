package backend

import (
	"net/http"

	"github.com/go-chi/chi"
)

// SetRoutes
func SetRoutes(r *chi.Mux) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/one", http.HandlerFunc(one))
		r.Get("/two", http.HandlerFunc(two))
	})
}
