package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type getNewBotConfigRequest struct {
	UserData UserData `json:"user_data"`
}

func (s *APIServer) getNewBotConfig(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req getNewBotConfigRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}

	botName, e := s.doGenerateBotName(req.UserData)
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
