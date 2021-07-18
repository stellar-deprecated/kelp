package plugins

import (
	"fmt"
	"testing"

	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

func Test_createStateEvents(t *testing.T) {

	events := createStateEvents()

	assert.NotNil(t, events)
}

func Test_binanceExchangeWs_GetTickerPrice(t *testing.T) {
	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	pairs := []model.TradingPair{pair}

	testBinanceExchangeWs, err := makeBinanceWs()

	if !assert.NoError(t, err) {
		return
	}

	m, e := testBinanceExchangeWs.GetTickerPrice(pairs)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, 1, len(m))

	ticker := m[pair]
	fmt.Printf("ticker price: bid=%s, ask=%s, last=%s\n", ticker.BidPrice.AsString(), ticker.AskPrice.AsString(), ticker.LastPrice.AsString())

	if !assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString()) {
		return
	}
	if !assert.True(t, ticker.BidPrice.AsFloat() < 1, ticker.BidPrice.AsString()) {
		return
	}
	if !assert.True(t, ticker.BidPrice.AsFloat() < ticker.AskPrice.AsFloat(), fmt.Sprintf("bid price (%s) should be less than ask price (%s)", ticker.BidPrice.AsString(), ticker.AskPrice.AsString())) {
		return
	}
	if !assert.True(t, ticker.LastPrice.AsFloat() < 1, ticker.LastPrice.AsString()) {
		return
	}
}
