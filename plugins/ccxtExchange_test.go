package plugins

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

type exchangeAuthData struct {
	apiKey api.ExchangeAPIKey
	params []api.ExchangeParam
}

var supportedExchanges = []string{"binance", "coinbasepro"}
var emptyAPIKey = api.ExchangeAPIKey{}
var emptyParams = api.ExchangeParam{}
var supportedTradingExchanges = map[string]exchangeAuthData{
	"binance": exchangeAuthData{
		apiKey: api.ExchangeAPIKey{},
		params: []api.ExchangeParam{},
	},
}

var testOrderConstraints = map[string]map[model.TradingPair]model.OrderConstraints{
	"binance": map[model.TradingPair]model.OrderConstraints{
		*model.MakeTradingPair(model.XLM, model.USDT): *model.MakeOrderConstraints(4, 5, 0.1),
		*model.MakeTradingPair(model.XLM, model.BTC):  *model.MakeOrderConstraints(8, 4, 1.0),
	},
	"kraken": map[model.TradingPair]model.OrderConstraints{
		*model.MakeTradingPair(model.XLM, model.USD): *model.MakeOrderConstraints(6, 8, 30.0),
		*model.MakeTradingPair(model.XLM, model.BTC): *model.MakeOrderConstraints(8, 8, 30.0),
	},
	"bitstamp": map[model.TradingPair]model.OrderConstraints{
		*model.MakeTradingPair(model.XLM, model.USD): *model.MakeOrderConstraints(5, 2, 25.0),
	},
}

func getEsParamFactory(exchangeName string) ccxtExchangeSpecificParamFactory {
	if v, ok := ccxtExchangeSpecificParamFactoryMap["ccxt-"+exchangeName]; ok {
		return v
	}
	return nil
}

func TestGetTickerPrice_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(
				exchangeName,
				testOrderConstraints[exchangeName],
				[]api.ExchangeAPIKey{emptyAPIKey},
				[]api.ExchangeParam{emptyParams},
				[]api.ExchangeHeader{},
				false,
				getEsParamFactory(exchangeName),
			)
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
			assert.True(t, ticker.BidPrice.AsFloat() < 1, ticker.BidPrice.AsString())
			assert.True(t, ticker.BidPrice.AsFloat() < ticker.AskPrice.AsFloat(), fmt.Sprintf("bid price (%s) should be less than ask price (%s)", ticker.BidPrice.AsString(), ticker.AskPrice.AsString()))
			assert.True(t, ticker.LastPrice.AsFloat() < 1, ticker.LastPrice.AsString())
		})
	}
}

func TestGetOrderBook_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		for _, obDepth := range []int32{1, 5, 8, 10, 15, 16, 20} {
			t.Run(fmt.Sprintf("%s_%d", exchangeName, obDepth), func(t *testing.T) {
				testCcxtExchange, e := makeCcxtExchange(
					exchangeName,
					testOrderConstraints[exchangeName],
					[]api.ExchangeAPIKey{emptyAPIKey},
					[]api.ExchangeParam{emptyParams},
					[]api.ExchangeHeader{},
					false,
					getEsParamFactory(exchangeName),
				)
				if !assert.NoError(t, e) {
					return
				}

				pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
				ob, e := testCcxtExchange.GetOrderBook(&pair, obDepth)
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
				assert.True(t, ob.Asks()[0].Price.AsFloat() > 0, ob.Asks()[0].Price.AsString())
				assert.True(t, ob.Asks()[0].Volume.AsFloat() > 0)
				assert.True(t, ob.Bids()[0].Price.AsFloat() > 0, ob.Bids()[0].Price.AsString())
				assert.True(t, ob.Bids()[0].Volume.AsFloat() > 0)
			})
		}
	}
}

