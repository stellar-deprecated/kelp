package priceFeed

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
)

// encapsulates a priceFeed from an exchange
type exchangeFeed struct {
	exchange    *exchange.Exchange
	pairs       []assets.TradingPair
	useBidPrice bool // bid price if true, else ask price
}

// ensure that it implements priceFeed
var _ priceFeed = &exchangeFeed{}

func newExchangeFeed(exchange *exchange.Exchange, pair *assets.TradingPair, useBidPrice bool) *exchangeFeed {
	return &exchangeFeed{
		exchange:    exchange,
		pairs:       []assets.TradingPair{*pair},
		useBidPrice: useBidPrice,
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

	if f.useBidPrice {
		return p.BidPrice.AsFloat(), nil
	}
	return p.AskPrice.AsFloat(), nil
}
