package plugins

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

var testKrakenExchange api.Exchange = &krakenExchange{
	assetConverter:           model.KrakenAssetConverter,
	assetConverterOpenOrders: model.KrakenAssetConverterOpenOrders,
	apis:                     []*krakenapi.KrakenApi{krakenapi.New("", "")},
	apiNextIndex:             0,
	delimiter:                "",
	ocOverridesHandler:       MakeEmptyOrderConstraintsOverridesHandler(),
	withdrawKeys:             asset2Address2Key{},
	isSimulated:              true,
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

	if !assert.True(t, len(tradeHistoryResult.Trades) >= 0) {
		return
	}

	if !assert.NotNil(t, tradeHistoryResult.Cursor) {
		return
	}

	assert.Fail(t, "force fail")
}

func makeTrade(txID string, ts time.Time) model.Trade {
	pair := model.TradingPair{Base: model.XLM, Quote: model.USD}

	return model.Trade{
		Order: model.Order{
			Pair:        &pair,
			OrderAction: model.OrderActionBuy,
			OrderType:   model.OrderTypeLimit,
			Price:       model.NumberFromFloat(1.0, 6),
			Volume:      model.NumberFromFloat(10.0, 6),
			Timestamp:   model.MakeTimestampFromTime(ts),
		},
		TransactionID: model.MakeTransactionID(txID),
		Cost:          model.NumberFromFloat(10.0, 6),
		Fee:           model.NumberFromFloat(0.0, 6),
	}
}

