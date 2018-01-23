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
}

func (k krakenExchange) parsePair(p string) (*assets.TradingPair, error) {
	a, e := k.assetConverter.FromString(p[0:4])
	if e != nil {
		return nil, e
	}

	b, e := k.assetConverter.FromString(p[4:8])
	if e != nil {
		return nil, e
	}

	return &assets.TradingPair{AssetA: a, AssetB: b}, nil
}
