package orderbook

import (
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
)

// OrderBook encapsulates the concept of an orderbook on a market
// TODO 2 - add the trading pair to this struct
type OrderBook struct {
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

// MakeAsk creates a new Ask Order
func MakeAsk(price *number.Number, volume *number.Number, timestamp *dates.Timestamp) Order {
	return Order{
		OrderType: TypeAsk,
		Price:     price,
		Volume:    volume,
		Timestamp: timestamp,
	}
}

// MakeBid creates a new Bid Order
func MakeBid(price *number.Number, volume *number.Number, timestamp *dates.Timestamp) Order {
	return Order{
		OrderType: TypeBid,
		Price:     price,
		Volume:    volume,
		Timestamp: timestamp,
	}
}

// MakeOrderBook creates a new OrderBook from the asks and the bids
func MakeOrderBook(asks []Order, bids []Order) *OrderBook {
	return &OrderBook{
		asks: asks,
		bids: bids,
	}
}
