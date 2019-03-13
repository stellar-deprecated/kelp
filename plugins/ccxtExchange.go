package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

const ccxtBalancePrecision = 10

// ensure that ccxtExchange conforms to the Exchange interface
var _ api.Exchange = ccxtExchange{}

// ccxtExchange is the implementation for the CCXT REST library that supports many exchanges (https://github.com/franz-see/ccxt-rest, https://github.com/ccxt/ccxt/)
type ccxtExchange struct {
	assetConverter   *model.AssetConverter
	delimiter        string
	orderConstraints map[model.TradingPair]model.OrderConstraints
	api              *sdk.Ccxt
	simMode          bool
}

// makeCcxtExchange is a factory method to make an exchange using the CCXT interface
func makeCcxtExchange(
	exchangeName string,
	orderConstraintOverrides map[model.TradingPair]model.OrderConstraints,
	apiKeys []api.ExchangeAPIKey,
	simMode bool,
) (api.Exchange, error) {
	if len(apiKeys) == 0 {
		return nil, fmt.Errorf("need at least 1 ExchangeAPIKey, even if it is an empty key")
	}

	if len(apiKeys) != 1 {
		return nil, fmt.Errorf("need exactly 1 ExchangeAPIKey")
	}

	c, e := sdk.MakeInitializedCcxtExchange(exchangeName, apiKeys[0])
	if e != nil {
		return nil, fmt.Errorf("error making a ccxt exchange: %s", e)
	}

	if orderConstraintOverrides == nil {
		orderConstraintOverrides = map[model.TradingPair]model.OrderConstraints{}
	}

	return ccxtExchange{
		assetConverter:   model.CcxtAssetConverter,
		delimiter:        "/",
		orderConstraints: orderConstraintOverrides,
		api:              c,
		simMode:          simMode,
	}, nil
}

// GetTickerPrice impl.
func (c ccxtExchange) GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]api.Ticker, error) {
	pairsMap, e := model.TradingPairs2Strings(c.assetConverter, c.delimiter, pairs)
	if e != nil {
		return nil, e
	}

	priceResult := map[model.TradingPair]api.Ticker{}
	for _, p := range pairs {
		tickerMap, e := c.api.FetchTicker(pairsMap[p])
		if e != nil {
			return nil, fmt.Errorf("error while fetching ticker price for trading pair %s: %s", pairsMap[p], e)
		}

		askPrice, e := utils.CheckFetchFloat(tickerMap, "ask")
		if e != nil {
			return nil, fmt.Errorf("unable to correctly fetch value from tickerMap: %s", e)
		}
		bidPrice, e := utils.CheckFetchFloat(tickerMap, "bid")
		if e != nil {
			return nil, fmt.Errorf("unable to correctly fetch value from tickerMap: %s", e)
		}

		priceResult[p] = api.Ticker{
			AskPrice: model.NumberFromFloat(askPrice, c.GetOrderConstraints(&p).PricePrecision),
			BidPrice: model.NumberFromFloat(bidPrice, c.GetOrderConstraints(&p).PricePrecision),
		}
	}

	return priceResult, nil
}

// GetAssetConverter impl
func (c ccxtExchange) GetAssetConverter() *model.AssetConverter {
	return c.assetConverter
}

// GetOrderConstraints impl
func (c ccxtExchange) GetOrderConstraints(pair *model.TradingPair) *model.OrderConstraints {
	if oc, ok := c.orderConstraints[*pair]; ok {
		return &oc
	}

	pairString, e := pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		// this should never really panic because we would have converted this trading pair to a string previously
		panic(e)
	}

	// load from CCXT's cache
	ccxtMarket := c.api.GetMarket(pairString)
	if ccxtMarket == nil {
		panic(fmt.Errorf("CCXT does not have precision and limit data for the passed in market: %s", pairString))
	}
	oc := *model.MakeOrderConstraintsWithCost(ccxtMarket.Precision.Price, ccxtMarket.Precision.Amount, ccxtMarket.Limits.Amount.Min, ccxtMarket.Limits.Cost.Min)

	// cache it before returning
	c.orderConstraints[*pair] = oc

	return &oc
}

// GetAccountBalances impl
func (c ccxtExchange) GetAccountBalances(assetList []interface{}) (map[interface{}]model.Number, error) {
	balanceResponse, e := c.api.FetchBalance()
	if e != nil {
		return nil, e
	}

	m := map[interface{}]model.Number{}
	for _, elem := range assetList {
		var asset model.Asset
		if v, ok := elem.(model.Asset); ok {
			asset = v
		} else {
			return nil, fmt.Errorf("invalid type of asset passed in, only model.Asset accepted")
		}

		ccxtAssetString, e := c.GetAssetConverter().ToString(asset)
		if e != nil {
			return nil, e
		}

		if ccxtBalance, ok := balanceResponse[ccxtAssetString]; ok {
			m[asset] = *model.NumberFromFloat(ccxtBalance.Total, ccxtBalancePrecision)
		} else {
			m[asset] = *model.NumberConstants.Zero
		}
	}
	return m, nil
}

