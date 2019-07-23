package backend

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/query"
	"github.com/stellar/kelp/support/utils"
	"github.com/stellar/kelp/trader"
)

const buysell = "buysell"

func (s *APIServer) getBotInfo(w http.ResponseWriter, r *http.Request) {
	botName, e := s.parseBotName(r)
	if e != nil {
		s.writeError(w, fmt.Sprintf("error parsing bot name in getBotInfo: %s\n", e))
		return
	}

	// s.runGetBotInfoViaIPC(w, botName)
	s.runGetBotInfoDirect(w, botName)
}

func (s *APIServer) runGetBotInfoViaIPC(w http.ResponseWriter, botName string) {
	p, exists := s.kos.GetProcess(botName)
	if !exists {
		log.Printf("kelp bot process with name '%s' does not exist; processes available: %v\n", botName, s.kos.RegisteredProcesses())
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{}"))
		return
	}

	log.Printf("getBotInfo is making IPC request for botName: %s\n", botName)
	p.PipeIn.Write([]byte("getBotInfo\n"))
	scanner := bufio.NewScanner(p.PipeOut)
	output := ""
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, utils.IPCBoundary) {
			break
		}
		output += text
	}
	var buf bytes.Buffer
	e := json.Indent(&buf, []byte(output), "", "  ")
	if e != nil {
		log.Printf("cannot indent json response (error=%s), json_response: %s\n", e, output)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{}"))
		return
	}
	log.Printf("getBotInfo returned IPC response for botName '%s': %s\n", botName, buf.String())

	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (s *APIServer) runGetBotInfoDirect(w http.ResponseWriter, botName string) {
	log.Printf("getBotInfo is invoking logic directly for botName: %s\n", botName)

	filenamePair := model2.GetBotFilenames(botName, buysell)
	traderFilePath := fmt.Sprintf("%s/%s", s.configsDir, filenamePair.Trader)
	var botConfig trader.BotConfig
	e := config.Read(traderFilePath, &botConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("cannot read bot config at path '%s': %s\n", traderFilePath, e))
		return
	}
	e = botConfig.Init()
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("cannot init bot config at path '%s': %s\n", traderFilePath, e))
		return
	}

	assetBase := botConfig.AssetBase()
	assetQuote := botConfig.AssetQuote()
	tradingPair := &model.TradingPair{
		Base:  model.Asset(utils.Asset2CodeString(assetBase)),
		Quote: model.Asset(utils.Asset2CodeString(assetQuote)),
	}
	balanceBase := 100.0
	balanceQuote := 1000.0
	numBids := 9
	numAsks := 90
	spread := 0.01
	spreadPct := 0.05
	bi := query.BotInfo{
		Strategy:      buysell,
		TradingPair:   tradingPair,
		AssetBase:     assetBase,
		AssetQuote:    assetQuote,
		BalanceBase:   balanceBase,
		BalanceQuote:  balanceQuote,
		NumBids:       numBids,
		NumAsks:       numAsks,
		SpreadValue:   spread,
		SpreadPercent: spreadPct,
	}

	marshalledJson, e := json.MarshalIndent(bi, "", "  ")
	if e != nil {
		log.Printf("cannot marshall to json response (error=%s), BotInfo: %+v\n", e, bi)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{}"))
		return
	}
	marshalledJsonString := string(marshalledJson)
	log.Printf("getBotInfo returned direct response for botName '%s': %s\n", botName, marshalledJsonString)

	w.WriteHeader(http.StatusOK)
	w.Write(marshalledJson)
}
