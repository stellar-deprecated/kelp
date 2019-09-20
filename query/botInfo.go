package query

import (
	"fmt"
	"strings"
	"time"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// BotInfo is the response from the getBotInfo IPC request
type BotInfo struct {
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

func (s *Server) getBotInfo() (*BotInfo, error) {
	assetBase, assetQuote, e := s.sdex.Assets()
	if e != nil {
		return nil, fmt.Errorf("error getting assets from sdex: %s", e)
	}

	balanceBase, e := s.exchangeShim.GetBalanceHack(assetBase)
	if e != nil {
		return nil, fmt.Errorf("error getting base asset balance: %s", e)
	}

	balanceQuote, e := s.exchangeShim.GetBalanceHack(assetQuote)
	if e != nil {
		return nil, fmt.Errorf("error getting quote asset balance: %s", e)
	}

	offers, e := s.exchangeShim.LoadOffersHack()
	if e != nil {
		return nil, fmt.Errorf("error loading offers: %s", e)
	}
	sellingAOffers, buyingAOffers := utils.FilterOffers(offers, assetBase, assetQuote)
	numBids := len(buyingAOffers)
	numAsks := len(sellingAOffers)

	ob, e := s.exchangeShim.GetOrderBook(s.tradingPair, 20)
	if e != nil {
		return nil, fmt.Errorf("error loading orderbook (maxCount=20): %s", e)
	}
	topAsk := ob.TopAsk()
	topBid := ob.TopBid()
	spreadValue := model.NumberFromFloat(-1.0, 16)
	midPrice := model.NumberFromFloat(-1.0, 16)
	spreadPct := model.NumberFromFloat(-1.0, 16)
	if topBid != nil && topAsk != nil {
		spreadValue = topAsk.Price.Subtract(*topBid.Price)
		midPrice = topAsk.Price.Add(*topBid.Price).Scale(0.5)
		spreadPct = spreadValue.Divide(*midPrice)
	}

	return &BotInfo{
		LastUpdated:    time.Now().UTC().Format("1/_2/2006 15:04:05 MST"),
		TradingAccount: s.sdex.TradingAccount,
		Strategy:       s.strategyName,
		IsTestnet:      strings.Contains(s.sdex.API.HorizonURL, "test"),
		TradingPair:    s.tradingPair,
		AssetBase:      assetBase,
		AssetQuote:     assetQuote,
		BalanceBase:    balanceBase.Balance,
		BalanceQuote:   balanceQuote.Balance,
		NumBids:        numBids,
		NumAsks:        numAsks,
		SpreadValue:    spreadValue.AsFloat(),
		SpreadPercent:  spreadPct.AsFloat(),
	}, nil
}
