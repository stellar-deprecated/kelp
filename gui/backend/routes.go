package backend

import (
	"net/http"

	"github.com/go-chi/chi"
)

// SetRoutes
func SetRoutes(r *chi.Mux, s *APIServer) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", http.HandlerFunc(s.version))
		r.Get("/listBots", http.HandlerFunc(s.listBots))
		r.Get("/autogenerate", http.HandlerFunc(s.autogenerateBot))
		r.Post("/start", http.HandlerFunc(s.startBot))
		r.Post("/stop", http.HandlerFunc(s.stopBot))
		r.Post("/deleteBot", http.HandlerFunc(s.deleteBot))
		r.Post("/getState", http.HandlerFunc(s.getBotState))
		r.Post("/getBotInfo", http.HandlerFunc(s.getBotInfo))
	})
}
