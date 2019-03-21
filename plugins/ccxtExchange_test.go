package plugins

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

var supportedExchanges = []string{"binance"}
var emptyAPIKey = api.ExchangeAPIKey{}
var supportedTradingExchanges = map[string]api.ExchangeAPIKey{
	"binance": {},
}

var testOrderConstraints = map[model.TradingPair]model.OrderConstraints{
	*model.MakeTradingPair(model.XLM, model.USDT): *model.MakeOrderConstraints(4, 5, 0.1),
	*model.MakeTradingPair(model.XLM, model.BTC):  *model.MakeOrderConstraints(8, 4, 1.0),
}

func TestGetTickerPrice_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{emptyAPIKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
			pairs := []model.TradingPair{pair}

			m, e := testCcxtExchange.GetTickerPrice(pairs)
			if !assert.NoError(t, e) {
				return
			}
			assert.Equal(t, 1, len(m))

			ticker := m[pair]
			assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString())
		})
	}
}

func TestGetOrderBook_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{emptyAPIKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
			ob, e := testCcxtExchange.GetOrderBook(&pair, 10)
			if !assert.NoError(t, e) {
				return
			}
			assert.Equal(t, ob.Pair(), &pair)

			assert.True(t, len(ob.Asks()) > 0, fmt.Sprintf("%d", len(ob.Asks())))
			assert.True(t, len(ob.Bids()) > 0, fmt.Sprintf("%d", len(ob.Bids())))
			assert.True(t, ob.Asks()[0].OrderAction.IsSell())
			assert.True(t, ob.Asks()[0].OrderType.IsLimit())
			assert.True(t, ob.Bids()[0].OrderAction.IsBuy())
			assert.True(t, ob.Bids()[0].OrderType.IsLimit())
			assert.True(t, ob.Asks()[0].Price.AsFloat() > 0)
			assert.True(t, ob.Asks()[0].Volume.AsFloat() > 0)
			assert.True(t, ob.Bids()[0].Price.AsFloat() > 0)
			assert.True(t, ob.Bids()[0].Volume.AsFloat() > 0)
		})
	}
}

func TestGetTrades_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{emptyAPIKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
			// TODO test with cursor once implemented
			tradeResult, e := testCcxtExchange.GetTrades(&pair, nil)
			if !assert.NoError(t, e) {
				return
			}
			assert.Equal(t, nil, tradeResult.Cursor)

			validateTrades(t, pair, tradeResult.Trades)
		})
	}
}

func TestGetTradeHistory_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, apiKey := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{apiKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
			// TODO test with cursor once implemented
			tradeHistoryResult, e := testCcxtExchange.GetTradeHistory(pair, nil, nil)
			if !assert.NoError(t, e) {
				return
			}
			if len(tradeHistoryResult.Trades) > 0 {
				assert.NotNil(t, tradeHistoryResult.Cursor)
			} else {
				assert.Nil(t, tradeHistoryResult.Cursor)
			}

			validateTrades(t, pair, tradeHistoryResult.Trades)
		})
	}
}

func validateTrades(t *testing.T, pair model.TradingPair, trades []model.Trade) {
	for _, trade := range trades {
		if !assert.Equal(t, &pair, trade.Pair) {
			return
		}
		if !assert.True(t, trade.Price.AsFloat() > 0, fmt.Sprintf("%.7f", trade.Price.AsFloat())) {
			return
		}
		if !assert.True(t, trade.Volume.AsFloat() > 0, fmt.Sprintf("%.7f", trade.Volume.AsFloat())) {
			return
		}
		if !assert.Equal(t, trade.OrderType, model.OrderTypeLimit) {
			return
		}
		if !assert.True(t, trade.Timestamp.AsInt64() > 0, fmt.Sprintf("%d", trade.Timestamp.AsInt64())) {
			return
		}
		if !assert.NotNil(t, trade.TransactionID) {
			return
		}
		if !assert.NotNil(t, trade.Fee) {
			return
		}
		if trade.OrderAction != model.OrderActionBuy && trade.OrderAction != model.OrderActionSell {
			assert.Fail(t, "trade.OrderAction should be either OrderActionBuy or OrderActionSell: %v", trade.OrderAction)
			return
		}
		if trade.Cost != nil && !assert.True(t, trade.Cost.AsFloat() > 0, fmt.Sprintf("%s x %s = %s", trade.Price.AsString(), trade.Volume.AsString(), trade.Cost.AsString())) {
			return
		}
	}
}

func TestGetLatestTradeCursor_Ccxt(t *testing.T) {
	for exchangeName, apiKey := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{apiKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			startIntervalMillis := time.Now().UnixNano() / int64(time.Millisecond)
			cursor, e := testCcxtExchange.GetLatestTradeCursor()
			if !assert.NoError(t, e) {
				return
			}
			endIntervalMillis := time.Now().UnixNano() / int64(time.Millisecond)

			if !assert.IsType(t, "string", cursor) {
				return
			}

			cursorString := cursor.(string)
			cursorInt, e := strconv.ParseInt(cursorString, 10, 64)
			if !assert.NoError(t, e) {
				return
			}

			if !assert.True(t, startIntervalMillis <= cursorInt, fmt.Sprintf("returned cursor (%d) should gte the start time of the function call in milliseconds (%d)", cursorInt, startIntervalMillis)) {
				return
			}
			if !assert.True(t, endIntervalMillis >= cursorInt, fmt.Sprintf("returned cursor (%d) should lte the end time of the function call in milliseconds (%d)", cursorInt, endIntervalMillis)) {
				return
			}
		})
	}
}

