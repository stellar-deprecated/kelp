package exchange

import (
	"testing"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/stretchr/testify/assert"
)

var testKrakenExchange Exchange = krakenExchange{
	assetConverter: assets.KrakenAssetConverter,
	api:            krakenapi.New("", ""),
}

func TestGetTickerPrice(t *testing.T) {
	xlmbtc := assets.TradingPair{AssetA: assets.XLM, AssetB: assets.BTC}
	pairs := []assets.TradingPair{xlmbtc}

	m, e := testKrakenExchange.GetTickerPrice(pairs)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, 1, len(m))

	ticker := m[xlmbtc]
	assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString())
}

func TestGetAccountBalances(t *testing.T) {
	a := assets.USD
	m, e := testKrakenExchange.GetAccountBalances([]assets.Asset{a})
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, 1, len(m))

	bal := m[a]
	assert.True(t, bal.AsFloat() > 0, bal.AsString())
}
