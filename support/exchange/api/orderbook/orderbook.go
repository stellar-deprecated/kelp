package orderbook

import "github.com/lightyeario/kelp/model/assets"

// OrderBook encapsulates the concept of an orderbook on a market
type OrderBook struct {
	pair *model.TradingPair
	asks []Order
	bids []Order
}

// Asks returns the asks in an orderbook
func (o OrderBook) Asks() []Order {
	return o.asks
}

// Bids returns the bids in an orderbook
func (o OrderBook) Bids() []Order {
	return o.bids
}

// MakeOrderBook creates a new OrderBook from the asks and the bids
func MakeOrderBook(pair *model.TradingPair, asks []Order, bids []Order) *OrderBook {
	return &OrderBook{
		pair: pair,
		asks: asks,
		bids: bids,
	}
}
