package trades

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/orderbook"

	"github.com/lightyeario/kelp/support/exchange/number"
)

// Trade represents a trade on an exchange
type Trade struct {
	orderbook.Order
	TransactionID *orderbook.TransactionID
	Cost          *number.Number
	Fee           *number.Number
}

func (t Trade) String() string {
	return fmt.Sprintf("Trades[txid: %s, ts: %d, pair: %s, action: %s, type: %s, price: %s, volume: %s, cost: %s, fee: %s]",
		*t.TransactionID,
		t.Timestamp,
		*t.Pair,
		t.OrderAction,
		t.OrderType,
		t.Price.AsString(),
		t.Volume.AsString(),
		t.Cost.AsString(),
		t.Fee.AsString(),
	)
}
