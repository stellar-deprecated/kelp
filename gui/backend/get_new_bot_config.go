package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (s *APIServer) getNewBotConfig(w http.ResponseWriter, r *http.Request) {
	botName, e := s.doGenerateBotName()
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("cannot generate a new bot name: %s", e))
		return
	}
	sampleTrader := s.makeSampleTrader("")
	strategy := "buysell"
	sampleBuysell := makeSampleBuysell()

	// remove asset data from the trader file for the new config
	sampleTrader.AssetCodeA = ""
	sampleTrader.IssuerA = ""
	sampleTrader.AssetCodeB = ""
	sampleTrader.IssuerB = ""

	response := botConfigResponse{
		Name:           botName,
		Strategy:       strategy,
		TraderConfig:   *sampleTrader,
		StrategyConfig: *sampleBuysell,
	}
	jsonBytes, e := json.MarshalIndent(response, "", "  ")
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("cannot marshal botConfigResponse: %s\n", e))
		return
	}
	log.Printf("getNewBotConfig response: %s\n", string(jsonBytes))
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
