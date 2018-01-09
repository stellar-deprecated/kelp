package exchange

import (
	"github.com/lightyeario/kelp/support/exchange/assets"
)

// Exchange is the interface we use as a generic API to all crypto exchanges
type Exchange interface {
	GetPrice(assets.TradingPair) float64
}
