package sdk

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

func TestMakeInstanceName(t *testing.T) {
	testCases := []struct {
		testName     string
		exchangeName string
		apiKey       api.ExchangeAPIKey
		params       []api.ExchangeParam
		headers      []api.ExchangeHeader
		wantName     string
	}{
		// keys cases
		{
			testName:     "binance, no key or secret",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "", Secret: ""},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{},
			wantName:     "binance___",
		}, {
			testName:     "binance, no key but has secret",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "", Secret: "secret"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{},
			wantName:     "binance___",
		}, {
			testName:     "binance, has key and secret",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{},
			wantName:     "binance_1746258028__",
		}, {
			testName:     "binance, different key with same secret",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "key2", Secret: "secret"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{},
			wantName:     "binance_944401402__",
		}, {
			testName:     "binance, different key and different secret",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "key2", Secret: "secret2"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{},
			wantName:     "binance_944401402__",
		}, {
			testName:     "kraken, has key and secret",
			exchangeName: "kraken",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{},
			wantName:     "kraken_1746258028__",
		},
		// params cases - value can be any type
		{
			testName:     "binance, no key or secret, has params",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "", Secret: ""},
			params:       []api.ExchangeParam{{Param: "p", Value: "v"}, {Param: "p2", Value: "true"}},
			headers:      []api.ExchangeHeader{},
			wantName:     "binance__3356960995_",
		}, {
			testName:     "kraken, has key and secret, has params",
			exchangeName: "kraken",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{{Param: "p", Value: "v"}, {Param: "p2", Value: "true"}},
			headers:      []api.ExchangeHeader{},
			wantName:     "kraken_1746258028_3356960995_",
		}, {
			testName:     "kraken, has key and secret, has params with bool value",
			exchangeName: "kraken",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{{Param: "p", Value: "v"}, {Param: "p2", Value: true}},
			headers:      []api.ExchangeHeader{},
			wantName:     "kraken_1746258028_3623553427_",
		},
		// headers cases - headers is only string values
		{
			testName:     "binance, no key or secret, has headers",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "", Secret: ""},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{{Header: "h", Value: "v"}, {Header: "h", Value: "true"}},
			wantName:     "binance___2734440189",
		}, {
			testName:     "kraken, has key and secret, has headers set 1",
			exchangeName: "kraken",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{{Header: "h", Value: "v"}, {Header: "h", Value: "true"}},
			wantName:     "kraken_1746258028__2734440189",
		}, {
			testName:     "kraken, has key and secret, has headers set 2",
			exchangeName: "kraken",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{},
			headers:      []api.ExchangeHeader{{Header: "h", Value: "v"}, {Header: "h2", Value: "true"}},
			wantName:     "kraken_1746258028__2688111915",
		},
		// all parts included
		{
			testName:     "binance, no key or secret, has params, has headers",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "", Secret: ""},
			params:       []api.ExchangeParam{{Param: "p", Value: "v"}, {Param: "p2", Value: "true"}},
			headers:      []api.ExchangeHeader{{Header: "h", Value: "v"}, {Header: "h", Value: "true"}},
			wantName:     "binance__3356960995_2734440189",
		}, {
			testName:     "binance, has key and secret, has params, has headers",
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{Key: "key", Secret: "secret"},
			params:       []api.ExchangeParam{{Param: "p", Value: "v"}, {Param: "p2", Value: "true"}},
			headers:      []api.ExchangeHeader{{Header: "h", Value: "v"}, {Header: "h", Value: "true"}},
			wantName:     "binance_1746258028_3356960995_2734440189",
		},
	}

	for _, k := range testCases {
		t.Run(k.testName, func(t *testing.T) {
			actualName, e := makeInstanceName(k.exchangeName, k.apiKey, k.params, k.headers)
			if !assert.Nil(t, e) {
				return
			}

			assert.Equal(t, k.wantName, actualName)
		})
	}
}

