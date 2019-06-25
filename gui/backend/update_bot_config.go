package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stellar/kelp/gui/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/toml"
	"github.com/stellar/kelp/trader"
)

type updateBotConfigRequest struct {
	Name           string                `json:"name"`
	Strategy       string                `json:"strategy"`
	TraderConfig   trader.BotConfig      `json:"trader_config"`
	StrategyConfig plugins.BuySellConfig `json:"strategy_config"`
}

type updateBotConfigResponse struct {
	Success bool `json:"success"`
}

func (s *APIServer) updateBotConfig(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s", e))
		return
	}
	log.Printf("updateBotConfig requestJson: %s\n", string(bodyBytes))

	var req updateBotConfigRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}

	botState, e := s.kos.QueryBotState(req.Name)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error getting bot state for bot '%s': %s", req.Name, e))
		return
	}
	if botState != kelpos.BotStateStopped {
		s.writeErrorJson(w, fmt.Sprintf("bot state needs to be '%s' when updating bot config, but was '%s'\n", kelpos.BotStateStopped, botState))
		return
	}

	filenamePair := model.GetBotFilenames(req.Name, req.Strategy)
	traderFilePath := fmt.Sprintf("%s/%s", s.configsDir, filenamePair.Trader)
	botConfig := req.TraderConfig
	log.Printf("updating bot config to file: %s\n", traderFilePath)
	e = toml.WriteFile(traderFilePath, &botConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error writing trader botConfig toml file for bot '%s': %s", req.Name, e))
		return
	}

	strategyFilePath := fmt.Sprintf("%s/%s", s.configsDir, filenamePair.Strategy)
	strategyConfig := req.StrategyConfig
	log.Printf("updating strategy config to file: %s\n", strategyFilePath)
	e = toml.WriteFile(strategyFilePath, &strategyConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error writing strategy toml file for bot '%s': %s", req.Name, e))
		return
	}

	s.writeJson(w, updateBotConfigResponse{Success: true})
}
