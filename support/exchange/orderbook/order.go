package orderbook

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
)

// Order represents an order in the orderbook
type Order struct {
	OrderType OrderType
	Price     *number.Number
	Volume    *number.Number
	Timestamp *dates.Timestamp
}

// String is the stringer function
func (o Order) String() string {
	return fmt.Sprintf("Order[type=%s, price=%s, vol=%s, ts=%d]", o.OrderType, o.Price.AsString(), o.Volume.AsString(), o.Timestamp.AsInt64())
}
