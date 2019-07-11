package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stellar/go/strkey"
	"github.com/stellar/kelp/gui/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/toml"
	"github.com/stellar/kelp/trader"
)

type upsertBotConfigRequest struct {
	Name           string                `json:"name"`
	Strategy       string                `json:"strategy"`
	TraderConfig   trader.BotConfig      `json:"trader_config"`
	StrategyConfig plugins.BuySellConfig `json:"strategy_config"`
}

type upsertBotConfigResponse struct {
	Success bool `json:"success"`
}

type upsertBotConfigResponseErrors struct {
	Error  string                 `json:"error"`
	Fields upsertBotConfigRequest `json:"fields"`
}

func makeUpsertError(fields upsertBotConfigRequest) *upsertBotConfigResponseErrors {
	return &upsertBotConfigResponseErrors{
		Error:  "There are some errors marked in red inline",
		Fields: fields,
	}
}

func (s *APIServer) upsertBotConfig(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s", e))
		return
	}
	log.Printf("upsertBotConfig requestJson: %s\n", string(bodyBytes))

	var req upsertBotConfigRequest
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
		s.writeErrorJson(w, fmt.Sprintf("bot state needs to be '%s' when upserting bot config, but was '%s'\n", kelpos.BotStateStopped, botState))
		return
	}

	if errResp := s.validateConfigs(req); errResp != nil {
		s.writeJson(w, errResp)
		return
	}

	filenamePair := model.GetBotFilenames(req.Name, req.Strategy)
	traderFilePath := fmt.Sprintf("%s/%s", s.configsDir, filenamePair.Trader)
	botConfig := req.TraderConfig
	log.Printf("upsert bot config to file: %s\n", traderFilePath)
	e = toml.WriteFile(traderFilePath, &botConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error writing trader botConfig toml file for bot '%s': %s", req.Name, e))
		return
	}

	strategyFilePath := fmt.Sprintf("%s/%s", s.configsDir, filenamePair.Strategy)
	strategyConfig := req.StrategyConfig
	log.Printf("upsert strategy config to file: %s\n", strategyFilePath)
	e = toml.WriteFile(strategyFilePath, &strategyConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error writing strategy toml file for bot '%s': %s", req.Name, e))
		return
	}

	s.writeJson(w, upsertBotConfigResponse{Success: true})
}

func (s *APIServer) validateConfigs(req upsertBotConfigRequest) *upsertBotConfigResponseErrors {
	hasError := false
	errResp := upsertBotConfigRequest{
		TraderConfig:   trader.BotConfig{},
		StrategyConfig: plugins.BuySellConfig{},
	}

	if _, e := strkey.Decode(strkey.VersionByteSeed, req.TraderConfig.TradingSecretSeed); e != nil {
		errResp.TraderConfig.TradingSecretSeed = "invalid Trader Secret Key"
		hasError = true
	}

	if req.TraderConfig.AssetCodeA == "" || len(req.TraderConfig.AssetCodeA) > 12 {
		errResp.TraderConfig.AssetCodeA = "1 - 12 characters"
		hasError = true
	}

	if req.TraderConfig.AssetCodeB == "" || len(req.TraderConfig.AssetCodeB) > 12 {
		errResp.TraderConfig.AssetCodeB = "1 - 12 characters"
		hasError = true
	}

	if _, e := strkey.Decode(strkey.VersionByteSeed, req.TraderConfig.SourceSecretSeed); req.TraderConfig.SourceSecretSeed != "" && e != nil {
		errResp.TraderConfig.SourceSecretSeed = "invalid Source Secret Key"
		hasError = true
	}

	if len(req.StrategyConfig.Levels) == 0 || hasNewLevel(req.StrategyConfig.Levels) {
		errResp.StrategyConfig.Levels = []plugins.StaticLevel{}
		hasError = true
	}

	if hasError {
		return makeUpsertError(errResp)
	}
	return nil
}

func hasNewLevel(levels []plugins.StaticLevel) bool {
	for _, l := range levels {
		if l.AMOUNT == 0 || l.SPREAD == 0 {
			return true
		}
	}
	return false
}
