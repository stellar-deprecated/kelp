package orderbook

import (
	"fmt"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/exchange/api/dates"
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
	Pair        *model.TradingPair
	OrderAction OrderAction
	OrderType   OrderType
	Price       *model.Number
	Volume      *model.Number
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
