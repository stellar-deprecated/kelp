package sdk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeValid(t *testing.T) {
	if testing.Short() {
		return
	}

	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "kraken")
	if e != nil {
		assert.Fail(t, fmt.Sprintf("unexpected error: %s", e))
		return
	}
	// success
}

func TestMakeInvalid(t *testing.T) {
	if testing.Short() {
		return
	}

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
	if testing.Short() {
		return
	}

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
	if testing.Short() {
		return
	}

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
	if testing.Short() {
		return
	}

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
	if testing.Short() {
		return
	}

	c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName)
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
		return
	}

	m, e := c.FetchOrderBook(k.tradingPair, k.limit)
	if e != nil && !k.expectError {
		assert.Fail(t, fmt.Sprintf("error when fetching orderbook: %s", e))
		return
	} else if e != nil {
		// success
		return
	} else if e == nil && k.expectError {
		assert.Fail(t, fmt.Sprintf("expected error when fetching orderbook but nothing thrown"))
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

func TestFetchTrades(t *testing.T) {
	if testing.Short() {
		return
	}

	poloniexFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp", "type"}
	binanceFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp"}
	bittrexFields := []string{"amount", "datetime", "id", "price", "side", "symbol", "timestamp", "type"}

	for _, k := range []tradesTest{
		{
			exchangeName:   "poloniex",
			tradingPair:    "BTC/USDT",
			expectedFields: poloniexFields,
		}, {
			exchangeName:   "poloniex",
			tradingPair:    "XLM/USDT",
			expectedFields: poloniexFields,
		}, {
			exchangeName:   "binance",
			tradingPair:    "BTC/USDT",
			expectedFields: binanceFields,
		}, {
			exchangeName:   "binance",
			tradingPair:    "XLM/USDT",
			expectedFields: binanceFields,
		}, {
			exchangeName:   "bittrex",
			tradingPair:    "XLM/BTC",
			expectedFields: bittrexFields,
		},
	} {
		t.Run(k.exchangeName, func(t *testing.T) {
			runTestFetchTrades(k, t)
		})
	}
}

type tradesTest struct {
	exchangeName   string
	tradingPair    string
	expectedFields []string
}

func runTestFetchTrades(k tradesTest, t *testing.T) {
	if testing.Short() {
		return
	}

	c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName)
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
		return
	}

	trades, e := c.FetchTrades(k.tradingPair)
	if e != nil {
		assert.Fail(t, fmt.Sprintf("error when fetching trades: %s", e))
		return
	}

	// convert expectedFields to a map and create the supportsField function
	fieldsMap := map[string]bool{}
	for _, f := range k.expectedFields {
		fieldsMap[f] = true
	}
	supportsField := func(field string) bool {
		_, ok := fieldsMap[field]
		return ok
	}

	assert.True(t, len(trades) > 0)
	for _, trade := range trades {
		if supportsField("amount") && !assert.True(t, trade.Amount > 0) {
			return
		}
		if supportsField("cost") && !assert.True(t, trade.Cost > 0) {
			return
		}
		if supportsField("datetime") && !assert.True(t, len(trade.Datetime) > 0) {
			return
		}
		if supportsField("id") && !assert.True(t, len(trade.ID) > 0) {
			return
		}
		if supportsField("price") && !assert.True(t, trade.Price > 0) {
			return
		}
		if supportsField("side") && !assert.True(t, trade.Side == "sell" || trade.Side == "buy") {
			return
		}
		if supportsField("symbol") && !assert.True(t, trade.Symbol == k.tradingPair) {
			return
		}
		if supportsField("timestamp") && !assert.True(t, trade.Timestamp > 0) {
			return
		}
	}
}
