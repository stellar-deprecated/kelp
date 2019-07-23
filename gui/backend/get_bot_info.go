package backend

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
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
	account, e := s.apiTestNet.AccountDetail(horizonclient.AccountRequest{AccountID: botConfig.TradingAccount()})
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("cannot get account data for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e))
		return
	}
	var balanceBase float64
	if assetBase == utils.NativeAsset {
		balanceBase, e = getNativeBalance(account)
		if e != nil {
			s.writeErrorJson(w, fmt.Sprintf("error getting native balanceBase for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e))
			return
		}
	} else {
		balanceBase, e = getCreditBalance(account, assetBase)
		if e != nil {
			s.writeErrorJson(w, fmt.Sprintf("error getting credit balanceBase for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e))
			return
		}
	}
	var balanceQuote float64
	if assetQuote == utils.NativeAsset {
		balanceQuote, e = getNativeBalance(account)
		if e != nil {
			s.writeErrorJson(w, fmt.Sprintf("error getting native balanceQuote for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e))
			return
		}
	} else {
		balanceQuote, e = getCreditBalance(account, assetQuote)
		if e != nil {
			s.writeErrorJson(w, fmt.Sprintf("error getting credit balanceQuote for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e))
			return
		}
	}

	offers, e := utils.LoadAllOffers(account.AccountID, horizon.DefaultTestNetClient)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error getting offers for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e))
		return
	}
	sellingAOffers, buyingAOffers := utils.FilterOffers(offers, assetBase, assetQuote)
	numBids := len(buyingAOffers)
	numAsks := len(sellingAOffers)

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

func getNativeBalance(account hProtocol.Account) (float64, error) {
	balanceString, e := account.GetNativeBalance()
	if e != nil {
		return 0.0, fmt.Errorf("cannot get native balance: %s\n", e)
	}

	balance, e := strconv.ParseFloat(balanceString, 64)
	if e != nil {
		return 0.0, fmt.Errorf("cannot parse native balance: %s (string value = %s)\n", e, balanceString)
	}

	return balance, nil
}

func getCreditBalance(account hProtocol.Account, asset horizon.Asset) (float64, error) {
	balanceString := account.GetCreditBalance(asset.Code, asset.Issuer)
	balance, e := strconv.ParseFloat(balanceString, 64)
	if e != nil {
		return 0.0, fmt.Errorf("cannot parse credit asset balance (%s:%s): %s (string value = %s)\n", asset.Code, asset.Issuer, e, balanceString)
	}

	return balance, nil
}
