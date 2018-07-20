package priceFeed

import (
	"fmt"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/exchange/api"
)

// encapsulates a priceFeed from an exchange
type exchangeFeed struct {
	exchange *api.Exchange
	pairs    []model.TradingPair
}

// ensure that it implements priceFeed
var _ priceFeed = &exchangeFeed{}

func newExchangeFeed(exchange *api.Exchange, pair *model.TradingPair) *exchangeFeed {
	return &exchangeFeed{
		exchange: exchange,
		pairs:    []model.TradingPair{*pair},
	}
}

func (f *exchangeFeed) getPrice() (float64, error) {
	ex := *f.exchange
	m, e := ex.GetTickerPrice(f.pairs)
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
