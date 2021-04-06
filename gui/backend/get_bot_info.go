package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/utils"
	"github.com/stellar/kelp/trader"
)

const buysell = "buysell"

// botInfo is the response from the getBotInfo request
type botInfo struct {
	LastUpdated    string             `json:"last_updated"`
	TradingAccount string             `json:"trading_account"`
	Strategy       string             `json:"strategy"`
	IsTestnet      bool               `json:"is_testnet"`
	TradingPair    *model.TradingPair `json:"trading_pair"`
	AssetBase      hProtocol.Asset    `json:"asset_base"`
	AssetQuote     hProtocol.Asset    `json:"asset_quote"`
	BalanceBase    float64            `json:"balance_base"`
	BalanceQuote   float64            `json:"balance_quote"`
	NumBids        int                `json:"num_bids"`
	NumAsks        int                `json:"num_asks"`
	SpreadValue    float64            `json:"spread_value"`
	SpreadPercent  float64            `json:"spread_pct"`
}

type getBotInfoRequest struct {
	UserData UserData `json:"user_data"`
	BotName  string   `json:"bot_name"`
}

func (s *APIServer) getBotInfo(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error when reading request input: %s\n", e))
		return
	}
	var req getBotInfoRequest
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

	s.runGetBotInfoDirect(w, req.UserData, botName)
}

