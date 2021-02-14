package plugins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const wantLowerBoundXLM = 0.02
const wantUpperBoundXLM = 1.1

func TestMakePriceFeed(t *testing.T) {
	testCases := []struct {
		typ                    string
		url                    string
		wantLowerOrEqualBound  float64
		wantHigherOrEqualBound float64
	}{
		{
			typ:                    "exchange",
			url:                    "kraken/XXLM/ZUSD/mid",
			wantLowerOrEqualBound:  wantLowerBoundXLM,
			wantHigherOrEqualBound: wantUpperBoundXLM,
		}, {
			typ:                    "exchange",
			url:                    "ccxt-binance/XLM/USDT/bid",
			wantLowerOrEqualBound:  wantLowerBoundXLM,
			wantHigherOrEqualBound: wantUpperBoundXLM,
		}, {
			typ:                    "exchange",
			url:                    "ccxt-coinbasepro/XLM/USD/ask",
			wantLowerOrEqualBound:  wantLowerBoundXLM,
			wantHigherOrEqualBound: wantUpperBoundXLM,
		}, {
			typ:                    "fixed",
			url:                    "1.23456",
			wantLowerOrEqualBound:  1.23456,
			wantHigherOrEqualBound: 1.23456,
		}, {
			typ:                    "sdex",
			url:                    "XLM:/USD:GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX",
			wantLowerOrEqualBound:  wantLowerBoundXLM,
			wantHigherOrEqualBound: wantUpperBoundXLM,
		}, {
			typ:                    "sdex",
			url:                    "USD:GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX/XLM:",
			wantLowerOrEqualBound:  1 / wantUpperBoundXLM,
			wantHigherOrEqualBound: 1 / wantLowerBoundXLM,
		}, {
			typ:                    "function",
			url:                    "max(fixed/1.0,fixed/1.4)",
			wantLowerOrEqualBound:  1.4,
			wantHigherOrEqualBound: 1.4,
		}, {
			typ:                    "function",
			url:                    "max(fixed/0.02,exchange/ccxt-binance/XLM/USDT/last)",
			wantLowerOrEqualBound:  wantLowerBoundXLM,
			wantHigherOrEqualBound: wantUpperBoundXLM,
		}, {
			typ:                    "function",
			url:                    "invert(fixed/0.02)",
			wantLowerOrEqualBound:  50.0,
			wantHigherOrEqualBound: 50.0,
			// }, { disable ccxt-kraken based tests for now because of the 403 Forbidden Security check API error
			// 	typ:                    "exchange",
			// 	url:                    "ccxt-kraken/XLM/USD/last",
			// 	wantLowerOrEqualBound:  wantLowerBoundXLM,
			// 	wantHigherOrEqualBound: wantUpperBoundXLM,
		},
		// not testing fiat here because it requires an access key
		// not testing crypto here because it's returning an error when passed an actual URL but works in practice
	}

	// cannot run this in parallel because ccxt fails (by not recognizing exchanges) when hit with too many requests at once
	for _, k := range testCases {
		t.Run(k.typ+"/"+k.url, func(t *testing.T) {
			pf, e := MakePriceFeed(k.typ, k.url)
			if !assert.NoError(t, e) {
				return
			}

			price, e := pf.GetPrice()
			if !assert.NoError(t, e) {
				return
			}

			assert.True(t, price >= k.wantLowerOrEqualBound, fmt.Sprintf("price was %.10f, should have been >= %.10f", price, k.wantLowerOrEqualBound))
			assert.True(t, price <= k.wantHigherOrEqualBound, fmt.Sprintf("price was %.10f, should have been <= %.10f", price, k.wantHigherOrEqualBound))
		})
	}
}
