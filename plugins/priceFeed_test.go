package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stellar/kelp/tests"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const wantLowerBoundXLM = 0.02
const wantUpperBoundXLM = 1.1

// uses real network calls
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
		},
		// disable ccxt-kraken based tests for now because of the 403 Forbidden Security check API error
		// {
		// 	typ:                    "exchange",
		// 	url:                    "ccxt-kraken/XLM/USD/last",
		// 	wantLowerOrEqualBound:  wantLowerBoundXLM,
		// 	wantHigherOrEqualBound: wantUpperBoundXLM,
		// },
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

// uses mock call
func TestMakePriceFeed_FiatFeed_Success(t *testing.T) {
	response := fiatAPIReturn{
		Success: true,
		Quotes:  map[string]float64{"SOME_CODE": 1},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer ts.Close()

	priceFeed, err := MakePriceFeed("fiat", ts.URL)
	assert.NoError(t, err)

	price, err := priceFeed.GetPrice()
	assert.Equal(t, response.Quotes["SOME_CODE"], price)
	assert.NoError(t, err)
}

// uses mock call
func TestMakePriceFeed_FiatFeedOxr_Success(t *testing.T) {
	symbol := tests.RandomString()
	response := createOxrResponse(symbol)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer ts.Close()

	priceFeed, err := MakePriceFeed("fiat-oxr", ts.URL)
	assert.NoError(t, err)

	price, err := priceFeed.GetPrice()
	assert.Equal(t, response.Rates[symbol], price)
	assert.NoError(t, err)
}

// uses mock call
func TestMakePriceFeed_CryptoFeed_Success(t *testing.T) {
	response := []cmcAPIReturn{{Price: strconv.Itoa(tests.RandomInt())}}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer ts.Close()

	priceFeed, err := MakePriceFeed("crypto", ts.URL)
	assert.NoError(t, err)

	expected, err := strconv.ParseFloat(response[0].Price, 64)
	require.NoError(t, err)

	price, err := priceFeed.GetPrice()
	assert.Equal(t, expected, price)
	assert.NoError(t, err)
}
