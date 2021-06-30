package backend

import (
	"net/http"

	"github.com/go-chi/chi"
)

// SetRoutes adds the handlers for the endpoints
func SetRoutes(r *chi.Mux, s *APIServer) {
	r.Route("/api/v1", func(r chi.Router) {
		if !s.enableKaas {
			// /quit is only enabled when we are not in KaaS mode
			r.Get("/quit", http.HandlerFunc(s.quit))
		}

		r.Get("/version", http.HandlerFunc(s.version))
		r.Get("/serverMetadata", http.HandlerFunc(s.serverMetadata))
		r.Get("/newSecretKey", http.HandlerFunc(s.newSecretKey))
		r.Get("/optionsMetadata", http.HandlerFunc(s.optionsMetadata))
		var router chi.Router = r
		if s.guiConfig.Auth0Config != nil && s.guiConfig.Auth0Config.Auth0Enabled {
			// setting the router to use the JWT middleware to handle auth0 style JWT tokens
			router = r.With(JWTMiddlewareVar.Handler)
		}

		router.Post("/listBots", http.HandlerFunc(s.listBots))
		router.Post("/genBotName", http.HandlerFunc(s.generateBotName))
		router.Post("/getNewBotConfig", http.HandlerFunc(s.getNewBotConfig))
		router.Post("/autogenerate", http.HandlerFunc(s.autogenerateBot))
		router.Post("/fetchKelpErrors", http.HandlerFunc(s.fetchKelpErrors))
		router.Post("/removeKelpErrors", http.HandlerFunc(s.removeKelpErrors))
		router.Post("/start", http.HandlerFunc(s.startBot))
		router.Post("/stop", http.HandlerFunc(s.stopBot))
		router.Post("/deleteBot", http.HandlerFunc(s.deleteBot))
		router.Post("/getState", http.HandlerFunc(s.getBotState))
		router.Post("/getBotInfo", http.HandlerFunc(s.getBotInfo))
		router.Post("/getBotConfig", http.HandlerFunc(s.getBotConfig))
		router.Post("/fetchPrice", http.HandlerFunc(s.fetchPrice))
		router.Post("/upsertBotConfig", http.HandlerFunc(s.upsertBotConfig))
		router.Post("/sendMetricEvent", http.HandlerFunc(s.sendMetricEvent))
	})
	r.Get("/ping", http.HandlerFunc(s.ping))
}