package orderbook

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/api/assets"
	"github.com/lightyeario/kelp/support/exchange/api/dates"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// TransactionID is typed for the concept of a transaction ID
type TransactionID string

// String is the stringer function
func (t TransactionID) String() string {
	return string(t)
}

// MakeTransactionID is a factory method for convenience
func MakeTransactionID(s string) *TransactionID {
	t := TransactionID(s)
	return &t
}

// Order represents an order in the orderbook
type Order struct {
	Pair        *assets.TradingPair
	OrderAction OrderAction
	OrderType   OrderType
	Price       *number.Number
	Volume      *number.Number
	Timestamp   *dates.Timestamp
}

// String is the stringer function
func (o Order) String() string {
	return fmt.Sprintf("Order[pair=%s, action=%s, type=%s, price=%s, vol=%s, ts=%d]",
		o.Pair,
		o.OrderAction,
		o.OrderType,
		o.Price.AsString(),
		o.Volume.AsString(),
		o.Timestamp.AsInt64(),
	)
}
