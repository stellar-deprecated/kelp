package backend

import (
	"net/http"

	"github.com/go-chi/chi"
)

// SetRoutes
func SetRoutes(r *chi.Mux, s *APIServer) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", http.HandlerFunc(s.version))
		r.Get("/quit", http.HandlerFunc(s.quit))
		r.Get("/listBots", http.HandlerFunc(s.listBots))
		r.Get("/autogenerate", http.HandlerFunc(s.autogenerateBot))
		r.Get("/genBotName", http.HandlerFunc(s.generateBotName))
		r.Get("/getNewBotConfig", http.HandlerFunc(s.getNewBotConfig))
		r.Get("/newSecretKey", http.HandlerFunc(s.newSecretKey))
		r.Get("/optionsMetadata", http.HandlerFunc(s.optionsMetadata))

		r.Post("/start", http.HandlerFunc(s.startBot))
		r.Post("/stop", http.HandlerFunc(s.stopBot))
		r.Post("/deleteBot", http.HandlerFunc(s.deleteBot))
		r.Post("/getState", http.HandlerFunc(s.getBotState))
		r.Post("/getBotInfo", http.HandlerFunc(s.getBotInfo))
		r.Post("/getBotConfig", http.HandlerFunc(s.getBotConfig))
		r.Post("/fetchPrice", http.HandlerFunc(s.fetchPrice))
		r.Post("/upsertBotConfig", http.HandlerFunc(s.upsertBotConfig))
	})
	r.Get("/ping", http.HandlerFunc(s.ping))
}
