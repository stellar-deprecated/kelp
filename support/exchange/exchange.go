package exchange

import (
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/number"
)

// Ticker encapsulates all the data for a given Trading Pair
type Ticker struct {
	AskPrice  *number.Number
	AskVolume *number.Number
	BidPrice  *number.Number
	BidVolume *number.Number
}

// Exchange is the interface we use as a generic API to all crypto exchanges
type Exchange interface {
	GetTickerPrice([]assets.TradingPair) (map[assets.TradingPair]Ticker, error)
}
