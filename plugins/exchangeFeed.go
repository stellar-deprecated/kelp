package plugins

import (
	"fmt"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
)

// encapsulates a priceFeed from a tickerAPI
type exchangeFeed struct {
	tickerAPI *api.TickerAPI
	pairs     []model.TradingPair
}

// ensure that it implements PriceFeed
var _ api.PriceFeed = &exchangeFeed{}

func newExchangeFeed(tickerAPI *api.TickerAPI, pair *model.TradingPair) *exchangeFeed {
	return &exchangeFeed{
		tickerAPI: tickerAPI,
		pairs:     []model.TradingPair{*pair},
	}
}

// GetPrice impl
func (f *exchangeFeed) GetPrice() (float64, error) {
	tickerAPI := *f.tickerAPI
	m, e := tickerAPI.GetTickerPrice(f.pairs)
	if e != nil {
		return 0, e
	}

	p, ok := m[f.pairs[0]]
	if !ok {
		return 0, fmt.Errorf("could not get price for trading pair: %s", f.pairs[0].String())
	}

	centerPrice := (p.BidPrice.AsFloat() + p.AskPrice.AsFloat()) / 2
	return centerPrice, nil
}