// GetOrderBook impl
func (c ccxtExchange) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {
	pairString, e := pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", e)
	}

	limit := int(maxCount)
	ob, e := c.api.FetchOrderBook(pairString, &limit)
	if e != nil {
		return nil, fmt.Errorf("error while fetching orderbook for trading pair '%s': %s", pairString, e)
	}

	if _, ok := ob["asks"]; !ok {
		return nil, fmt.Errorf("orderbook did not contain the 'asks' field: %v", ob)
	}
	if _, ok := ob["bids"]; !ok {
		return nil, fmt.Errorf("orderbook did not contain the 'bids' field: %v", ob)
	}

	asks := c.readOrders(ob["asks"], pair, model.OrderActionSell)
	bids := c.readOrders(ob["bids"], pair, model.OrderActionBuy)
	return model.MakeOrderBook(pair, asks, bids), nil
}

func (c ccxtExchange) readOrders(orders []sdk.CcxtOrder, pair *model.TradingPair, orderAction model.OrderAction) []model.Order {
	pricePrecision := c.GetOrderConstraints(pair).PricePrecision
	volumePrecision := c.GetOrderConstraints(pair).VolumePrecision

	result := []model.Order{}
	for _, o := range orders {
		result = append(result, model.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   model.OrderTypeLimit,
			Price:       model.NumberFromFloat(o.Price, pricePrecision),
			Volume:      model.NumberFromFloat(o.Amount, volumePrecision),
			Timestamp:   nil,
		})
	}
	return result
}

// GetTradeHistory impl
func (c ccxtExchange) GetTradeHistory(pair model.TradingPair, maybeCursorStart interface{}, maybeCursorEnd interface{}) (*api.TradeHistoryResult, error) {
	pairString, e := pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", e)
	}

	// TODO fix limit logic to check result so we get full history instead of just 50 trades
	const limit = 50
	tradesRaw, e := c.api.FetchMyTrades(pairString, limit, maybeCursorStart)
	if e != nil {
		return nil, fmt.Errorf("error while fetching trade history for trading pair '%s': %s", pairString, e)
	}

	trades := []model.Trade{}
	for _, raw := range tradesRaw {
		t, e := c.readTrade(&pair, pairString, raw)
		if e != nil {
			return nil, fmt.Errorf("error while reading trade: %s", e)
		}
		trades = append(trades, *t)
	}

	// TODO implement cursor logic
	return &api.TradeHistoryResult{
		Cursor: nil,
		Trades: trades,
	}, nil
}

// GetTrades impl
func (c ccxtExchange) GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*api.TradesResult, error) {
	pairString, e := pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", e)
	}

	// TODO use cursor when fetching trades
	tradesRaw, e := c.api.FetchTrades(pairString)
	if e != nil {
		return nil, fmt.Errorf("error while fetching trades for trading pair '%s': %s", pairString, e)
	}

	trades := []model.Trade{}
	for _, raw := range tradesRaw {
		t, e := c.readTrade(pair, pairString, raw)
		if e != nil {
			return nil, fmt.Errorf("error while reading trade: %s", e)
		}
		trades = append(trades, *t)
	}

	// TODO implement cursor logic
	return &api.TradesResult{
		Cursor: nil,
		Trades: trades,
	}, nil
}

func (c ccxtExchange) readTrade(pair *model.TradingPair, pairString string, rawTrade sdk.CcxtTrade) (*model.Trade, error) {
	if rawTrade.Symbol != pairString {
		return nil, fmt.Errorf("expected '%s' for 'symbol' field, got: %s", pairString, rawTrade.Symbol)
	}

	pricePrecision := c.GetOrderConstraints(pair).PricePrecision
	volumePrecision := c.GetOrderConstraints(pair).VolumePrecision

	trade := model.Trade{
		Order: model.Order{
			Pair:      pair,
			Price:     model.NumberFromFloat(rawTrade.Price, pricePrecision),
			Volume:    model.NumberFromFloat(rawTrade.Amount, volumePrecision),
			OrderType: model.OrderTypeLimit,
			Timestamp: model.MakeTimestamp(rawTrade.Timestamp),
		},
		TransactionID: model.MakeTransactionID(rawTrade.ID),
		Fee:           nil,
	}

	if rawTrade.Side == "sell" {
		trade.OrderAction = model.OrderActionSell
	} else if rawTrade.Side == "buy" {
		trade.OrderAction = model.OrderActionBuy
	} else {
		return nil, fmt.Errorf("unrecognized value for 'side' field: %s", rawTrade.Side)
	}

	if rawTrade.Cost != 0.0 {
		// use bigger precision for cost since it's logically derived from amount and price
		costPrecision := pricePrecision
		if volumePrecision > pricePrecision {
			costPrecision = volumePrecision
		}
		trade.Cost = model.NumberFromFloat(rawTrade.Cost, costPrecision)
	}

	return &trade, nil
}

