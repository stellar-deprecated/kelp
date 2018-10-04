package plugins

import (
	"fmt"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/sdk"
	"github.com/lightyeario/kelp/support/utils"
)

// TODO should confirm to the api.Exchange interface
// ensure that ccxtExchange conforms to the TickerAPI interface for now
var _ api.TickerAPI = ccxtExchange{}

// ccxtExchange is the implementation for the CCXT REST library that supports many exchanges (https://github.com/franz-see/ccxt-rest, https://github.com/ccxt/ccxt/)
type ccxtExchange struct {
	assetConverter *model.AssetConverter
	delimiter      string
	api            *sdk.Ccxt
	precision      int8
}

// makeCcxtExchange is a factory method to make an exchange using the CCXT interface
// TODO should return api.Exchange
func makeCcxtExchange(ccxtBaseURL string, exchangeName string) (api.TickerAPI, error) {
	c, e := sdk.MakeInitializedCcxtExchange(ccxtBaseURL, exchangeName)
	if e != nil {
		return nil, fmt.Errorf("error making a ccxt exchange: %s", e)
	}

	return ccxtExchange{
		assetConverter: model.CcxtAssetConverter,
		delimiter:      "/",
		api:            c,
		precision:      utils.SdexPrecision,
	}, nil
}

// GetTickerPrice impl.
func (c ccxtExchange) GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]api.Ticker, error) {
	pairsMap, e := model.TradingPairs2Strings(c.assetConverter, c.delimiter, pairs)
	if e != nil {
		return nil, e
	}

	priceResult := map[model.TradingPair]api.Ticker{}
	for _, p := range pairs {
		tickerMap, e := c.api.FetchTicker(pairsMap[p])
		if e != nil {
			return nil, fmt.Errorf("error while fetching ticker price for trading pair %s: %s", pairsMap[p], e)
		}

		priceResult[p] = api.Ticker{
			AskPrice:  model.NumberFromFloat(tickerMap["ask"].(float64), c.precision),
			AskVolume: model.NumberFromFloat(tickerMap["askVolume"].(float64), c.precision),
			BidPrice:  model.NumberFromFloat(tickerMap["bid"].(float64), c.precision),
			BidVolume: model.NumberFromFloat(tickerMap["bidVolume"].(float64), c.precision),
		}
	}

	return priceResult, nil
}
