package plugins

import (
	"fmt"
	"testing"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

// import (
// 	"fmt"
// 	"log"
// 	"math"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"

// 	"github.com/stellar/kelp/api"
// 	"github.com/stellar/kelp/model"
// )

var testOrderConstraintsBinanceWs = map[string]map[model.TradingPair]model.OrderConstraints{
	"binance": {
		*model.MakeTradingPair(model.XLM, model.USDT): *model.MakeOrderConstraints(4, 5, 0.1),
		*model.MakeTradingPair(model.XLM, model.BTC):  *model.MakeOrderConstraints(8, 4, 1.0),
	},
	"kraken": {
		*model.MakeTradingPair(model.XLM, model.USD): *model.MakeOrderConstraints(6, 8, 30.0),
		*model.MakeTradingPair(model.XLM, model.BTC): *model.MakeOrderConstraints(8, 8, 30.0),
	},
	"bitstamp": {
		*model.MakeTradingPair(model.XLM, model.USD): *model.MakeOrderConstraints(5, 2, 25.0),
	},
}

func Test_createStateEvents(t *testing.T) {

	events := createStateEvents()

	assert.NotNil(t, events)
}

func Test_binanceExchangeWs_GetTickerPrice(t *testing.T) {
	if testing.Short() {
		return
	}

	testCcxtExchange, e := makeCcxtExchange(
		"binance",
		testOrderConstraints["binance"],
		[]api.ExchangeAPIKey{emptyAPIKey},
		[]api.ExchangeParam{emptyParams},
		[]api.ExchangeHeader{},
		false,
		getEsParamFactory("binance"),
	)

	if !assert.NoError(t, e) {
		return
	}

	testBinanceExchangeWs, e := makeBinanceWs(testCcxtExchange.(ccxtExchange))
	if !assert.NoError(t, e) {
		return
	}

	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	pairs := []model.TradingPair{pair}

	m, e := testBinanceExchangeWs.GetTickerPrice(pairs)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, 1, len(m))

	ticker := m[pair]
	assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString())
	assert.True(t, ticker.BidPrice.AsFloat() < 1, ticker.BidPrice.AsString())
	assert.True(t, ticker.BidPrice.AsFloat() < ticker.AskPrice.AsFloat(), fmt.Sprintf("bid price (%s) should be less than ask price (%s)", ticker.BidPrice.AsString(), ticker.AskPrice.AsString()))
	assert.True(t, ticker.LastPrice.AsFloat() < 1, ticker.LastPrice.AsString())
}
