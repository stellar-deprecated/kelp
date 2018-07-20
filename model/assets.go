package model

import (
	"errors"
	"os"

	"github.com/stellar/go/support/log"
)

// Asset is typed and enlists the allowed assets that are understood by the bot
type Asset string

// this is the list of assets understood by the bot.
// This string can be converted by the specific exchange adapter as is needed by the exchange's API
const (
	XLM Asset = "XLM"
	BTC Asset = "BTC"
	USD Asset = "USD"
	ETH Asset = "ETH"
	LTC Asset = "LTC"
	REP Asset = "REP"
)

// AssetConverter converts to and from the asset type, it is specific to an exchange
type AssetConverter struct {
	asset2String map[Asset]string
	string2Asset map[string]Asset
}

// makeAssetConverter is a factory method for AssetConverter
func makeAssetConverter(asset2String map[Asset]string) *AssetConverter {
	string2Asset := map[string]Asset{}
	for a, s := range asset2String {
		string2Asset[s] = a
	}

	return &AssetConverter{
		asset2String: asset2String,
		string2Asset: string2Asset,
	}
}

// ToString converts an asset to a string
func (c AssetConverter) ToString(a Asset) (string, error) {
	s, ok := c.asset2String[a]
	if !ok {
		return "", errors.New("could not recognize Asset: " + string(a))
	}
	return s, nil
}

// FromString converts from a string to an asset
func (c AssetConverter) FromString(s string) (Asset, error) {
	a, ok := c.string2Asset[s]
	if !ok {
		return "", errors.New("asset converter could not recognize string: " + s)
	}
	return a, nil
}

// MustFromString converts from a string to an asset, failing on errors
func (c AssetConverter) MustFromString(s string) Asset {
	a, e := c.FromString(s)
	if e != nil {
		log.Info(e)
		os.Exit(1)
	}
	return a
}

// Display is a basic converter for display purposes
var Display = makeAssetConverter(map[Asset]string{
	XLM: string(XLM),
	BTC: string(BTC),
	USD: string(USD),
	ETH: string(ETH),
	LTC: string(LTC),
	REP: string(REP),
})

// KrakenAssetConverter is the asset converter for the Kraken exchange
var KrakenAssetConverter = makeAssetConverter(map[Asset]string{
	XLM: "XXLM",
	BTC: "XXBT",
	USD: "ZUSD",
	ETH: "XETH",
	LTC: "XLTC",
	REP: "XREP",
})