func TestGetAccountBalances_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, apiKey := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{apiKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			balances, e := testCcxtExchange.GetAccountBalances([]interface{}{
				model.XLM,
				model.BTC,
				model.USD,
			})
			if !assert.NoError(t, e) {
				return
			}

			log.Printf("balances: %+v\n", balances)
			if !assert.Equal(t, 20.0, balances[model.XLM].AsFloat()) {
				return
			}

			if !assert.Equal(t, 0.0, balances[model.BTC].AsFloat()) {
				return
			}

			if !assert.Equal(t, 0.0, balances[model.USD].AsFloat()) {
				return
			}

			assert.Fail(t, "force fail")
		})
	}
}

func TestGetOpenOrders_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	tradingPairs := []model.TradingPair{
		{Base: model.XLM, Quote: model.BTC},
		{Base: model.XLM, Quote: model.USDT},
	}

	for exchangeName, apiKey := range supportedTradingExchanges {
		for _, pair := range tradingPairs {
			t.Run(exchangeName, func(t *testing.T) {
				testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{apiKey}, false)
				if !assert.NoError(t, e) {
					return
				}

				m, e := testCcxtExchange.GetOpenOrders([]*model.TradingPair{&pair})
				if !assert.NoError(t, e) {
					return
				}

				if !assert.Equal(t, 1, len(m)) {
					return
				}

				openOrders := m[pair]
				if !assert.True(t, len(openOrders) > 0, fmt.Sprintf("%d", len(openOrders))) {
					return
				}

				for _, o := range openOrders {
					log.Printf("open order: %+v\n", o)

					isValid := validateOpenOrder(t, &pair, o)
					if !isValid {
						return
					}
				}
				assert.Fail(t, "force fail")
			})
		}
	}
}

func validateOpenOrder(t *testing.T, pair *model.TradingPair, o model.OpenOrder) bool {
	if !assert.Equal(t, pair, o.Order.Pair) {
		return false
	}

	// OrderAction has it's underlying type as a boolean so will always be valid

	if !assert.Equal(t, model.OrderTypeLimit, o.Order.OrderType) {
		return false
	}

	if !assert.True(t, o.Order.Price.AsFloat() > 0, o.Order.Price.AsString()) {
		return false
	}

	if !assert.True(t, o.Order.Volume.AsFloat() > 0, o.Order.Volume.AsString()) {
		return false
	}

	if !assert.NotNil(t, o.Order.Timestamp) {
		return false
	}

	if !assert.True(t, len(o.ID) > 0, o.ID) {
		return false
	}

	if !assert.NotNil(t, o.StartTime) {
		return false
	}

	// ExpireTime is always nil for now
	if !assert.Nil(t, o.ExpireTime) {
		return false
	}

	if !assert.NotNil(t, o.VolumeExecuted) {
		return false
	}

	// additional check to see if the two timestamps match
	if !assert.Equal(t, o.Order.Timestamp, o.StartTime) {
		return false
	}

	return true
}

func TestAddOrder_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, apiKey := range supportedTradingExchanges {
		for _, kase := range []struct {
			pair        *model.TradingPair
			orderAction model.OrderAction
			orderType   model.OrderType
			price       *model.Number
			volume      *model.Number
		}{
			{
				pair:        &model.TradingPair{Base: model.XLM, Quote: model.BTC},
				orderAction: model.OrderActionSell,
				orderType:   model.OrderTypeLimit,
				price:       model.NumberFromFloat(0.000041, 6),
				volume:      model.NumberFromFloat(60.12345678, 6),
			}, {
				pair:        &model.TradingPair{Base: model.XLM, Quote: model.BTC},
				orderAction: model.OrderActionBuy,
				orderType:   model.OrderTypeLimit,
				price:       model.NumberFromFloat(0.000026, 6),
				volume:      model.NumberFromFloat(40.012345, 6),
			}, {
				pair:        &model.TradingPair{Base: model.XLM, Quote: model.USDT},
				orderAction: model.OrderActionSell,
				orderType:   model.OrderTypeLimit,
				price:       model.NumberFromFloat(0.15, 6),
				volume:      model.NumberFromFloat(51.5, 6),
			},
		} {
			t.Run(exchangeName, func(t *testing.T) {
				testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{apiKey}, false)
				if !assert.NoError(t, e) {
					return
				}

				txID, e := testCcxtExchange.AddOrder(&model.Order{
					Pair:        kase.pair,
					OrderAction: kase.orderAction,
					OrderType:   kase.orderType,
					Price:       kase.price,
					Volume:      kase.volume,
				})
				if !assert.NoError(t, e) {
					return
				}

				log.Printf("transactionID from order: %s\n", txID)
				if !assert.NotNil(t, txID) {
					return
				}

				if !assert.NotEqual(t, "", txID.String()) {
					return
				}

				assert.Fail(t, "force fail")
			})
		}
	}
}

func TestCancelOrder_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, apiKey := range supportedTradingExchanges {
		for _, kase := range []struct {
			orderID string
			pair    *model.TradingPair
		}{
			{
				orderID: "",
				pair:    &model.TradingPair{Base: model.XLM, Quote: model.BTC},
			}, {
				orderID: "",
				pair:    &model.TradingPair{Base: model.XLM, Quote: model.USDT},
			},
		} {
			t.Run(exchangeName, func(t *testing.T) {
				testCcxtExchange, e := makeCcxtExchange(exchangeName, testOrderConstraints, []api.ExchangeAPIKey{apiKey}, false)
				if !assert.NoError(t, e) {
					return
				}

				result, e := testCcxtExchange.CancelOrder(model.MakeTransactionID(kase.orderID), *kase.pair)
				if !assert.NoError(t, e) {
					return
				}

				log.Printf("result from cancel order (transactionID=%s): %s\n", kase.orderID, result.String())
				if !assert.Equal(t, model.CancelResultCancelSuccessful, result) {
					return
				}
			})
		}
	}
}
