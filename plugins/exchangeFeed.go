package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// encapsulates a priceFeed from a tickerAPI
type exchangeFeed struct {
	name      string
	tickerAPI *api.TickerAPI
	pairs     []model.TradingPair
	modifier  string
}

// ensure that it implements PriceFeed
var _ api.PriceFeed = &exchangeFeed{}

func newExchangeFeed(name string, tickerAPI *api.TickerAPI, pair *model.TradingPair, modifier string) *exchangeFeed {
	return &exchangeFeed{
		name:      name,
		tickerAPI: tickerAPI,
		pairs:     []model.TradingPair{*pair},
		modifier:  modifier,
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

	midPrice := (p.BidPrice.AsFloat() + p.AskPrice.AsFloat()) / 2
	var price float64
	if f.modifier == "ask" {
		price = p.AskPrice.AsFloat()
	} else if f.modifier == "bid" {
		price = p.BidPrice.AsFloat()
	} else {
		price = midPrice
	}
	log.Printf("(modifier: %s) price from exchange feed (%s): bidPrice=%.7f, askPrice=%.7f, midPrice=%.7f; price=%.7f", f.modifier, f.name, p.BidPrice.AsFloat(), p.AskPrice.AsFloat(), midPrice, price)
	return price, nil
}
