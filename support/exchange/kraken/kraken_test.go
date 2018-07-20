package kraken

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/lightyeario/kelp/support/exchange/api"
	"github.com/lightyeario/kelp/support/exchange/api/trades"

	"github.com/lightyeario/kelp/support/exchange/api/number"
	"github.com/lightyeario/kelp/support/exchange/api/orderbook"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/lightyeario/kelp/model/assets"
	"github.com/stretchr/testify/assert"
)

var testKrakenExchange api.Exchange = krakenExchange{
	assetConverter: model.KrakenAssetConverter,
	api:            krakenapi.New("", ""),
	delimiter:      "",
	precision:      8,
	withdrawKeys:   asset2Address2Key{},
	isSimulated:    true,
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
	assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString())
}

func TestGetAccountBalances(t *testing.T) {
	assetList := []model.Asset{
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
		assert.True(t, bal.AsFloat() > 0, bal.AsString())
	}

	assert.Fail(t, "force fail")
}

func TestGetOrderBook(t *testing.T) {
	pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
	ob, e := testKrakenExchange.GetOrderBook(&pair, 10)
	if !assert.NoError(t, e) {
		return
	}

	assert.True(t, len(ob.Asks()) > 0, len(ob.Asks()))
	assert.True(t, len(ob.Bids()) > 0, len(ob.Bids()))
	assert.True(t, ob.Asks()[0].OrderAction.IsSell())
	assert.True(t, ob.Asks()[0].OrderType.IsLimit())
	assert.True(t, ob.Bids()[0].OrderAction.IsBuy())
	assert.True(t, ob.Bids()[0].OrderType.IsLimit())
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
	tradeHistoryResult, e := testKrakenExchange.GetTradeHistory(nil, nil)
	if !assert.NoError(t, e) {
		return
	}

	// print here for convenience
	fmt.Printf("total number of trades: %d\n", len(tradeHistoryResult.Trades))
	for _, t := range tradeHistoryResult.Trades {
		fmt.Println(t.String())
	}

	assert.True(t, len(tradeHistoryResult.Trades) > 0)

	assert.Fail(t, "force fail")
}

func TestGetOpenOrders(t *testing.T) {
	m, e := testKrakenExchange.GetOpenOrders()
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

	assert.True(t, len(m) > 0, "there were no open orders")

	assert.Fail(t, "force fail")
}

func TestAddOrder(t *testing.T) {
	txID, e := testKrakenExchange.AddOrder(&orderbook.Order{
		Pair:        &model.TradingPair{Base: model.REP, Quote: model.ETH},
		OrderAction: orderbook.ActionBuy,
		OrderType:   orderbook.TypeLimit,
		Price:       number.FromFloat(0.00001, 5),
		Volume:      number.FromFloat(0.3145, 5),
	})
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("transactionID from order: %s\n", txID)
	assert.NotNil(t, txID)

	assert.Fail(t, "force fail")
}

func TestCancelOrder(t *testing.T) {
	// need to add some transactionID here to run this test
	txID := orderbook.MakeTransactionID("")
	result, e := testKrakenExchange.CancelOrder(txID)
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("result from cancel order (transactionID=%s): %s\n", txID.String(), result.String())
	assert.Equal(t, trades.CancelResultCancelSuccessful, result)

	assert.Fail(t, "force fail")
}

func TestPrepareDeposit(t *testing.T) {
	result, e := testKrakenExchange.PrepareDeposit(model.BTC, number.FromFloat(1.0, 7))
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("fee=%v, address=%v, expireTs=%v\n", result.Fee, result.Address, result.ExpireTs)
	assert.Fail(t, "force fail")
}

func TestGetWithdrawInfo(t *testing.T) {
	result, e := testKrakenExchange.GetWithdrawInfo(model.BTC, number.FromFloat(1.0, 7), "")
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("amountToReceive=%v\n", result.AmountToReceive.AsFloat())
	assert.Fail(t, "force fail")
}

func TestWithdrawFunds(t *testing.T) {
	result, e := testKrakenExchange.WithdrawFunds(model.XLM, number.FromFloat(0.0000001, 7), "")
	if !assert.NoError(t, e) {
		return
	}

	fmt.Printf("refid=%v\n", result.WithdrawalID)
	assert.Fail(t, "force fail")
}
