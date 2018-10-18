package plugins

import (
	"fmt"
	"testing"

	"github.com/lightyeario/kelp/model"
	"github.com/stretchr/testify/assert"
)

var supportedExchanges = []string{"binance", "poloniex", "bittrex"}

func TestGetTickerPrice_Ccxt(t *testing.T) {
	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName)
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
	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName)
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
	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", exchangeName)
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
