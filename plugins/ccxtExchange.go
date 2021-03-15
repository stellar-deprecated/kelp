package plugins

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

const ccxtBalancePrecision = 10

// ensure that ccxtExchange conforms to the Exchange interface
var _ api.Exchange = ccxtExchange{}

// ccxtExchangeSpecificParamFactory knows how to create the exchange-specific params for each exchange
type ccxtExchangeSpecificParamFactory interface {
	getInitParams() map[string]interface{}
	getParamsForGetOrderBook() map[string]interface{}
	getParamsForAddOrder(submitMode api.SubmitMode) interface{}
	getParamsForGetTradeHistory() interface{}
	useSignToDenoteSideForTrades() bool
	getCursorFetchTrades(model.Trade) (interface{}, error)
}

// ccxtExchange is the implementation for the CCXT REST library that supports many exchanges (https://github.com/franz-see/ccxt-rest, https://github.com/ccxt/ccxt/)
type ccxtExchange struct {
	assetConverter     model.AssetConverterInterface
	delimiter          string
	ocOverridesHandler *OrderConstraintsOverridesHandler
	api                *sdk.Ccxt
	simMode            bool
	esParamFactory     ccxtExchangeSpecificParamFactory
}

// makeCcxtExchange is a factory method to make an exchange using the CCXT interface
func makeCcxtExchange(
	exchangeName string,
	orderConstraintOverrides map[model.TradingPair]model.OrderConstraints,
	apiKeys []api.ExchangeAPIKey,
	exchangeParams []api.ExchangeParam,
	headers []api.ExchangeHeader,
	simMode bool,
	esParamFactory ccxtExchangeSpecificParamFactory,
) (api.Exchange, error) {
	if len(apiKeys) == 0 {
		return nil, fmt.Errorf("need at least 1 ExchangeAPIKey, even if it is an empty key")
	}

	if len(apiKeys) != 1 {
		return nil, fmt.Errorf("need exactly 1 ExchangeAPIKey")
	}

	defaultExchangeParams := []api.ExchangeParam{}
	if esParamFactory != nil {
		initParamMap := esParamFactory.getInitParams()
		if initParamMap != nil {
			for k, v := range initParamMap {
				defaultExchangeParams = append(defaultExchangeParams, api.ExchangeParam{
					Param: k,
					Value: v,
				})
			}
		}
	}
	if len(defaultExchangeParams) > 0 {
		// prepend default params so we can override from config if needed
		exchangeParams = append(defaultExchangeParams, exchangeParams...)
	}
	c, e := sdk.MakeInitializedCcxtExchange(exchangeName, apiKeys[0], exchangeParams, headers)
	if e != nil {
		return nil, fmt.Errorf("error making a ccxt exchange: %s", e)
	}

	ocOverridesHandler := MakeEmptyOrderConstraintsOverridesHandler()
	if orderConstraintOverrides != nil {
		ocOverridesHandler = MakeOrderConstraintsOverridesHandler(orderConstraintOverrides)
	}

	return ccxtExchange{
		assetConverter:     model.CcxtAssetConverter,
		delimiter:          "/",
		ocOverridesHandler: ocOverridesHandler,
		api:                c,
		simMode:            simMode,
		esParamFactory:     esParamFactory,
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
			return nil, fmt.Errorf("unable to correctly fetch 'ask' value from tickerMap: %s", e)
		}
		bidPrice, e := utils.CheckFetchFloat(tickerMap, "bid")
		if e != nil {
			return nil, fmt.Errorf("unable to correctly fetch 'bid' value from tickerMap: %s", e)
		}
		lastPrice, e := utils.CheckFetchFloat(tickerMap, "last")
		if e != nil {
			return nil, fmt.Errorf("unable to correctly fetch 'last' value from tickerMap: %s", e)
		}

		pricePrecision := c.GetOrderConstraints(&p).PricePrecision
		priceResult[p] = api.Ticker{
			AskPrice:  model.NumberFromFloat(askPrice, pricePrecision),
			BidPrice:  model.NumberFromFloat(bidPrice, pricePrecision),
			LastPrice: model.NumberFromFloat(lastPrice, pricePrecision),
		}
	}

	return priceResult, nil
}

// GetAssetConverter impl
func (c ccxtExchange) GetAssetConverter() model.AssetConverterInterface {
	return c.assetConverter
}

// GetOrderConstraints impl
func (c ccxtExchange) GetOrderConstraints(pair *model.TradingPair) *model.OrderConstraints {
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
	volumePrecision := ccxtMarket.Precision.Amount
	if volumePrecision == 0 {
		volumePrecision = ccxtMarket.Precision.Price
	}
	oc := model.MakeOrderConstraintsWithCost(ccxtMarket.Precision.Price, volumePrecision, ccxtMarket.Limits.Amount.Min, ccxtMarket.Limits.Cost.Min)

	return c.ocOverridesHandler.Apply(pair, oc)
}

