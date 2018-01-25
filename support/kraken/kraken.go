package kraken

import (
	"github.com/Beldur/kraken-go-api-client"
	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
)

// ensure that krakenExchange conforms to the Exchange interface
var _ exchange.Exchange = krakenExchange{}

// krakenExchange is the implementation for the Kraken Exchange
type krakenExchange struct {
	assetConverter *assets.AssetConverter
	api            *krakenapi.KrakenApi
	delimiter      string
	precision      int8
	isSimulated    bool // will simulate add and cancel orders if this is true
}
