package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/trader"
)

type botConfigResponse struct {
	Name           string                `json:"name"`
	Strategy       string                `json:"strategy"`
	TraderConfig   trader.BotConfig      `json:"trader_config"`
	StrategyConfig plugins.BuySellConfig `json:"strategy_config"`
}

func (s *APIServer) getBotConfig(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error parsing bot name in getBotConfig: %s\n", e))
		return
	}

	filenamePair := model2.GetBotFilenames(botName, "buysell")
	traderFilePath := s.botConfigsPath.Join(filenamePair.Trader)
	var botConfig trader.BotConfig
	e = config.Read(traderFilePath.Native(), &botConfig)
	if e != nil {
		s.writeKelpError(w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot read bot config at path '%s': %s\n", traderFilePath, e),
		))
		return
	}
	strategyFilePath := s.botConfigsPath.Join(filenamePair.Strategy)
	var buysellConfig plugins.BuySellConfig
	e = config.Read(strategyFilePath.Native(), &buysellConfig)
	if e != nil {
		s.writeKelpError(w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot read strategy config at path '%s': %s\n", strategyFilePath, e),
		))
		return
	}

	response := botConfigResponse{
		Name:           botName,
		Strategy:       "buysell",
		TraderConfig:   botConfig,
		StrategyConfig: buysellConfig,
	}
	jsonBytes, e := json.MarshalIndent(response, "", "  ")
	if e != nil {
		s.writeKelpError(w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot marshal botConfigResponse: %s\n", e),
		))
		return
	}
	log.Printf("getBotConfig response for botName '%s': %s\n", botName, string(jsonBytes))
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