func TestGetTradeHistoryAdapter(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Second)
	t3 := t2.Add(time.Second)
	t4 := t3.Add(time.Second)
	t5 := t4.Add(time.Second)
	tx1 := makeTrade("tx1", t1)
	tx2 := makeTrade("tx2", t2)
	tx3 := makeTrade("tx3", t3)
	tx4 := makeTrade("tx4", t4)
	tx5 := makeTrade("tx5", t5)

	testCases := []struct {
		name                    string
		maxTrades               int
		cursor2HistoryResult    map[string]*api.TradeHistoryResult
		maybeCursorEndInclusive *model.TransactionID
		wantCursor              string
		wantTradeIDs            []string
	}{
		{
			name:      "max 3 trades different timings, search key tx5",
			maxTrades: 3,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"tx5": &api.TradeHistoryResult{
					Trades: []model.Trade{tx3, tx4, tx5},
					Cursor: "tx5",
				},
				"tx3": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1, tx2, tx3},
					Cursor: "tx3",
				},
				"tx1": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1},
					Cursor: "tx1",
				},
			},
			maybeCursorEndInclusive: tx5.TransactionID,
			wantCursor:              "tx5",
			wantTradeIDs:            []string{"tx1", "tx2", "tx3", "tx4", "tx5"},
		}, {
			name:      "max 3 trades different timings, search key tx3",
			maxTrades: 3,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"tx3": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1, tx2, tx3},
					Cursor: "tx3",
				},
				"tx1": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1},
					Cursor: "tx1",
				},
			},
			maybeCursorEndInclusive: tx3.TransactionID,
			wantCursor:              "tx3",
			wantTradeIDs:            []string{"tx1", "tx2", "tx3"},
		}, {
			name:      "max 3 trades different timings, search key nil",
			maxTrades: 3,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"(nil)": &api.TradeHistoryResult{
					Trades: []model.Trade{tx3, tx4, tx5},
					Cursor: "tx5",
				},
				"tx3": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1, tx2, tx3},
					Cursor: "tx3",
				},
				"tx1": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1},
					Cursor: "tx1",
				},
			},
			maybeCursorEndInclusive: nil,
			wantCursor:              "tx5",
			wantTradeIDs:            []string{"tx1", "tx2", "tx3", "tx4", "tx5"},
		}, {
			name:      "max 5 trades different timings, search key tx5",
			maxTrades: 5,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"tx5": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1, tx2, tx3, tx4, tx5},
					Cursor: "tx5",
				},
				"tx1": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1},
					Cursor: "tx1",
				},
			},
			maybeCursorEndInclusive: tx5.TransactionID,
			wantCursor:              "tx5",
			wantTradeIDs:            []string{"tx1", "tx2", "tx3", "tx4", "tx5"},
		}, {
			name:      "max 6 trades different timings, search key tx5",
			maxTrades: 6,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"tx5": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1, tx2, tx3, tx4, tx5},
					Cursor: "tx5",
				},
				"tx1": &api.TradeHistoryResult{
					Trades: []model.Trade{tx1},
					Cursor: "tx1",
				},
			},
			maybeCursorEndInclusive: tx5.TransactionID,
			wantCursor:              "tx5",
			wantTradeIDs:            []string{"tx1", "tx2", "tx3", "tx4", "tx5"},
		}, {
			name:      "max 3 trades repeat timings, search key tx5 - kraken API has this behavior :(",
			maxTrades: 3,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"tx5": &api.TradeHistoryResult{
					Trades: []model.Trade{makeTrade("tx3", t4), makeTrade("tx4", t4), makeTrade("tx5", t5)},
					Cursor: "tx5",
				},
				"tx3": &api.TradeHistoryResult{
					Trades: []model.Trade{makeTrade("tx2", t2), makeTrade("tx3", t4), makeTrade("tx4", t4)},
					Cursor: "tx4",
				},
				"tx2": &api.TradeHistoryResult{
					Trades: []model.Trade{makeTrade("tx1", t1), makeTrade("tx2", t2)},
					Cursor: "tx2",
				},
				"tx1": &api.TradeHistoryResult{
					Trades: []model.Trade{makeTrade("tx1", t1)},
					Cursor: "tx1",
				},
			},
			maybeCursorEndInclusive: tx5.TransactionID,
			wantCursor:              "tx5",
			wantTradeIDs:            []string{"tx1", "tx2", "tx3", "tx4", "tx5"},
		}, {
			name:      "max 3 trades repeat timings, search key tx5 - kraken API has this behavior :( - abridged",
			maxTrades: 3,
			cursor2HistoryResult: map[string]*api.TradeHistoryResult{
				"tx5": &api.TradeHistoryResult{
					Trades: []model.Trade{makeTrade("tx3", t4), makeTrade("tx4", t4), makeTrade("tx5", t5)},
					Cursor: "tx5",
				},
				"tx3": &api.TradeHistoryResult{
					Trades: []model.Trade{makeTrade("tx3", t4), makeTrade("tx4", t4)},
					Cursor: "tx4",
				},
			},
			maybeCursorEndInclusive: tx5.TransactionID,
			wantCursor:              "tx5",
			wantTradeIDs:            []string{"tx3", "tx4", "tx5"},
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			// build underlying function API call mock using cursor2HistoryResult
			fetchPartialTradesFromEndAscMock := func(mcei *string) (*api.TradeHistoryResult, error) {
				result, ok := k.cursor2HistoryResult[*mcei]
				if !ok {
					return nil, fmt.Errorf("searched for key in cursor2HistoryResult map that did not exist (%s). Either this is a bug in getTradeHistoryAdapter which should not have requested this key or the test should have included it", *mcei)
				}

				return result, nil
			}

			// convert *model.TransactionID to *string
			mceiString := "(nil)"
			if k.maybeCursorEndInclusive != nil {
				mceiString = k.maybeCursorEndInclusive.String()
			}

			// call function being tested
			tradeHistoryResult, e := getTradeHistoryAdapter(&mceiString, fetchPartialTradesFromEndAscMock)
			if !assert.NoError(t, e) {
				return
			}

			// assert cursor
			if !assert.Equal(t, k.wantCursor, tradeHistoryResult.Cursor) {
				return
			}

			// assert trades
			if !assert.Equal(t, len(k.wantTradeIDs), len(tradeHistoryResult.Trades)) {
				return
			}
			for i, wantTradeID := range k.wantTradeIDs {
				assert.Equal(t, wantTradeID, tradeHistoryResult.Trades[i].TransactionID.String())
			}
		})
	}
}

func TestGetLatestTradeCursor(t *testing.T) {
	startIntervalSecs := time.Now().Unix()
	cursor, e := testKrakenExchange.GetLatestTradeCursor()
	if !assert.NoError(t, e) {
		return
	}
	endIntervalSecs := time.Now().Unix()

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
	txID, e := testKrakenExchange.AddOrder(
		&model.Order{
			Pair:        tradingPair,
			OrderAction: model.OrderActionSell,
			OrderType:   model.OrderTypeLimit,
			Price:       model.NumberFromFloat(5.123456, testKrakenExchange.GetOrderConstraints(tradingPair).PricePrecision),
			Volume:      model.NumberFromFloat(30.12345678, testKrakenExchange.GetOrderConstraints(tradingPair).VolumePrecision),
		},
		api.SubmitModeBoth,
	)
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