func TestGetTrades_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for _, exchangeName := range supportedExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(
				exchangeName,
				testOrderConstraints[exchangeName],
				[]api.ExchangeAPIKey{emptyAPIKey},
				[]api.ExchangeParam{},
				[]api.ExchangeHeader{},
				false,
				getEsParamFactory(exchangeName),
			)
			if !assert.NoError(t, e) {
				return
			}

			pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
			// TODO test with cursor once implemented
			tradeResult, e := testCcxtExchange.GetTrades(&pair, nil)
			if !assert.NoError(t, e) {
				return
			}
			wantCursorInt64 := tradeResult.Trades[len(tradeResult.Trades)-1].Timestamp.AsInt64() + 1
			assert.Equal(t, strconv.FormatInt(wantCursorInt64, 10), tradeResult.Cursor)

			validateTrades(t, pair, tradeResult.Trades)
		})
	}
}

func TestGetTradeHistory_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, authData := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(
				exchangeName,
				testOrderConstraints[exchangeName],
				[]api.ExchangeAPIKey{authData.apiKey},
				authData.params,
				[]api.ExchangeHeader{},
				false,
				getEsParamFactory(exchangeName),
			)
			if !assert.NoError(t, e) {
				return
			}

			pair := model.TradingPair{Base: model.XLM, Quote: model.BTC}
			// TODO test with cursor once implemented
			tradeHistoryResult, e := testCcxtExchange.GetTradeHistory(pair, nil, nil)
			if !assert.NoError(t, e) {
				return
			}
			if len(tradeHistoryResult.Trades) > 0 {
				assert.NotNil(t, tradeHistoryResult.Cursor)
			} else {
				assert.Nil(t, tradeHistoryResult.Cursor)
			}

			validateTrades(t, pair, tradeHistoryResult.Trades)
		})
	}
}

func validateTrades(t *testing.T, pair model.TradingPair, trades []model.Trade) {
	for _, trade := range trades {
		if !assert.Equal(t, &pair, trade.Pair) {
			return
		}
		if !assert.True(t, trade.Price.AsFloat() > 0, trade.Price.AsString()) {
			return
		}
		if !assert.True(t, trade.Volume.AsFloat() > 0, trade.Volume.AsString()) {
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
		if !assert.NotNil(t, trade.Cost) {
			return
		}
		if !assert.NotNil(t, trade.Fee) {
			return
		}
		if trade.OrderAction != model.OrderActionBuy && trade.OrderAction != model.OrderActionSell {
			assert.Fail(t, "trade.OrderAction should be either OrderActionBuy or OrderActionSell: %v", trade.OrderAction)
			return
		}
		minPrecision := math.Min(float64(trade.Price.Precision()), float64(trade.Volume.Precision()))
		nonZeroCalculatedCost := trade.Price.AsFloat()*trade.Volume.AsFloat() > math.Pow(10, -minPrecision)
		if nonZeroCalculatedCost && !assert.True(t, trade.Cost.AsFloat() > 0, fmt.Sprintf("(price) %s x (volume) %s = (cost) %s", trade.Price.AsString(), trade.Volume.AsString(), trade.Cost.AsString())) {
			return
		}
	}
}

func TestGetLatestTradeCursor_Ccxt(t *testing.T) {
	for exchangeName, authData := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(
				exchangeName,
				testOrderConstraints[exchangeName],
				[]api.ExchangeAPIKey{authData.apiKey},
				authData.params,
				[]api.ExchangeHeader{},
				false,
				getEsParamFactory(exchangeName),
			)
			if !assert.NoError(t, e) {
				return
			}

			startIntervalMillis := time.Now().UnixNano() / int64(time.Millisecond)
			cursor, e := testCcxtExchange.GetLatestTradeCursor()
			if !assert.NoError(t, e) {
				return
			}
			endIntervalMillis := time.Now().UnixNano() / int64(time.Millisecond)

			if !assert.IsType(t, "string", cursor) {
				return
			}

			cursorString := cursor.(string)
			cursorInt, e := strconv.ParseInt(cursorString, 10, 64)
			if !assert.NoError(t, e) {
				return
			}

			if !assert.True(t, startIntervalMillis <= cursorInt, fmt.Sprintf("returned cursor (%d) should gte the start time of the function call in milliseconds (%d)", cursorInt, startIntervalMillis)) {
				return
			}
			if !assert.True(t, endIntervalMillis >= cursorInt, fmt.Sprintf("returned cursor (%d) should lte the end time of the function call in milliseconds (%d)", cursorInt, endIntervalMillis)) {
				return
			}
		})
	}
}

