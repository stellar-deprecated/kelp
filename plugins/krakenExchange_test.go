package plugins

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stellar/kelp/api"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

var testKrakenExchange api.Exchange = &krakenExchange{
	assetConverter:           model.KrakenAssetConverter,
	assetConverterOpenOrders: model.KrakenAssetConverterOpenOrders,
	apis:               []*krakenapi.KrakenApi{krakenapi.New("", "")},
	apiNextIndex:       0,
	delimiter:          "",
	ocOverridesHandler: MakeEmptyOrderConstraintsOverridesHandler(),
	withdrawKeys:       asset2Address2Key{},
	isSimulated:        true,
}

func TestGetTickerPrice(t *testing.T) {
	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	pairs := []model.TradingPair{pair}

	m, e := testKrakenExchange.GetTickerPrice(pairs)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, 1, len(m))

	ticker := m[pair]
	fmt.Printf("ticker price: bid=%.8f, ask=%.8f\n", ticker.BidPrice.AsFloat(), ticker.AskPrice.AsFloat())

	if !assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString()) {
		return
	}
	if !assert.True(t, ticker.BidPrice.AsFloat() < 1, ticker.BidPrice.AsString()) {
		return
	}
}

func TestGetAccountBalances(t *testing.T) {
	if testing.Short() {
		return
	}

	assetList := []interface{}{
		model.USD,
		model.XLM,
		model.BTC,
		model.LTC,
		model.ETH,
		model.REP,
	}
	m, e := testKrakenExchange.GetAccountBalances(assetList)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, 6, len(m))

	// print balances here for convenience
	for _, assetKey := range assetList {
		fmt.Printf("Balance %s = %.8f\n", assetKey, m[assetKey].AsFloat())
	}

	for _, a := range assetList {
		bal := m[a]
		assert.True(t, bal.AsFloat() >= 0, bal.AsString())
	}

	assert.Fail(t, "force fail")
}

func TestGetOrderBook(t *testing.T) {
	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	ob, e := testKrakenExchange.GetOrderBook(&pair, 10)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, ob.Pair(), &pair)

	if !assert.True(t, len(ob.Asks()) > 0, len(ob.Asks())) {
		return
	}
	if !assert.True(t, len(ob.Bids()) > 0, len(ob.Bids())) {
		return
	}
	assert.True(t, ob.Asks()[0].OrderAction.IsSell())
	assert.True(t, ob.Asks()[0].OrderType.IsLimit())
	assert.True(t, ob.Bids()[0].OrderAction.IsBuy())
	assert.True(t, ob.Bids()[0].OrderType.IsLimit())
	assert.True(t, ob.Asks()[0].Price.AsFloat() > 0)
	assert.True(t, ob.Asks()[0].Volume.AsFloat() > 0)
	assert.True(t, ob.Bids()[0].Price.AsFloat() > 0)
	assert.True(t, ob.Bids()[0].Volume.AsFloat() > 0)

	// print here for convenience
	fmt.Printf("first 2 bids:\n")
	fmt.Println(ob.Bids()[0])
	fmt.Println(ob.Bids()[1])
	fmt.Printf("first 2 asks:\n")
	fmt.Println(ob.Asks()[0])
	fmt.Println(ob.Asks()[1])
}

func TestGetTrades(t *testing.T) {
	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	trades, e := testKrakenExchange.GetTrades(&pair, nil)
	if !assert.NoError(t, e) {
		return
	}

	cursor := trades.Cursor.(int64)
	assert.True(t, cursor > 0, strconv.FormatInt(cursor, 10))
	assert.True(t, len(trades.Trades) > 0)

	// print here for convenience
	fmt.Printf("total number of trades: %d\n", len(trades.Trades))
	for _, t := range trades.Trades {
		fmt.Println(t.String())
	}

	// assert.Fail(t, "force fail")
}

func TestGetTradeHistory(t *testing.T) {
	if testing.Short() {
		return
	}

	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	tradeHistoryResult, e := testKrakenExchange.GetTradeHistory(pair, nil, nil)
	if !assert.NoError(t, e) {
		return
	}

	// print here for convenience
	fmt.Printf("total number of trades: %d\n", len(tradeHistoryResult.Trades))
	for _, t := range tradeHistoryResult.Trades {
		fmt.Println(t.String())
	}

	if !assert.True(t, len(tradeHistoryResult.Trades) >= 0) {
		return
	}

	if !assert.NotNil(t, tradeHistoryResult.Cursor) {
		return
	}

	assert.Fail(t, "force fail")
}