// OverrideOrderConstraints impl, can partially override values for specific pairs
func (c ccxtExchange) OverrideOrderConstraints(pair *model.TradingPair, override *model.OrderConstraintsOverride) {
	c.ocOverridesHandler.Upsert(pair, override)
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
	maxCountInt := int(maxCount)

	// if the exchange has limitations on how many orders we can fetch, these params will handle it and adjust the fetchLimit accordingly
	fetchLimit := maxCountInt
	if c.esParamFactory != nil {
		orderbookParamsMap := c.esParamFactory.getParamsForGetOrderBook()
		if orderbookParamsMap != nil {
			if transformLimitFnResult, ok := orderbookParamsMap["transform_limit"]; ok {
				transformLimitFn := transformLimitFnResult.(func(limit int) (int, error))
				newLimit, e := transformLimitFn(int(maxCount))
				if e != nil {
					return nil, fmt.Errorf("error while transforming maxCount limit: %s", e)
				}
				fetchLimit = newLimit
			}
		}
	}

	pairString, e := pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", e)
	}

	ob, e := c.api.FetchOrderBook(pairString, &fetchLimit)
	if e != nil {
		return nil, fmt.Errorf("error while fetching orderbook for trading pair '%s': %s", pairString, e)
	}
	if _, ok := ob["asks"]; !ok {
		return nil, fmt.Errorf("orderbook did not contain the 'asks' field: %v", ob)
	}
	if _, ok := ob["bids"]; !ok {
		return nil, fmt.Errorf("orderbook did not contain the 'bids' field: %v", ob)
	}

	askCcxtOrders := ob["asks"]
	bidCcxtOrders := ob["bids"]
	if fetchLimit != maxCountInt {
		// we may not have fetched all the requested levels because the exchange may not have had that many levels in depth
		if len(askCcxtOrders) > maxCountInt {
			askCcxtOrders = askCcxtOrders[:maxCountInt]
		}
		if len(bidCcxtOrders) > maxCountInt {
			bidCcxtOrders = bidCcxtOrders[:maxCountInt]
		}
	}

	asks := c.readOrders(askCcxtOrders, pair, model.OrderActionSell)
	bids := c.readOrders(bidCcxtOrders, pair, model.OrderActionBuy)
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

	var maybeExchangeSpecificParams interface{}
	if c.esParamFactory != nil {
		maybeExchangeSpecificParams = c.esParamFactory.getParamsForGetTradeHistory()
	}

	trades := []model.Trade{}
	for _, raw := range tradesRaw {
		var t *model.Trade
		t, e = c.readTrade(&pair, pairString, raw)
		if e != nil {
			return nil, fmt.Errorf("error while reading trade: %s", e)
		}

		orderID := ""
		if maybeExchangeSpecificParams != nil {
			paramsMap := maybeExchangeSpecificParams.(map[string]interface{})
			if oidRes, ok := paramsMap["order_id"]; ok {
				oidFn := oidRes.(func(info interface{}) (string, error))
				orderID, e = oidFn(raw.Info)
				if e != nil {
					return nil, fmt.Errorf("error while reading 'order_id' from raw.Info for exchange with specific params: %s", e)
				}
			}
		}
		t.OrderID = orderID

		trades = append(trades, *t)
	}

	sort.Sort(model.TradesByTsID(trades))
	cursor := maybeCursorStart
	if len(trades) > 0 {
		cursor, e = c.getCursor(trades)
		if e != nil {
			return nil, fmt.Errorf("error getting cursor when fetching trades: %s", e)
		}
	}

	return &api.TradeHistoryResult{
		Cursor: cursor,
		Trades: trades,
	}, nil
}

// GetLatestTradeCursor impl.
func (c ccxtExchange) GetLatestTradeCursor() (interface{}, error) {
	timeNowMillis := time.Now().UnixNano() / int64(time.Millisecond)
	latestTradeCursor := fmt.Sprintf("%d", timeNowMillis)
	return latestTradeCursor, nil
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
		var t *model.Trade
		t, e = c.readTrade(pair, pairString, raw)
		if e != nil {
			return nil, fmt.Errorf("error while reading trade: %s", e)
		}
		trades = append(trades, *t)
	}

	sort.Sort(model.TradesByTsID(trades))
	cursor := maybeCursor
	if len(trades) > 0 {
		cursor, e = c.getCursor(trades)
		if e != nil {
			return nil, fmt.Errorf("error getting cursor when fetching trades: %s", e)
		}
	}

	return &api.TradesResult{
		Cursor: cursor,
		Trades: trades,
	}, nil
}

