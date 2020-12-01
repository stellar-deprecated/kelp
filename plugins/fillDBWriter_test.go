package plugins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarketID(t *testing.T) {
	testCases := []struct {
		exchangeName string
		baseAsset    string
		quoteAsset   string
		wantMarketID string
	}{
		{
			exchangeName: "kraken",
			baseAsset:    "XLM",
			quoteAsset:   "USD",
			wantMarketID: "96eda0a6ec",
		}, {
			exchangeName: "kraken",
			baseAsset:    "XLM",
			quoteAsset:   "BTC",
			wantMarketID: "02bee86ba8",
		}, {
			exchangeName: "ccxt-kraken",
			baseAsset:    "XLM",
			quoteAsset:   "USD",
			wantMarketID: "8eb89f0940",
		}, {
			exchangeName: "ccxt-binance",
			baseAsset:    "XLM",
			quoteAsset:   "USDT",
			wantMarketID: "fadc072837",
		}, {
			exchangeName: "ccxt-coinbasepro",
			baseAsset:    "XLM",
			quoteAsset:   "USD",
			wantMarketID: "43b2018ed3",
		}, {
			exchangeName: "ccxt-poloniex",
			baseAsset:    "XLM",
			quoteAsset:   "USD",
			wantMarketID: "f908f42d25",
		}, {
			exchangeName: "ccxt-bitstamp",
			baseAsset:    "XLM",
			quoteAsset:   "USD",
			wantMarketID: "d438b87fff",
		},
	}

	for _, k := range testCases {
		t.Run(fmt.Sprintf("%s_%s_%s", k.exchangeName, k.baseAsset, k.quoteAsset), func(t *testing.T) {
			marketID := MakeMarketID(k.exchangeName, k.baseAsset, k.quoteAsset)

			assert.Equal(t, k.wantMarketID, marketID)
		})
	}
}