func TestGetAccountBalances_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, authData := range supportedTradingExchanges {
		t.Run(exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(
				exchangeName,
				testOrderConstraints[exchangeName],
				[]api.ExchangeAPIKey{authData.apiKey},
				authData.params,
				[]api.ExchangeHeader{},
				false,
				getEsParamFactory(exchangeName),
			)
			if !assert.NoError(t, e) {
				return
			}

			balances, e := testCcxtExchange.GetAccountBalances([]interface{}{
				model.XLM,
				model.BTC,
				model.USD,
			})
			if !assert.NoError(t, e) {
				return
			}

			log.Printf("balances: %+v\n", balances)
			if !assert.Equal(t, 20.0, balances[model.XLM].AsFloat()) {
				return
			}

			if !assert.Equal(t, 0.0, balances[model.BTC].AsFloat()) {
				return
			}

			if !assert.Equal(t, 0.0, balances[model.USD].AsFloat()) {
				return
			}

			assert.Fail(t, "force fail")
		})
	}
}

func TestGetOpenOrders_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	tradingPairs := []model.TradingPair{
		{Base: model.XLM, Quote: model.BTC},
		{Base: model.XLM, Quote: model.USDT},
	}

	for exchangeName, authData := range supportedTradingExchanges {
		for _, pair := range tradingPairs {
			t.Run(exchangeName, func(t *testing.T) {
				testCcxtExchange, e := makeCcxtExchange(
					exchangeName,
					testOrderConstraints[exchangeName],
					[]api.ExchangeAPIKey{authData.apiKey},
					authData.params,
					[]api.ExchangeHeader{},
					false,
					getEsParamFactory(exchangeName),
				)
				if !assert.NoError(t, e) {
					return
				}

				m, e := testCcxtExchange.GetOpenOrders([]*model.TradingPair{&pair})
				if !assert.NoError(t, e) {
					return
				}

				if !assert.Equal(t, 1, len(m)) {
					return
				}

				openOrders := m[pair]
				if !assert.True(t, len(openOrders) > 0, fmt.Sprintf("%d", len(openOrders))) {
					return
				}

				for _, o := range openOrders {
					log.Printf("open order: %+v\n", o)

					isValid := validateOpenOrder(t, &pair, o)
					if !isValid {
						return
					}
				}
				assert.Fail(t, "force fail")
			})
		}
	}
}

func validateOpenOrder(t *testing.T, pair *model.TradingPair, o model.OpenOrder) bool {
	if !assert.Equal(t, pair, o.Order.Pair) {
		return false
	}

	// OrderAction has it's underlying type as a boolean so will always be valid

	if !assert.Equal(t, model.OrderTypeLimit, o.Order.OrderType) {
		return false
	}

	if !assert.True(t, o.Order.Price.AsFloat() > 0, o.Order.Price.AsString()) {
		return false
	}

	if !assert.True(t, o.Order.Volume.AsFloat() > 0, o.Order.Volume.AsString()) {
		return false
	}

	if !assert.NotNil(t, o.Order.Timestamp) {
		return false
	}

	if !assert.True(t, len(o.ID) > 0, o.ID) {
		return false
	}

	if !assert.NotNil(t, o.StartTime) {
		return false
	}

	// ExpireTime is always nil for now
	if !assert.Nil(t, o.ExpireTime) {
		return false
	}

	if !assert.NotNil(t, o.VolumeExecuted) {
		return false
	}

	// additional check to see if the two timestamps match
	if !assert.Equal(t, o.Order.Timestamp, o.StartTime) {
		return false
	}

	return true
}