// GetOpenOrders impl
func (c ccxtExchange) GetOpenOrders(pairs []*model.TradingPair) (map[model.TradingPair][]model.OpenOrder, error) {
	pairStrings := []string{}
	string2Pair := map[string]model.TradingPair{}
	for _, pair := range pairs {
		pairString, e := pair.ToString(c.assetConverter, c.delimiter)
		if e != nil {
			return nil, fmt.Errorf("error converting pairs to strings: %s", e)
		}
		pairStrings = append(pairStrings, pairString)
		string2Pair[pairString] = *pair
	}

	openOrdersMap, e := c.api.FetchOpenOrders(pairStrings)
	if e != nil {
		return nil, fmt.Errorf("error while fetching open orders for trading pairs '%v': %s", pairStrings, e)
	}

	result := map[model.TradingPair][]model.OpenOrder{}
	for asset, ccxtOrderList := range openOrdersMap {
		pair, ok := string2Pair[asset]
		if !ok {
			return nil, fmt.Errorf("symbol %s returned from FetchOpenOrders was not in the original list of trading pairs: %v", asset, pairStrings)
		}

		openOrderList := []model.OpenOrder{}
		for _, o := range ccxtOrderList {
			openOrder, e := c.convertOpenOrderFromCcxt(&pair, o)
			if e != nil {
				return nil, fmt.Errorf("cannot convertOpenOrderFromCcxt: %s", e)
			}
			openOrderList = append(openOrderList, *openOrder)
		}
		result[pair] = openOrderList
	}
	return result, nil
}

func (c ccxtExchange) convertOpenOrderFromCcxt(pair *model.TradingPair, o sdk.CcxtOpenOrder) (*model.OpenOrder, error) {
	if o.Type != "limit" {
		return nil, fmt.Errorf("we currently only support limit order types")
	}

	orderAction := model.OrderActionSell
	if o.Side == "buy" {
		orderAction = model.OrderActionBuy
	}
	ts := model.MakeTimestamp(o.Timestamp)

	return &model.OpenOrder{
		Order: model.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   model.OrderTypeLimit,
			Price:       model.NumberFromFloat(o.Price, c.GetOrderConstraints(pair).PricePrecision),
			Volume:      model.NumberFromFloat(o.Amount, c.GetOrderConstraints(pair).VolumePrecision),
			Timestamp:   ts,
		},
		ID:             o.ID,
		StartTime:      ts,
		ExpireTime:     nil,
		VolumeExecuted: model.NumberFromFloat(o.Filled, c.GetOrderConstraints(pair).VolumePrecision),
	}, nil
}

// AddOrder impl
func (c ccxtExchange) AddOrder(order *model.Order) (*model.TransactionID, error) {
	pairString, e := order.Pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", e)
	}

	side := "sell"
	if order.OrderAction.IsBuy() {
		side = "buy"
	}

	log.Printf("ccxt is submitting order: pair=%s, orderAction=%s, orderType=%s, volume=%s, price=%s\n",
		pairString, order.OrderAction.String(), order.OrderType.String(), order.Volume.AsString(), order.Price.AsString())
	ccxtOpenOrder, e := c.api.CreateLimitOrder(pairString, side, order.Volume.AsFloat(), order.Price.AsFloat())
	if e != nil {
		return nil, fmt.Errorf("error while creating limit order %s: %s", *order, e)
	}

	return model.MakeTransactionID(ccxtOpenOrder.ID), nil
}

// CancelOrder impl
func (c ccxtExchange) CancelOrder(txID *model.TransactionID, pair model.TradingPair) (model.CancelOrderResult, error) {
	log.Printf("ccxt is canceling order: ID=%s, tradingPair: %s\n", txID.String(), pair.String())

	resp, e := c.api.CancelOrder(txID.String(), pair.String())
	if e != nil {
		return model.CancelResultFailed, e
	}

	if resp == nil {
		return model.CancelResultFailed, fmt.Errorf("response from CancelOrder was nil")
	}
	return model.CancelResultCancelSuccessful, nil
}

// PrepareDeposit impl
func (c ccxtExchange) PrepareDeposit(asset model.Asset, amount *model.Number) (*api.PrepareDepositResult, error) {
	// TODO implement
	return nil, nil
}

// GetWithdrawInfo impl
func (c ccxtExchange) GetWithdrawInfo(asset model.Asset, amountToWithdraw *model.Number, address string) (*api.WithdrawInfo, error) {
	// TODO implement
	return nil, nil
}

// WithdrawFunds impl
func (c ccxtExchange) WithdrawFunds(
	asset model.Asset,
	amountToWithdraw *model.Number,
	address string,
) (*api.WithdrawFunds, error) {
	// TODO implement
	return nil, nil
}
