package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/trader"
)

type getBotConfigRequest struct {
	UserData UserData `json:"user_data"`
	BotName  string   `json:"bot_name"`
}

type botConfigResponse struct {
	Name           string                `json:"name"`
	Strategy       string                `json:"strategy"`
	TraderConfig   trader.BotConfig      `json:"trader_config"`
	StrategyConfig plugins.BuySellConfig `json:"strategy_config"`
}

func (s *APIServer) getBotConfig(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req getBotConfigRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}
	if strings.TrimSpace(req.UserData.ID) == "" {
		s.writeErrorJson(w, fmt.Sprintf("cannot have empty userID"))
		return
	}
	botName := req.BotName

	filenamePair := model2.GetBotFilenames(botName, "buysell")
	traderFilePath := s.botConfigsPathForUser(req.UserData.ID).Join(filenamePair.Trader)
	var botConfig trader.BotConfig
	e = config.Read(traderFilePath.Native(), &botConfig)
	if e != nil {
		s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot read bot config at path '%s': %s\n", traderFilePath, e),
		))
		return
	}
	strategyFilePath := s.botConfigsPathForUser(req.UserData.ID).Join(filenamePair.Strategy)
	var buysellConfig plugins.BuySellConfig
	e = config.Read(strategyFilePath.Native(), &buysellConfig)
	if e != nil {
		s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
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
		s.writeKelpError(req.UserData, w, makeKelpErrorResponseWrapper(
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