func TestAddOrder_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	for exchangeName, authData := range supportedTradingExchanges {
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
				testCcxtExchange, e := makeCcxtExchange(
					exchangeName,
					testOrderConstraints[exchangeName],
					[]api.ExchangeAPIKey{authData.apiKey},
					authData.params,
					[]api.ExchangeHeader{},
					false,
					getEsParamFactory(exchangeName),
				)
				if !assert.NoError(t, e) {
					return
				}

				txID, e := testCcxtExchange.AddOrder(
					&model.Order{
						Pair:        kase.pair,
						OrderAction: kase.orderAction,
						OrderType:   kase.orderType,
						Price:       kase.price,
						Volume:      kase.volume,
					},
					api.SubmitModeBoth,
				)
				if !assert.NoError(t, e) {
					return
				}

				log.Printf("transactionID from order: %s\n", txID)
				if !assert.NotNil(t, txID) {
					return
				}

				if !assert.NotEqual(t, "", txID.String()) {
					return
				}

				assert.Fail(t, "force fail")
			})
		}
	}
}

func TestCancelOrder_Ccxt(t *testing.T) {
	if testing.Short() {
		return
	}

	// TODO error converting type and ID for bitstamp
	for exchangeName, authData := range supportedTradingExchanges {
		for _, kase := range []struct {
			orderID string
			pair    *model.TradingPair
		}{
			{
				orderID: "",
				pair:    &model.TradingPair{Base: model.XLM, Quote: model.BTC},
			}, {
				orderID: "",
				pair:    &model.TradingPair{Base: model.XLM, Quote: model.USDT},
			},
		} {
			t.Run(exchangeName, func(t *testing.T) {
				testCcxtExchange, e := makeCcxtExchange(
					exchangeName,
					testOrderConstraints[exchangeName],
					[]api.ExchangeAPIKey{authData.apiKey},
					authData.params,
					[]api.ExchangeHeader{},
					false,
					getEsParamFactory(exchangeName),
				)
				if !assert.NoError(t, e) {
					return
				}

				result, e := testCcxtExchange.CancelOrder(model.MakeTransactionID(kase.orderID), *kase.pair)
				if !assert.NoError(t, e) {
					return
				}

				log.Printf("result from cancel order (transactionID=%s): %s\n", kase.orderID, result.String())
				if !assert.Equal(t, model.CancelResultCancelSuccessful, result) {
					return
				}
			})
		}
	}
}

func TestGetOrderConstraints_Ccxt_Precision(t *testing.T) {
	// coinbasepro gives incorrect precision values so we do not test it here
	testCases := []struct {
		exchangeName       string
		pair               *model.TradingPair
		wantPricePrecision int8
		wantVolPrecision   int8
	}{
		{
			// disable ccxt-kraken based tests for now because of the 403 Forbidden Security check API error
			// 	exchangeName:       "kraken",
			// 	pair:               &model.TradingPair{Base: model.XLM, Quote: model.USD},
			// 	wantPricePrecision: 6,
			// 	wantVolPrecision:   8,
			// }, {
			exchangeName:       "binance",
			pair:               &model.TradingPair{Base: model.XLM, Quote: model.USDT},
			wantPricePrecision: 5,
			wantVolPrecision:   1,
		}, {
			exchangeName:       "binance",
			pair:               &model.TradingPair{Base: model.XLM, Quote: model.BTC},
			wantPricePrecision: 8,
			wantVolPrecision:   8,
		}, {
			exchangeName:       "bitstamp",
			pair:               &model.TradingPair{Base: model.XLM, Quote: model.USD},
			wantPricePrecision: 5,
			wantVolPrecision:   8,
		},
	}

	for _, kase := range testCases {
		t.Run(kase.exchangeName, func(t *testing.T) {
			testCcxtExchange, e := makeCcxtExchange(
				kase.exchangeName,
				nil,
				[]api.ExchangeAPIKey{emptyAPIKey},
				[]api.ExchangeParam{emptyParams},
				[]api.ExchangeHeader{},
				false,
				getEsParamFactory(kase.exchangeName),
			)
			if !assert.NoError(t, e) {
				return
			}

			result := testCcxtExchange.GetOrderConstraints(kase.pair)
			if !assert.Equal(t, kase.wantPricePrecision, result.PricePrecision) {
				return
			}
			if !assert.Equal(t, kase.wantVolPrecision, result.VolumePrecision) {
				return
			}
		})
	}
}
