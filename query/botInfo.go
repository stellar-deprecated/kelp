package query

import (
	"fmt"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

type botInfo struct {
	Strategy     string             `json:"strategy"`
	TradingPair  *model.TradingPair `json:"trading_pair"`
	AssetBase    horizon.Asset      `json:"asset_base"`
	AssetQuote   horizon.Asset      `json:"asset_quote"`
	BalanceBase  float64            `json:"balance_base"`
	BalanceQuote float64            `json:"balance_quote"`
	NumBids      int                `json:"num_bids"`
	NumAsks      int                `json:"num_asks"`
}

func (s *Server) getBotInfo() (*botInfo, error) {
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

	return &botInfo{
		Strategy:     s.strategyName,
		TradingPair:  s.tradingPair,
		AssetBase:    assetBase,
		AssetQuote:   assetQuote,
		BalanceBase:  balanceBase.Balance,
		BalanceQuote: balanceQuote.Balance,
		NumBids:      numBids,
		NumAsks:      numAsks,
	}, nil
}
