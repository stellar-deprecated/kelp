package sdk

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

func TestHashString(t *testing.T) {
	testCases := []struct {
		s        string
		wantHash uint32
	}{
		{
			s:        "hello",
			wantHash: 1335831723,
		}, {
			s:        "world",
			wantHash: 933488787,
		},
	}

	for _, kase := range testCases {
		t.Run(kase.s, func(t *testing.T) {
			result, e := hashString(kase.s)
			if !assert.Nil(t, e) {
				return
			}

			assert.Equal(t, kase.wantHash, result)
		})
	}
}

func TestMakeValid(t *testing.T) {
	if testing.Short() {
		return
	}

	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "kraken", api.ExchangeAPIKey{})
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

	_, e := MakeInitializedCcxtExchange("http://localhost:3000", "missing-exchange", api.ExchangeAPIKey{})
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

	c, e := MakeInitializedCcxtExchange("http://localhost:3000", "binance", api.ExchangeAPIKey{})
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

	c, e := MakeInitializedCcxtExchange("http://localhost:3000", "binance", api.ExchangeAPIKey{})
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

	c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, api.ExchangeAPIKey{})
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

	for _, k := range []struct {
		exchangeName   string
		tradingPair    string
		expectedFields []string
	}{
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
			c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, api.ExchangeAPIKey{})
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

	poloniexFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp", "type"}
	binanceFields := []string{"amount", "cost", "datetime", "id", "price", "side", "symbol", "timestamp"}
	bittrexFields := []string{"amount", "datetime", "id", "price", "side", "symbol", "timestamp", "type"}

	for _, k := range []struct {
		exchangeName   string
		tradingPair    string
		expectedFields []string
		apiKey         api.ExchangeAPIKey
	}{
		{
			exchangeName:   "poloniex",
			tradingPair:    "BTC/USDT",
			expectedFields: poloniexFields,
			apiKey:         api.ExchangeAPIKey{},
		}, {
			exchangeName:   "binance",
			tradingPair:    "XLM/USDT",
			expectedFields: binanceFields,
			apiKey:         api.ExchangeAPIKey{},
		}, {
			exchangeName:   "bittrex",
			tradingPair:    "XLM/BTC",
			expectedFields: bittrexFields,
			apiKey:         api.ExchangeAPIKey{},
		},
	} {
		t.Run(k.exchangeName, func(t *testing.T) {
			c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, k.apiKey)
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			trades, e := c.FetchMyTrades(k.tradingPair)
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
		if supportsField("symbol") && !assert.True(t, trade.Symbol == tradingPair) {
			return
		}
		if supportsField("timestamp") && !assert.True(t, trade.Timestamp > 0) {
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
			c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, k.apiKey)
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
			c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, k.apiKey)
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
			c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, k.apiKey)
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			openOrder, e := c.CreateLimitOrder(k.tradingPair.String(), k.side, k.amount, k.price)
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
			c, e := MakeInitializedCcxtExchange("http://localhost:3000", k.exchangeName, k.apiKey)
			if e != nil {
				assert.Fail(t, fmt.Sprintf("error when making ccxt exchange: %s", e))
				return
			}

			openOrder, e := c.CancelOrder(k.orderID, k.tradingPair.String())
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

			if !assert.True(t, openOrder.Amount > 0, fmt.Sprintf("%f", openOrder.Amount)) {
				return
			}

			if !assert.True(t, openOrder.Price > 0, fmt.Sprintf("%f", openOrder.Price)) {
				return
			}

			if !assert.Equal(t, "limit", openOrder.Type) {
				return
			}

			if !assert.True(t, openOrder.Side == "buy" || openOrder.Side == "sell", openOrder.Side) {
				return
			}

			if !assert.Equal(t, 0.0, openOrder.Cost) {
				return
			}

			if !assert.Equal(t, 0.0, openOrder.Filled) {
				return
			}

			if !assert.Equal(t, "canceled", openOrder.Status) {
				return
			}

			log.Printf("canceled order %+v\n", openOrder)

			assert.Fail(t, "force fail")
		})
	}
}