func TestMakeValid(t *testing.T) {
	if testing.Short() {
		return
	}

	_, e := MakeInitializedCcxtExchange("kraken", api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
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

	_, e := MakeInitializedCcxtExchange("missing-exchange", api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
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

	c, e := MakeInitializedCcxtExchange("binance", api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
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
	assert.True(t, m["ask"].(float64) > 0)
	assert.True(t, m["bid"].(float64) > 0)
	assert.True(t, m["bid"].(float64) < m["ask"].(float64), fmt.Sprintf("bid price (%f) should be less than ask price (%f)", m["bid"].(float64), m["ask"].(float64)))
	assert.True(t, m["last"].(float64) > 0)
}

func TestFetchTickersWithMissingSymbol(t *testing.T) {
	if testing.Short() {
		return
	}

	c, e := MakeInitializedCcxtExchange("binance", api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
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
			exchangeName: "kraken",
			tradingPair:  "BTC/USD",
			limit:        nil,
			expectError:  false,
		}, {
			exchangeName: "kraken",
			tradingPair:  "BTC/USD",
			limit:        &limit5,
			expectError:  false,
		}, {
			exchangeName: "kraken",
			tradingPair:  "BTC/USD",
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
			exchangeName: "kraken",
			tradingPair:  "XLM/USD",
			limit:        &limit2,
			expectError:  false,
		},
	} {
		limitString := "nil"
		if k.limit != nil {
			limitString = fmt.Sprintf("%d", *k.limit)
		}

		tradingPairString := strings.Replace(k.tradingPair, "/", "_", -1)
		name := fmt.Sprintf("%s-%s-%v", k.exchangeName, tradingPairString, limitString)
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

	c, e := MakeInitializedCcxtExchange(k.exchangeName, api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
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

	// "id" is not always part of a trade result on Kraken for the public trades API
	krakenFields := []string{"amount", "cost", "datetime", "price", "side", "symbol", "timestamp"}
	poloniexFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp", "type"}
	binanceFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp"}
	bittrexFields := []string{"amount", "datetime", "id", "price", "side", "symbol", "timestamp", "type"}

	for _, k := range []struct {
		exchangeName   string
		tradingPair    string
		expectedFields []string
	}{
		{
			exchangeName:   "kraken",
			tradingPair:    "BTC/USD",
			expectedFields: krakenFields,
		}, {
			exchangeName:   "kraken",
			tradingPair:    "XLM/USD",
			expectedFields: krakenFields,
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
		}, {
			exchangeName:   "poloniex",
			tradingPair:    "XLM/BTC",
			expectedFields: poloniexFields,
		},
	} {
		tradingPairString := strings.Replace(k.tradingPair, "/", "_", -1)
		t.Run(fmt.Sprintf("%s-%s", k.exchangeName, tradingPairString), func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange(k.exchangeName, api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			trades, e := c.FetchTrades(k.tradingPair)
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when fetching trades: %s", e))
				return
			}

			validateTrades(trades, k.expectedFields, k.tradingPair, t)
		})
	}
}

func TestFetchMyTrades(t *testing.T) {
	if testing.Short() {
		return
	}

	krakenFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp", "fee"}
	binanceFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp", "fee"}
	bittrexFields := []string{"amount", "datetime", "id", "price", "side", "symbol", "timestamp", "type", "fee"}

	for _, k := range []struct {
		exchangeName     string
		tradingPair      string
		maybeCursorStart interface{}
		expectedFields   []string
		apiKey           api.ExchangeAPIKey
	}{
		{
			exchangeName:     "kraken",
			tradingPair:      "BTC/USD",
			maybeCursorStart: nil,
			expectedFields:   krakenFields,
			apiKey:           api.ExchangeAPIKey{},
		}, {
			exchangeName:     "binance",
			tradingPair:      "XLM/USDT",
			maybeCursorStart: nil,
			expectedFields:   binanceFields,
			apiKey:           api.ExchangeAPIKey{},
		}, {
			exchangeName:     "bittrex",
			tradingPair:      "XLM/BTC",
			maybeCursorStart: nil,
			expectedFields:   bittrexFields,
			apiKey:           api.ExchangeAPIKey{},
		},
	} {
		tradingPairString := strings.Replace(k.tradingPair, "/", "_", -1)
		t.Run(fmt.Sprintf("%s-%s", k.exchangeName, tradingPairString), func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange(k.exchangeName, k.apiKey, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			trades, e := c.FetchMyTrades(k.tradingPair, 50, k.maybeCursorStart)
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when fetching my trades: %s", e))
				return
			}

			validateTrades(trades, k.expectedFields, k.tradingPair, t)
		})
	}
}

func validateTrades(trades []CcxtTrade, expectedFields []string, tradingPair string, t *testing.T) {
	// convert expectedFields to a map and create the supportsField function
	fieldsMap := map[string]bool{}
	for _, f := range expectedFields {
		fieldsMap[f] = true
	}
	supportsField := func(field string) bool {
		_, ok := fieldsMap[field]
		return ok
	}

	if !assert.True(t, len(trades) > 0) {
		return
	}
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
		if supportsField("id") && !assert.True(t, len(trade.ID) > 0, fmt.Sprintf("ID='%s'", trade.ID)) {
			return
		}
		if supportsField("price") && !assert.True(t, trade.Price > 0) {
			return
		}
		if supportsField("side") && !assert.True(t, trade.Side == "sell" || trade.Side == "buy") {
			return
		}
		if supportsField("symbol") && !assert.True(t, trade.Symbol == tradingPair) {
			return
		}
		if supportsField("timestamp") && !assert.True(t, trade.Timestamp > 0) {
			return
		}
		if supportsField("fee") && !assert.NotNil(t, trade.Fee) && !assert.True(t, trade.Fee.Cost > 0) {
			return
		}
	}
}

func TestFetchBalance(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, k := range []struct {
		exchangeName string
		apiKey       api.ExchangeAPIKey
	}{
		{
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{},
		},
	} {
		t.Run(k.exchangeName, func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange(k.exchangeName, k.apiKey, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			balances, e := c.FetchBalance()
			if !assert.Nil(t, e) {
				return
			}

			if !assert.True(t, len(balances) > 0, fmt.Sprintf("%d", len(balances))) {
				return
			}

			for asset, ccxtBalance := range balances {
				if !assert.True(t, ccxtBalance.Total > 0, fmt.Sprintf("total balance for asset '%s' should have been > 0, was %f", asset, ccxtBalance.Total)) {
					return
				}

				log.Printf("balance for asset '%s': %+v\n", asset, ccxtBalance)
			}

			assert.Fail(t, "force fail")
		})
	}
}

func TestOpenOrders(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, k := range []struct {
		exchangeName string
		apiKey       api.ExchangeAPIKey
		tradingPair  model.TradingPair
	}{
		{
			exchangeName: "binance",
			apiKey:       api.ExchangeAPIKey{},
			tradingPair: model.TradingPair{
				Base:  model.XLM,
				Quote: model.BTC,
			},
		},
	} {
		t.Run(k.exchangeName, func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange(k.exchangeName, k.apiKey, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			openOrders, e := c.FetchOpenOrders([]string{k.tradingPair.String()})
			if !assert.NoError(t, e) {
				return
			}

			if !assert.True(t, len(openOrders) > 0, fmt.Sprintf("%d", len(openOrders))) {
				return
			}

			for asset, orderList := range openOrders {
				if !assert.Equal(t, k.tradingPair.String(), asset) {
					return
				}

				for _, o := range orderList {
					if !assert.Equal(t, k.tradingPair.String(), o.Symbol) {
						return
					}

					if !assert.True(t, o.Amount > 0, o.Amount) {
						return
					}

					if !assert.True(t, o.Price > 0, o.Price) {
						return
					}

					if !assert.Equal(t, "limit", o.Type) {
						return
					}

					log.Printf("order: %+v\n", o)
				}
			}

			assert.Fail(t, "force fail")
		})
	}
}

func TestCreateLimitOrder(t *testing.T) {
	if testing.Short() {
		return
	}

	apiKey := api.ExchangeAPIKey{}
	for _, k := range []struct {
		exchangeName string
		apiKey       api.ExchangeAPIKey
		tradingPair  model.TradingPair
		side         string
		amount       float64
		price        float64
	}{
		{
			exchangeName: "binance",
			apiKey:       apiKey,
			tradingPair: model.TradingPair{
				Base:  model.XLM,
				Quote: model.BTC,
			},
			side:   "sell",
			amount: 40,
			price:  0.00004228,
		}, {
			exchangeName: "binance",
			apiKey:       apiKey,
			tradingPair: model.TradingPair{
				Base:  model.XLM,
				Quote: model.BTC,
			},
			side:   "buy",
			amount: 42,
			price:  0.00002536,
		},
	} {
		t.Run(k.exchangeName, func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange(k.exchangeName, k.apiKey, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			openOrder, e := c.CreateLimitOrder(k.tradingPair.String(), k.side, k.amount, k.price, nil)
			if !assert.NoError(t, e) {
				return
			}

			if !assert.NotNil(t, openOrder) {
				return
			}

			if !assert.Equal(t, k.tradingPair.String(), openOrder.Symbol) {
				return
			}

			if !assert.NotEqual(t, "", openOrder.ID) {
				return
			}

			if !assert.Equal(t, k.amount, openOrder.Amount) {
				return
			}

			if !assert.Equal(t, k.price, openOrder.Price) {
				return
			}

			if !assert.Equal(t, "limit", openOrder.Type) {
				return
			}

			if !assert.Equal(t, k.side, openOrder.Side) {
				return
			}

			if !assert.Equal(t, 0.0, openOrder.Cost) {
				return
			}

			if !assert.Equal(t, 0.0, openOrder.Filled) {
				return
			}

			if !assert.Equal(t, "open", openOrder.Status) {
				return
			}
		})
	}
}

func TestCancelOrder(t *testing.T) {
	if testing.Short() {
		return
	}

	apiKey := api.ExchangeAPIKey{}
	for _, k := range []struct {
		exchangeName string
		apiKey       api.ExchangeAPIKey
		orderID      string
		tradingPair  model.TradingPair
	}{
		{
			exchangeName: "binance",
			apiKey:       apiKey,
			orderID:      "67391789",
			tradingPair: model.TradingPair{
				Base:  model.XLM,
				Quote: model.BTC,
			},
		}, {
			exchangeName: "binance",
			apiKey:       apiKey,
			orderID:      "67391791",
			tradingPair: model.TradingPair{
				Base:  model.XLM,
				Quote: model.BTC,
			},
		},
	} {
		t.Run(k.exchangeName, func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange(k.exchangeName, k.apiKey, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			openOrder, e := c.CancelOrder(k.orderID, k.tradingPair.String())
			if !assert.NoError(t, e) {
				return
			}

			isValid := validateCanceledOrder(t, openOrder, k.tradingPair.String())
			if !isValid {
				return
			}

			log.Printf("canceled order %+v\n", openOrder)

			assert.Fail(t, "force fail")
		})
	}
}

func validateCanceledOrder(t *testing.T, openOrder *CcxtOpenOrder, tradingPair string) bool {
	if !assert.NotNil(t, openOrder) {
		return false
	}

	if !assert.Equal(t, tradingPair, openOrder.Symbol) {
		return false
	}

	if !assert.NotEqual(t, "", openOrder.ID) {
		return false
	}

	if !assert.True(t, openOrder.Amount > 0, fmt.Sprintf("%f", openOrder.Amount)) {
		return false
	}

	if !assert.True(t, openOrder.Price > 0, fmt.Sprintf("%f", openOrder.Price)) {
		return false
	}

	if !assert.Equal(t, "limit", openOrder.Type) {
		return false
	}

	if !assert.True(t, openOrder.Side == "buy" || openOrder.Side == "sell", openOrder.Side) {
		return false
	}

	if !assert.Equal(t, 0.0, openOrder.Cost) {
		return false
	}

	if !assert.Equal(t, 0.0, openOrder.Filled) {
		return false
	}

	if !assert.Equal(t, "canceled", openOrder.Status) {
		return false
	}

	return true
}
