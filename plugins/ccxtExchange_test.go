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
