package plugins

import (
	"fmt"
	"testing"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

var supportedExchanges = []string{"binance", "poloniex", "bittrex"}
var emptyAPIKey = api.ExchangeAPIKey{}
var supportedTradingExchanges = map[string]api.ExchangeAPIKey{
	"binance": api.ExchangeAPIKey{},
}

func TestGetTickerPrice_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName, []api.ExchangeAPIKey{emptyAPIKey}, false)
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
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName, []api.ExchangeAPIKey{emptyAPIKey}, false)
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
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName, []api.ExchangeAPIKey{emptyAPIKey}, false)
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

			for _, trade := range tradeResult.Trades {
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
				if !assert.Nil(t, trade.Fee) {
					return
				}
				if trade.OrderAction != model.OrderActionBuy && trade.OrderAction != model.OrderActionSell {
					assert.Fail(t, "trade.OrderAction should be either OrderActionBuy or OrderActionSell: %v", trade.OrderAction)
					return
				}
				if trade.Cost != nil && !assert.True(t, trade.Cost.AsFloat() > 0, fmt.Sprintf("%.7f", trade.Cost.AsFloat())) {
					return
				}
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
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName, []api.ExchangeAPIKey{apiKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			balances, e := testCcxtExchange.GetAccountBalances([]model.Asset{
				model.XLM,
				model.BTC,
				model.USD,
			})
			if !assert.NoError(t, e) {
				return
			}

			if !assert.Equal(t, balances[model.XLM].AsFloat(), 20.0) {
				return
			}

			if !assert.Equal(t, balances[model.BTC].AsFloat(), 0.0) {
				return
			}

			if !assert.Equal(t, balances[model.USD].AsFloat(), 0.0) {
				return
			}
		})
	}
}

func TestGetOpenOrders_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, apiKey := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName, []api.ExchangeAPIKey{apiKey}, false)
			if !assert.NoError(t, e) {
				return
			}

			pair := &model.TradingPair{Base: model.XLM, Quote: model.BTC}
			m, e := testCcxtExchange.GetOpenOrders([]*model.TradingPair{pair})
			if !assert.NoError(t, e) {
				return
			}

			if !assert.Equal(t, 1, len(m)) {
				return
			}

			openOrders := m[*pair]
			if !assert.True(t, len(openOrders) > 0, fmt.Sprintf("%d", len(openOrders))) {
				return
			}

			for _, o := range openOrders {
				if !assert.Equal(t, pair, o.Order.Pair) {
					return
				}

				// OrderAction has it's underlying type as a boolean so will always be valid

				if !assert.Equal(t, model.OrderTypeLimit, o.Order.OrderType) {
					return
				}

				if !assert.True(t, o.Order.Price.AsFloat() > 0, o.Order.Price.AsString()) {
					return
				}

				if !assert.True(t, o.Order.Volume.AsFloat() > 0, o.Order.Volume.AsString()) {
					return
				}

				if !assert.NotNil(t, o.Order.Timestamp) {
					return
				}

				if !assert.True(t, len(o.ID) > 0, o.ID) {
					return
				}

				if !assert.NotNil(t, o.StartTime) {
					return
				}

				// ExpireTime is always nil for now
				if !assert.Nil(t, o.ExpireTime) {
					return
				}

				if !assert.NotNil(t, o.VolumeExecuted) {
					return
				}

				// additional check to see if the two timestamps match
				if !assert.Equal(t, o.Order.Timestamp, o.StartTime) {
					return
				}
			}
		})
	}
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
				testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName, []api.ExchangeAPIKey{apiKey}, false)
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

				fmt.Printf("transactionID from order: %s\n", txID)
				if !assert.NotNil(t, txID) {
					return
				}

				if !assert.NotEqual(t, "", txID.String()) {
					return
				}
			})
		}
	}
}
