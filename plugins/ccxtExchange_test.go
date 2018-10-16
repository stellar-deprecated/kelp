package plugins

import (
	"testing"

	"github.com/lightyeario/kelp/model"
	"github.com/stretchr/testify/assert"
)

func TestGetTickerPrice_Ccxt(t *testing.T) {
	testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", "binance")
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
}

func TestGetOrderBook_Ccxt(t *testing.T) {
	testCcxtExchange, e := makeCcxtExchange("http://localhost:3000", "binance")
	if !assert.NoError(t, e) {
		return
	}

	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	ob, e := testCcxtExchange.GetOrderBook(&pair, 10)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, ob.Pair(), &pair)

	assert.True(t, len(ob.Asks()) > 0, len(ob.Asks()))
	assert.True(t, len(ob.Bids()) > 0, len(ob.Bids()))
	assert.True(t, ob.Asks()[0].OrderAction.IsSell())
	assert.True(t, ob.Asks()[0].OrderType.IsLimit())
	assert.True(t, ob.Bids()[0].OrderAction.IsBuy())
	assert.True(t, ob.Bids()[0].OrderType.IsLimit())
	assert.True(t, ob.Asks()[0].Price.AsFloat() > 0)
	assert.True(t, ob.Asks()[0].Volume.AsFloat() > 0)
	assert.True(t, ob.Bids()[0].Price.AsFloat() > 0)
	assert.True(t, ob.Bids()[0].Volume.AsFloat() > 0)
}
