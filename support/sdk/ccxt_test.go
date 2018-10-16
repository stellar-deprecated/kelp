package sdk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeValid(t *testing.T) {
	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "kraken")
	if e != nil {
		assert.Fail(t, fmt.Sprintf("unexpected error: %s", e))
		return
	}
	// success
}

func TestMakeInvalid(t *testing.T) {
	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "missing-exchange")
	if e == nil {
		assert.Fail(t, "expected an error when trying to make and initialize an exchange that is missing: 'missing-exchange'")
		return
	}

	if !strings.Contains(e.Error(), "exchange name 'missing-exchange' is not in the list") {
		assert.Fail(t, fmt.Sprintf("unexpected error: %s", e))
		return
	}
	// success
}

func TestFetchTickers(t *testing.T) {
	c, e := MakeInitializedCcxtExchange("http://localhost:3000", "binance")
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
		return
	}

	m, e := c.FetchTicker("BTC/USDT")
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when fetching tickers: %s", e))
		return
	}

	assert.Equal(t, "BTC/USDT", m["symbol"].(string))
	assert.True(t, m["last"].(float64) > 0)
}

func TestFetchTickersWithMissingSymbol(t *testing.T) {
	c, e := MakeInitializedCcxtExchange("http://localhost:3000", "binance")
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
		return
	}

	_, e = c.FetchTicker("BTC/ABCDEFGH")
	if e == nil {
		assert.Fail(t, fmt.Sprintf("expected an error related to symbol '%s' not found but did not receive any error", "BTC/ABCDEFGH"))
		return
	}

	if !strings.Contains(e.Error(), "trading pair 'BTC/ABCDEFGH' does not exist") {
		assert.Fail(t, fmt.Sprintf("unexpected error: %s", e))
		return
	}
	// success
}

func TestFetchOrderBook(t *testing.T) {
	limit5 := 5
	limit2 := 2
	for _, k := range []orderbookTest{
		{
			exchangeName: "poloniex",
			tradingPair:  "BTC/USDT",
			limit:        nil,
			expectError:  false,
		}, {
			exchangeName: "poloniex",
			tradingPair:  "BTC/USDT",
			limit:        &limit5,
			expectError:  false,
		}, {
			exchangeName: "poloniex",
			tradingPair:  "BTC/USDT",
			limit:        &limit2,
			expectError:  false,
		}, {
			exchangeName: "binance",
			tradingPair:  "BTC/USDT",
			limit:        nil,
			expectError:  false,
		}, {
			exchangeName: "binance",
			tradingPair:  "BTC/USDT",
			limit:        &limit5,
			expectError:  false,
		}, {
			exchangeName: "binance",
			tradingPair:  "BTC/USDT",
			limit:        &limit2,
			expectError:  true,
		}, {
			exchangeName: "poloniex",
			tradingPair:  "XLM/USDT",
			limit:        &limit2,
			expectError:  false,
		},
	} {
		name := fmt.Sprintf("%s-nil", k.exchangeName)
		if k.limit != nil {
			name = fmt.Sprintf("%s-%v", k.exchangeName, *k.limit)
		}

		t.Run(name, func(t *testing.T) {
			runTestFetchOrderBook(k, t)
		})
	}
}

type orderbookTest struct {
	exchangeName string
	tradingPair  string
	limit        *int
	expectError  bool
}

func runTestFetchOrderBook(k orderbookTest, t *testing.T) {
	c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName)
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
		return
	}

	m, e := c.FetchOrderBook(k.tradingPair, k.limit)
	if e != nil && !k.expectError {
		assert.Fail(t, fmt.Sprintf("error when fetching tickers: %s", e))
		return
	} else if e != nil {
		// success
		return
	} else if e == nil && k.expectError {
		assert.Fail(t, fmt.Sprintf("expected error when fetching tickers but nothing thrown"))
		return
	}
	// else run checks below

	validateOrders := func(key string) {
		orders, ok := m[key]
		assert.True(t, ok)
		if k.limit != nil {
			assert.Equal(t, len(orders), *k.limit)
		}
		for _, o := range orders {
			assert.True(t, o.Price > 0)
			assert.True(t, o.Amount > 0)
		}
	}
	validateOrders("asks")
	validateOrders("bids")
}
