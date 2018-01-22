package trades

import (
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
)

// Trade represents a trade on an exchange
type Trade struct {
	Type      *TradeType
	Price     *number.Number
	Volume    *number.Number
	Timestamp *dates.Timestamp
}
