package trades

import (
	"fmt"
	"log"

	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
)

// Trade represents a trade on an exchange
type Trade struct {
	TransactionID *string
	Timestamp     *dates.Timestamp
	Type          *TradeType
	Pair          *assets.TradingPair
	Price         *number.Number
	Volume        *number.Number
	Cost          *number.Number
	Fee           *number.Number
}

func (t Trade) String() string {
	pair, e := t.Pair.ToString(assets.Display, "_")
	if e != nil {
		log.Panic(e)
	}

	return fmt.Sprintf("Trades[txid: %s, ts: %d, pair: %s, type: %s, price: %s, volume: %s, cost: %s, fee: %s]",
		*t.TransactionID,
		t.Timestamp,
		pair,
		t.Type,
		t.Price.AsString(),
		t.Volume.AsString(),
		t.Cost.AsString(),
		t.Fee.AsString(),
	)
}
