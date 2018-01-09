package exchange

import (
	"github.com/lightyeario/kelp/support/exchange/assets"
)

// krakenExchange is the implementation for the Kraken Exchange
type krakenExchange struct {
	assetConverter *assets.AssetConverter
}

// GetPrice impl.
func (k krakenExchange) GetPrice(p assets.TradingPair) float64 {
	return 3.14159
}

// KrakenExchange is the singleton instance of the kraken implementation
var KrakenExchange Exchange = krakenExchange{
	assetConverter: assets.KrakenAssetConverter,
}
