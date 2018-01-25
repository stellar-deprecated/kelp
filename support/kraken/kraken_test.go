package kraken

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/lightyeario/kelp/support/exchange/trades"

	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/orderbook"

	"github.com/Beldur/kraken-go-api-client"
	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/stretchr/testify/assert"
)

var testKrakenExchange exchange.Exchange = krakenExchange{
	assetConverter: assets.KrakenAssetConverter,
	api:            krakenapi.New("", ""),
	delimiter:      "",
	precision:      8,
	isSimulated:    true,
}

func TestGetTickerPrice(t *testing.T) {
	pair := assets.TradingPair{Base: assets.XLM, Quote: assets.BTC}
	pairs := []assets.TradingPair{pair}

	m, e := testKrakenExchange.GetTickerPrice(pairs)
	if !assert.NoError(t, e) {
		return
	}
	assert.Equal(t, 1, len(m))

	ticker := m[pair]
	assert.True(t, ticker.AskPrice.AsFloat() < 1, ticker.AskPrice.AsString())
}

func TestGetAccountBalances(t *testing.T) {
	assetList := []assets.Asset{
		assets.USD,
		assets.XLM,
		assets.BTC,
		assets.LTC,
		assets.ETH,
		assets.REP,
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
}

func TestGetOrderBook(t *testing.T) {
	pair := assets.TradingPair{Base: assets.XLM, Quote: assets.BTC}
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
	pair := assets.TradingPair{Base: assets.XLM, Quote: assets.BTC}
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
}

func TestAddOrder(t *testing.T) {
	txID, e := testKrakenExchange.AddOrder(&orderbook.Order{
		Pair:        &assets.TradingPair{Base: assets.REP, Quote: assets.ETH},
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
}