func (s *APIServer) runGetBotInfoDirect(w http.ResponseWriter, userData UserData, botName string) {
	log.Printf("getBotInfo is invoking logic directly for botName: %s\n", botName)

	botState, e := s.doGetBotState(userData, botName)
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot read bot state for bot '%s': %s\n", botName, e),
		))
		return
	}
	if botState == kelpos.BotStateInitializing {
		log.Printf("bot state is initializing for bot '%s'\n", botName)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	filenamePair := model2.GetBotFilenames(botName, buysell)
	traderFilePath := s.botConfigsPathForUser(userData.ID).Join(filenamePair.Trader)
	var botConfig trader.BotConfig
	e = config.Read(traderFilePath.Native(), &botConfig)
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot read bot config at path '%s': %s\n", traderFilePath.AsString(), e),
		))
		return
	}
	e = botConfig.Init()
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot init bot config at path '%s': %s\n", traderFilePath.AsString(), e),
		))
		return
	}

	assetBase := botConfig.AssetBase()
	assetQuote := botConfig.AssetQuote()
	tradingPair := &model.TradingPair{
		Base:  model.Asset(utils.Asset2CodeString(assetBase)),
		Quote: model.Asset(utils.Asset2CodeString(assetQuote)),
	}

	client := s.apiPubNet
	if strings.Contains(botConfig.HorizonURL, "test") {
		client = s.apiTestNet
	}

	account, e := client.AccountDetail(horizonclient.AccountRequest{AccountID: botConfig.TradingAccount()})
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("cannot get account data for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e),
		))
		return
	}
	var balanceBase float64
	if assetBase == utils.NativeAsset {
		balanceBase, e = getNativeBalance(account)
		if e != nil {
			s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelError,
				fmt.Sprintf("error getting native balanceBase for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e),
			))
			return
		}
	} else {
		balanceBase, e = getCreditBalance(account, assetBase)
		if e != nil {
			s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelError,
				fmt.Sprintf("error getting credit balanceBase for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e),
			))
			return
		}
	}
	var balanceQuote float64
	if assetQuote == utils.NativeAsset {
		balanceQuote, e = getNativeBalance(account)
		if e != nil {
			s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelError,
				fmt.Sprintf("error getting native balanceQuote for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e),
			))
			return
		}
	} else {
		balanceQuote, e = getCreditBalance(account, assetQuote)
		if e != nil {
			s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
				errorTypeBot,
				botName,
				time.Now().UTC(),
				errorLevelError,
				fmt.Sprintf("error getting credit balanceQuote for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e),
			))
			return
		}
	}

	offers, e := utils.LoadAllOffers(account.AccountID, client)
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("error getting offers for account '%s' for botName '%s': %s\n", botConfig.TradingAccount(), botName, e),
		))
		return
	}
	sellingAOffers, buyingAOffers := utils.FilterOffers(offers, assetBase, assetQuote)
	numBids := len(buyingAOffers)
	numAsks := len(sellingAOffers)

	obs, e := client.OrderBook(horizonclient.OrderBookRequest{
		SellingAssetType:   horizonclient.AssetType(assetBase.Type),
		SellingAssetCode:   assetBase.Code,
		SellingAssetIssuer: assetBase.Issuer,
		BuyingAssetType:    horizonclient.AssetType(assetQuote.Type),
		BuyingAssetCode:    assetQuote.Code,
		BuyingAssetIssuer:  assetQuote.Issuer,
		Limit:              1,
	})
	if e != nil {
		s.writeKelpError(userData, w, makeKelpErrorResponseWrapper(
			errorTypeBot,
			botName,
			time.Now().UTC(),
			errorLevelError,
			fmt.Sprintf("error getting orderbook for assets (base=%v, quote=%v) for botName '%s': %s\n", assetBase, assetQuote, botName, e),
		))
		return
	}
	spread := -1.0
	spreadPct := -1.0
	if len(obs.Asks) > 0 && len(obs.Bids) > 0 {
		topAsk := float64(obs.Asks[0].PriceR.N) / float64(obs.Asks[0].PriceR.D)
		topBid := float64(obs.Bids[0].PriceR.N) / float64(obs.Bids[0].PriceR.D)

		spread = topAsk - topBid
		midPrice := (topAsk + topBid) / 2
		spreadPct = 100.0 * spread / midPrice
	}

	bi := botInfo{
		LastUpdated:    time.Now().UTC().Format("1/_2/2006 15:04:05 MST"),
		TradingAccount: account.AccountID,
		Strategy:       buysell,
		IsTestnet:      strings.Contains(botConfig.HorizonURL, "test"),
		TradingPair:    tradingPair,
		AssetBase:      assetBase,
		AssetQuote:     assetQuote,
		BalanceBase:    balanceBase,
		BalanceQuote:   balanceQuote,
		NumBids:        numBids,
		NumAsks:        numAsks,
		SpreadValue:    model.NumberFromFloat(spread, 8).AsFloat(),
		SpreadPercent:  model.NumberFromFloat(spreadPct, 8).AsFloat(),
	}

	marshalledJSON, e := json.MarshalIndent(bi, "", "  ")
	if e != nil {
		log.Printf("cannot marshall to json response (error=%s), botInfo: %+v\n", e, bi)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{}"))
		return
	}
	marshalledJSONString := string(marshalledJSON)
	log.Printf("getBotInfo returned direct response for botName '%s': %s\n", botName, marshalledJSONString)

	w.WriteHeader(http.StatusOK)
	w.Write(marshalledJSON)
}

func getNativeBalance(account hProtocol.Account) (float64, error) {
	balanceString, e := account.GetNativeBalance()
	if e != nil {
		return 0.0, fmt.Errorf("cannot get native balance: %s", e)
	}

	balance, e := strconv.ParseFloat(balanceString, 64)
	if e != nil {
		return 0.0, fmt.Errorf("cannot parse native balance: %s (string value = %s)", e, balanceString)
	}

	return balance, nil
}

func getCreditBalance(account hProtocol.Account, asset hProtocol.Asset) (float64, error) {
	balanceString := account.GetCreditBalance(asset.Code, asset.Issuer)
	balance, e := strconv.ParseFloat(balanceString, 64)
	if e != nil {
		return 0.0, fmt.Errorf("cannot parse credit asset balance (%s:%s): %s (string value = %s)", asset.Code, asset.Issuer, e, balanceString)
	}

	return balance, nil
}