func (c ccxtExchange) getCursor(trades []model.Trade) (interface{}, error) {
	lastTrade := trades[len(trades)-1]

	var fetchedCursor interface{}
	var e error
	// getCursor from Trade object for specific exchange
	if c.esParamFactory != nil {
		fetchedCursor, e = c.esParamFactory.getCursorFetchTrades(lastTrade)
		if e != nil {
			return nil, fmt.Errorf("tried to convert string cursor to int64 in exchange-specific getCursor method but returned an error: %s", e)
		}
	}

	// fetched cursor can be nil after fetching from the esParamFactory in case the function is not implemented
	if fetchedCursor == nil {
		lastCursor := lastTrade.Order.Timestamp.AsInt64()
		// add 1 to lastCursor so we don't repeat the same cursor on the next run
		fetchedCursor = strconv.FormatInt(lastCursor+1, 10)
	}

	// update cursor accordingly
	return fetchedCursor, nil
}

func (c ccxtExchange) readTrade(pair *model.TradingPair, pairString string, rawTrade sdk.CcxtTrade) (*model.Trade, error) {
	if rawTrade.Symbol != pairString {
		return nil, fmt.Errorf("expected '%s' for 'symbol' field, got: %s", pairString, rawTrade.Symbol)
	}

	pricePrecision := c.GetOrderConstraints(pair).PricePrecision
	volumePrecision := c.GetOrderConstraints(pair).VolumePrecision
	// use bigger precision for fee and cost since they are logically derived from amount and price
	feecCostPrecision := pricePrecision
	if volumePrecision > pricePrecision {
		feecCostPrecision = volumePrecision
	}

	trade := model.Trade{
		Order: model.Order{
			Pair:      pair,
			Price:     model.NumberFromFloat(rawTrade.Price, pricePrecision),
			Volume:    model.NumberFromFloat(rawTrade.Amount, volumePrecision),
			OrderType: model.OrderTypeLimit,
			Timestamp: model.MakeTimestamp(rawTrade.Timestamp),
		},
		TransactionID: model.MakeTransactionID(rawTrade.ID),
		Cost:          model.NumberFromFloat(rawTrade.Cost, feecCostPrecision),
		Fee:           model.NumberFromFloat(rawTrade.Fee.Cost, feecCostPrecision),
		// OrderID read by calling function depending on override set for exchange params in "orderId" field of Info object
	}

	useSignToDenoteSide := false
	if c.esParamFactory != nil {
		useSignToDenoteSide = c.esParamFactory.useSignToDenoteSideForTrades()
	}

	if rawTrade.Side == "sell" {
		trade.OrderAction = model.OrderActionSell
	} else if rawTrade.Side == "buy" {
		trade.OrderAction = model.OrderActionBuy
	} else if useSignToDenoteSide {
		if trade.Cost.AsFloat() < 0 {
			trade.OrderAction = model.OrderActionSell
			trade.Order.Volume = trade.Order.Volume.Scale(-1.0)
			trade.Cost = trade.Cost.Scale(-1.0)
		} else {
			trade.OrderAction = model.OrderActionBuy
		}
	} else {
		return nil, fmt.Errorf("unrecognized value for 'side' field: %s (rawTrade = %+v)", rawTrade.Side, rawTrade)
	}

	if trade.Cost.AsFloat() < 0 {
		return nil, fmt.Errorf("trade.Cost was < 0 (%f)", trade.Cost.AsFloat())
	}
	if trade.Order.Volume.AsFloat() < 0 {
		return nil, fmt.Errorf("trade.Order.Volume was < 0 (%f)", trade.Order.Volume.AsFloat())
	}

	if rawTrade.Cost == 0.0 {
		trade.Cost = model.NumberFromFloat(rawTrade.Price*rawTrade.Amount, feecCostPrecision)
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
	// bitstamp does not use "limit" as the order type but has an empty string. this is reasonable general logic to support so added here instead of specifically for bitstamp
	if o.Type != "limit" && o.Type != "" {
		return nil, fmt.Errorf("we currently only support limit order types: %+v", o)
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
func (c ccxtExchange) AddOrder(order *model.Order, submitMode api.SubmitMode) (*model.TransactionID, error) {
	pairString, e := order.Pair.ToString(c.assetConverter, c.delimiter)
	if e != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", e)
	}

	side := "sell"
	if order.OrderAction.IsBuy() {
		side = "buy"
	}

	log.Printf("ccxt is submitting order: pair=%s, orderAction=%s, orderType=%s, volume=%s, price=%s, submitMode=%s\n",
		pairString, order.OrderAction.String(), order.OrderType.String(), order.Volume.AsString(), order.Price.AsString(), submitMode.String())

	var maybeExchangeSpecificParams interface{}
	if c.esParamFactory != nil {
		maybeExchangeSpecificParams = c.esParamFactory.getParamsForAddOrder(submitMode)
	}
	ccxtOpenOrder, e := c.api.CreateLimitOrder(pairString, side, order.Volume.AsFloat(), order.Price.AsFloat(), maybeExchangeSpecificParams)
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