func TestGetLatestTradeCursor(t *testing.T) {
	startIntervalSecs := time.Now().Unix() * 1000
	cursor, e := testKrakenExchange.GetLatestTradeCursor()
	if !assert.NoError(t, e) {
		return
	}
	endIntervalSecs := time.Now().Unix() * 1000

	if !assert.IsType(t, "string", cursor) {
		return
	}

	cursorString := cursor.(string)
	cursorInt, e := strconv.ParseInt(cursorString, 10, 64)
	if !assert.NoError(t, e) {
		return
	}

	if !assert.True(t, startIntervalSecs <= cursorInt, fmt.Sprintf("returned cursor (%d) should be gte the start time of the function call in millis (%d)", cursorInt, startIntervalSecs)) {
		return
	}
	if !assert.True(t, endIntervalSecs >= cursorInt, fmt.Sprintf("returned cursor (%d) should be lte the end time of the function call in millis (%d)", cursorInt, endIntervalSecs)) {
		return
	}
}

func TestGetOpenOrders(t *testing.T) {
	if testing.Short() {
		return
	}

	pair := &model.TradingPair{Base: model.XLM, Quote: model.USD}
	m, e := testKrakenExchange.GetOpenOrders([]*model.TradingPair{pair})
	if !assert.NoError(t, e) {
		return
	}

	// print open orders here for convenience
	for pair, openOrders := range m {
		fmt.Printf("Open Orders for pair: %s\n", pair.String())
		for _, o := range openOrders {
			fmt.Printf("    %s\n", o.String())
		}
	}

	if !assert.True(t, len(m) > 0, "there were no open orders") {
		return
	}

	assert.Fail(t, "force fail")
}

func TestAddOrder(t *testing.T) {
	if testing.Short() {
		return
	}

	tradingPair := &model.TradingPair{Base: model.XLM, Quote: model.USD}
	txID, e := testKrakenExchange.AddOrder(&model.Order{
		Pair:        tradingPair,
		OrderAction: model.OrderActionSell,
		OrderType:   model.OrderTypeLimit,
		Price:       model.NumberFromFloat(5.123456, testKrakenExchange.GetOrderConstraints(tradingPair).PricePrecision),
		Volume:      model.NumberFromFloat(30.12345678, testKrakenExchange.GetOrderConstraints(tradingPair).VolumePrecision),
	})
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("transactionID from order: %s\n", txID)
	if !assert.NotNil(t, txID) {
		return
	}

	assert.Fail(t, "force fail")
}

func TestCancelOrder(t *testing.T) {
	if testing.Short() {
		return
	}

	// need to add some transactionID here to run this test
	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	txID := model.MakeTransactionID("")
	result, e := testKrakenExchange.CancelOrder(txID, pair)
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("result from cancel order (transactionID=%s): %s\n", txID.String(), result.String())
	if !assert.Equal(t, model.CancelResultCancelSuccessful, result) {
		return
	}

	assert.Fail(t, "force fail")
}

func TestPrepareDeposit(t *testing.T) {
	if testing.Short() {
		return
	}

	result, e := testKrakenExchange.PrepareDeposit(model.BTC, model.NumberFromFloat(1.0, 7))
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("fee=%v, address=%v, expireTs=%v\n", result.Fee, result.Address, result.ExpireTs)
	assert.Fail(t, "force fail")
}

func TestGetWithdrawInfo(t *testing.T) {
	if testing.Short() {
		return
	}

	result, e := testKrakenExchange.GetWithdrawInfo(model.BTC, model.NumberFromFloat(1.0, 7), "")
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("amountToReceive=%v\n", result.AmountToReceive.AsFloat())
	assert.Fail(t, "force fail")
}

func TestWithdrawFunds(t *testing.T) {
	if testing.Short() {
		return
	}

	result, e := testKrakenExchange.WithdrawFunds(model.XLM, model.NumberFromFloat(0.0000001, 7), "")
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("refid=%v\n", result.WithdrawalID)
	assert.Fail(t, "force fail")
}
