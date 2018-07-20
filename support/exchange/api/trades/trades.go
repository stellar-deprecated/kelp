package trades

import (
	"fmt"

	"github.com/lightyeario/kelp/model"
)

// Trade represents a trade on an exchange
type Trade struct {
	model.Order
	TransactionID *model.TransactionID
	Cost          *model.Number
	Fee           *model.Number
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
