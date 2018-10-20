package plugins

import (
	"fmt"
	"log"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
)

// encapsulates a priceFeed from a tickerAPI
type exchangeFeed struct {
	name      string
	tickerAPI *api.TickerAPI
	pairs     []model.TradingPair
}

// ensure that it implements PriceFeed
var _ api.PriceFeed = &exchangeFeed{}

func newExchangeFeed(name string, tickerAPI *api.TickerAPI, pair *model.TradingPair) *exchangeFeed {
	return &exchangeFeed{
		name:      name,
		tickerAPI: tickerAPI,
		pairs:     []model.TradingPair{*pair},
	}
}

// GetPrice impl
func (f *exchangeFeed) GetPrice() (float64, error) {
	tickerAPI := *f.tickerAPI
	m, e := tickerAPI.GetTickerPrice(f.pairs)
	if e != nil {
		return 0, fmt.Errorf("error while getting price from exchange feed: %s", e)
	}

	p, ok := m[f.pairs[0]]
	if !ok {
		return 0, fmt.Errorf("could not get price for trading pair: %s", f.pairs[0].String())
	}

	centerPrice := (p.BidPrice.AsFloat() + p.AskPrice.AsFloat()) / 2
	log.Printf("price from exchange feed (%s): bidPrice=%.7f, askPrice=%.7f, centerPrice=%.7f", f.name, p.BidPrice.AsFloat(), p.AskPrice.AsFloat(), centerPrice)
	return centerPrice, nil
}
