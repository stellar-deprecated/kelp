package plugins

import (
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/Beldur/kraken-go-api-client"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/networking"
)

// ensure that krakenExchange conforms to the Exchange interface
var _ api.Exchange = &krakenExchange{}

const precisionBalances = 10
const tradesFetchSleepTimeSeconds = 60

// krakenExchange is the implementation for the Kraken Exchange
type krakenExchange struct {
	assetConverter           *model.AssetConverter
	assetConverterOpenOrders *model.AssetConverter // kraken uses different symbols when fetching open orders!
	apis                     []*krakenapi.KrakenApi
	apiNextIndex             uint8
	delimiter                string
	ocOverridesHandler       *OrderConstraintsOverridesHandler
	withdrawKeys             asset2Address2Key
	isSimulated              bool // will simulate add and cancel orders if this is true
}

type asset2Address2Key map[model.Asset]map[string]string

func (m asset2Address2Key) getKey(asset model.Asset, address string) (string, error) {
	address2Key, ok := m[asset]
	if !ok {
		return "", fmt.Errorf("asset (%v) is not registered in asset2Address2Key: %v", asset, m)
	}

	key, ok := address2Key[address]
	if !ok {
		return "", fmt.Errorf("address is not registered in asset2Address2Key: %v (asset = %v)", address, asset)
	}

	return key, nil
}

// makeKrakenExchange is a factory method to make the kraken exchange
// TODO 2, should take in config file for withdrawalKeys mapping
func makeKrakenExchange(apiKeys []api.ExchangeAPIKey, isSimulated bool) (api.Exchange, error) {
	if len(apiKeys) == 0 || len(apiKeys) > math.MaxUint8 {
		return nil, fmt.Errorf("invalid number of apiKeys: %d", len(apiKeys))
	}

	krakenAPIs := []*krakenapi.KrakenApi{}
	for _, apiKey := range apiKeys {
		krakenAPIClient := krakenapi.New(apiKey.Key, apiKey.Secret)
		krakenAPIs = append(krakenAPIs, krakenAPIClient)
	}

	return &krakenExchange{
		assetConverter:           model.KrakenAssetConverter,
		assetConverterOpenOrders: model.KrakenAssetConverterOpenOrders,
		apis:                     krakenAPIs,
		apiNextIndex:             0,
		delimiter:                "",
		ocOverridesHandler:       MakeEmptyOrderConstraintsOverridesHandler(),
		withdrawKeys:             asset2Address2Key{},
		isSimulated:              isSimulated,
	}, nil
}

// nextAPI rotates the API key being used so we can overcome rate limit issues
func (k *krakenExchange) nextAPI() *krakenapi.KrakenApi {
	log.Printf("returning kraken API key at index %d", k.apiNextIndex)
	api := k.apis[k.apiNextIndex]
	// rotate key for the next call
	k.apiNextIndex = (k.apiNextIndex + 1) % uint8(len(k.apis))
	return api
}

// AddOrder impl.
func (k *krakenExchange) AddOrder(order *model.Order, submitMode api.SubmitMode) (*model.TransactionID, error) {
	pairStr, e := order.Pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	if k.isSimulated {
		log.Printf("not adding order to Kraken in simulation mode, order=%s\n", *order)
		return model.MakeTransactionID("simulated"), nil
	}

	orderConstraints := k.GetOrderConstraints(order.Pair)
	if order.Price.Precision() > orderConstraints.PricePrecision {
		return nil, fmt.Errorf("kraken price precision can be a maximum of %d, got %d, value = %.12f", orderConstraints.PricePrecision, order.Price.Precision(), order.Price.AsFloat())
	}
	if order.Volume.Precision() > orderConstraints.VolumePrecision {
		return nil, fmt.Errorf("kraken volume precision can be a maximum of %d, got %d, value = %.12f", orderConstraints.VolumePrecision, order.Volume.Precision(), order.Volume.AsFloat())
	}

	args := map[string]string{
		"price": order.Price.AsString(),
	}
	if submitMode == api.SubmitModeMakerOnly {
		args["oflags"] = "post" // csv list as a string for multiple flags
	}
	log.Printf("kraken is submitting order: pair=%s, orderAction=%s, orderType=%s, volume=%s, price=%s, submitMode=%s\n",
		pairStr, order.OrderAction.String(), order.OrderType.String(), order.Volume.AsString(), order.Price.AsString(), submitMode.String())
	resp, e := k.nextAPI().AddOrder(
		pairStr,
		order.OrderAction.String(),
		order.OrderType.String(),
		order.Volume.AsString(),
		args,
	)
	if e != nil {
		return nil, e
	}

	// expected case for production orders
	if len(resp.TransactionIds) == 1 {
		return model.MakeTransactionID(resp.TransactionIds[0]), nil
	}

	if len(resp.TransactionIds) > 1 {
		return nil, fmt.Errorf("there was more than 1 transctionId: %s", resp.TransactionIds)
	}

	return nil, fmt.Errorf("no transactionIds returned from order creation")
}

// CancelOrder impl.
func (k *krakenExchange) CancelOrder(txID *model.TransactionID, pair model.TradingPair) (model.CancelOrderResult, error) {
	if k.isSimulated {
		return model.CancelResultCancelSuccessful, nil
	}
	log.Printf("kraken is canceling order: ID=%s, tradingPair=%s\n", txID.String(), pair.String())

	// we don't actually use the pair for kraken
	resp, e := k.nextAPI().CancelOrder(txID.String())
	if e != nil {
		return model.CancelResultFailed, e
	}

	if resp.Count > 1 {
		log.Printf("warning: count from a cancelled order is greater than 1: %d\n", resp.Count)
	}

	// TODO 2 - need to figure out whether count = 0 could also mean that it is pending cancellation
	if resp.Count == 0 {
		return model.CancelResultFailed, nil
	}
	// resp.Count == 1 here

	if resp.Pending {
		return model.CancelResultPending, nil
	}
	return model.CancelResultCancelSuccessful, nil
}

// GetAccountBalances impl.
func (k *krakenExchange) GetAccountBalances(assetList []interface{}) (map[interface{}]model.Number, error) {
	balanceResponse, e := k.nextAPI().Balance()
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

		krakenAssetString, e := k.assetConverter.ToString(asset)
		if e != nil {
			// discard partially built map for now
			return nil, e
		}
		bal := getFieldValue(*balanceResponse, krakenAssetString)
		m[asset] = *model.NumberFromFloat(bal, precisionBalances)
	}
	return m, nil
}

func getFieldValue(object krakenapi.BalanceResponse, fieldName string) float64 {
	r := reflect.ValueOf(object)
	f := reflect.Indirect(r).FieldByName(fieldName)
	return f.Interface().(float64)
}

// GetOrderConstraints impl
func (k *krakenExchange) GetOrderConstraints(pair *model.TradingPair) *model.OrderConstraints {
	oc, ok := krakenPrecisionMatrix[*pair]
	if ok {
		return k.ocOverridesHandler.Apply(pair, &oc)
	}

	if k.ocOverridesHandler.IsCompletelyOverriden(pair) {
		override := k.ocOverridesHandler.Get(pair)
		return model.MakeOrderConstraintsFromOverride(override)
	}
	panic(fmt.Sprintf("krakenExchange could not find orderConstraints for trading pair %v. Try using the \"ccxt-kraken\" integration instead.", pair))
}

// OverrideOrderConstraints impl, can partially override values for specific pairs
func (k *krakenExchange) OverrideOrderConstraints(pair *model.TradingPair, override *model.OrderConstraintsOverride) {
	k.ocOverridesHandler.Upsert(pair, override)
}

// GetAssetConverter impl.
func (k *krakenExchange) GetAssetConverter() model.AssetConverterInterface {
	return k.assetConverter
}

// GetOpenOrders impl.
func (k *krakenExchange) GetOpenOrders(pairs []*model.TradingPair) (map[model.TradingPair][]model.OpenOrder, error) {
	openOrdersResponse, e := k.nextAPI().OpenOrders(map[string]string{})
	if e != nil {
		return nil, fmt.Errorf("cannot load open orders for Kraken: %s", e)
	}

	// convert to a map so we can easily search for the existence of a trading pair
	// kraken uses different symbols when fetching open orders!
	pairsMap, e := model.TradingPairs2Strings2(k.assetConverterOpenOrders, "", pairs)
	if e != nil {
		return nil, e
	}

	assetConverters := []model.AssetConverterInterface{*k.assetConverterOpenOrders, model.Display}
	m := map[model.TradingPair][]model.OpenOrder{}
	for ID, o := range openOrdersResponse.Open {
		// kraken uses different symbols when fetching open orders!
		pair, e := model.TradingPairFromString2(3, assetConverters, o.Description.AssetPair)
		if e != nil {
			return nil, fmt.Errorf("error parsing trading pair '%s' in krakenExchange#GetOpenOrders: %s", o.Description.AssetPair, e)
		}

		if _, ok := pairsMap[*pair]; !ok {
			// skip open orders for pairs that were not requested
			continue
		}

		if _, ok := m[*pair]; !ok {
			m[*pair] = []model.OpenOrder{}
		}
		if _, ok := m[model.TradingPair{Base: pair.Quote, Quote: pair.Base}]; ok {
			return nil, fmt.Errorf("open orders are listed with repeated base/quote pairs for %s", *pair)
		}

		orderConstraints := k.GetOrderConstraints(pair)
		m[*pair] = append(m[*pair], model.OpenOrder{
			Order: model.Order{
				Pair:        pair,
				OrderAction: model.OrderActionFromString(o.Description.Type),
				OrderType:   model.OrderTypeFromString(o.Description.OrderType),
				Price:       model.MustNumberFromString(o.Description.PrimaryPrice, orderConstraints.PricePrecision),
				Volume:      model.MustNumberFromString(o.Volume, orderConstraints.VolumePrecision),
				Timestamp:   model.MakeTimestamp(int64(o.OpenTime)),
			},
			ID:             ID,
			StartTime:      model.MakeTimestamp(int64(o.StartTime)),
			ExpireTime:     model.MakeTimestamp(int64(o.ExpireTime)),
			VolumeExecuted: model.NumberFromFloat(o.VolumeExecuted, orderConstraints.VolumePrecision),
		})
	}
	return m, nil
}

// GetOrderBook impl.
func (k *krakenExchange) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {
	pairStr, e := pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	krakenob, e := k.nextAPI().Depth(pairStr, int(maxCount))
	if e != nil {
		return nil, e
	}

	asks := k.readOrders(krakenob.Asks, pair, model.OrderActionSell)
	bids := k.readOrders(krakenob.Bids, pair, model.OrderActionBuy)
	ob := model.MakeOrderBook(pair, asks, bids)
	return ob, nil
}

func (k *krakenExchange) readOrders(obi []krakenapi.OrderBookItem, pair *model.TradingPair, orderAction model.OrderAction) []model.Order {
	orderConstraints := k.GetOrderConstraints(pair)
	orders := []model.Order{}
	for _, item := range obi {
		orders = append(orders, model.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   model.OrderTypeLimit,
			Price:       model.NumberFromFloat(item.Price, orderConstraints.PricePrecision),
			Volume:      model.NumberFromFloat(item.Amount, orderConstraints.VolumePrecision),
			Timestamp:   model.MakeTimestamp(item.Ts),
		})
	}
	return orders
}

// GetTickerPrice impl.
func (k *krakenExchange) GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]api.Ticker, error) {
	pairsMap, e := model.TradingPairs2Strings(k.assetConverter, k.delimiter, pairs)
	if e != nil {
		return nil, e
	}

	resp, e := k.nextAPI().Ticker(values(pairsMap)...)
	if e != nil {
		return nil, e
	}

	priceResult := map[model.TradingPair]api.Ticker{}
	for _, p := range pairs {
		orderConstraints := k.GetOrderConstraints(&p)
		pairTickerInfo := resp.GetPairTickerInfo(pairsMap[p])
		priceResult[p] = api.Ticker{
			AskPrice:  model.MustNumberFromString(pairTickerInfo.Ask[0], orderConstraints.PricePrecision),
			BidPrice:  model.MustNumberFromString(pairTickerInfo.Bid[0], orderConstraints.PricePrecision),
			LastPrice: model.MustNumberFromString(pairTickerInfo.Close[0], orderConstraints.PricePrecision),
		}
	}

	return priceResult, nil
}

// values gives you the values of a map
// TODO 2 - move to autogenerated generic function
func values(m map[model.TradingPair]string) []string {
	values := []string{}
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// GetTradeHistory impl.
func (k *krakenExchange) GetTradeHistory(pair model.TradingPair, maybeCursorStartExclusive interface{}, maybeCursorEndInclusive interface{}) (*api.TradeHistoryResult, error) {
	var mcs *string
	if maybeCursorStartExclusive != nil {
		i := maybeCursorStartExclusive.(string)
		mcs = &i
	}

	var mce *string
	if maybeCursorEndInclusive != nil {
		i := maybeCursorEndInclusive.(string)
		mce = &i
	}

	fetchPartialTradesFromEndAsc := func(mcei *string) (*api.TradeHistoryResult, error) {
		return k.getTradeHistoryFromEndAscLimit50(pair, mcs, mcei)
	}
	return getTradeHistoryAdapter(mce, fetchPartialTradesFromEndAsc)
}

// getTradeHistoryAdapter is an adapter method against the kraken API because the kraken API returns results from the end instead of the beginning and only 50 at a time.
// we iterate over the fetchPartialTradesFromEndAsc method to solve this problem
// this is not attached to krakenExchange because we should be able to inject a fetchPartialTradesFromEndAsc that is unrelated to kraken for testing
// fetchPartialTradesFromEndAsc:
// fetchPartialTradesFromEndAsc will return an incomplete list of trades. If it returns a list of 0 items, or a list of items we have seen previously, then we have exhausted the search
// the start cursor and trading pair is bound to fetchPartialTradesFromEndAsc already.
// trades returned from fetchPartialTradesFromEndAsc are in ascending order with the cursor set to the last tradeID
// example:
// if we have trades with cursor1-cursor100 then calls to fetchPartialTradesFromEndAsc would return the following after
// adjusting maybeCursorEndInclusive: 81-100, 61-80, 41-60, 21-40, 1-20
func getTradeHistoryAdapter(
	maybeCursorEndInclusive *string,
	fetchPartialTradesFromEndAsc func(maybeCursorEndInclusive *string) (*api.TradeHistoryResult, error),
) (*api.TradeHistoryResult, error) {
	// accummulate results of the internal calls here, ignoring memory limits for now since these objects are small
	res := &api.TradeHistoryResult{
		Trades: []model.Trade{},
		Cursor: nil,
	}
	// dedupe trades with the same transactionID using this map
	seenTxIDs := map[string]bool{}

	for {
		innerRes, e := fetchPartialTradesFromEndAsc(maybeCursorEndInclusive)
		if e != nil {
			if strings.Contains(e.Error(), "EAPI:Rate limit exceeded") {
				log.Printf("error fetching trade history 50 at a time from the end in ascending order from kraken (%s). Sleeping for 60 seconds and then retrying request...", e)
				time.Sleep(time.Duration(tradesFetchSleepTimeSeconds) * time.Second)

				log.Printf("... retrying fetching of trades now")
				continue
			}
			return nil, fmt.Errorf("error fetching trade history 50 at a time from the end in ascending order from kraken: %s", e)
		}

		var tradesToPrepend []model.Trade
		if res.Cursor == nil {
			// for the first iteration we want to set the cursor and add all trades
			res.Cursor = innerRes.Cursor
			log.Printf("set cursor to innerRes.Cursor value '%s'\n", innerRes.Cursor)
			tradesToPrepend = innerRes.Trades
		} else {
			// for subsequent iterations we want to only prepend new trades. sometimes trades can be repeated between API calls by the inner Kraken API :(
			// this happens when there are multiple trades with the same timestamp
			for _, trade := range innerRes.Trades {
				if _, seen := seenTxIDs[trade.TransactionID.String()]; !seen {
					tradesToPrepend = append(tradesToPrepend, trade)
				}
			}
		}
		// update seenTxIDs with the new trades
		for _, trade := range tradesToPrepend {
			seenTxIDs[trade.TransactionID.String()] = true
		}

		// prepend to outer result since we are fetching from the back
		res.Trades = append(tradesToPrepend, res.Trades...)
		numSeen := len(innerRes.Trades) - len(tradesToPrepend)
		log.Printf("prepended %d new trades from API result of %d trades (i.e. there were total %d trades seen earlier; expecting 1 seen earlier for all but the first request in the series); total length of trades is now %d\n", len(tradesToPrepend), len(innerRes.Trades), numSeen, len(res.Trades))

		// this is the terminal condition for this function
		// Kraken should return exactly 50 items, but this is a more future-proof check, since we only check that there are no new trades now
		if len(tradesToPrepend) == 0 {
			log.Printf("there were no new trades, returning from getTradeHistoryAdapter\n")
			return res, nil
		}

		// update state to continue fetching trades; set first transactionID of inner result as the new cursor end (inclusive). leave cursor start as-is (exclusive).
		firstTxID := innerRes.Trades[0].TransactionID.String()
		maybeCursorEndInclusive = &firstTxID
		log.Printf("updated value of maybeCursorEndInclusive pointer to '%s'\n", firstTxID)
	}
}

// getTradeHistoryFromEndAscLimit50 fetches trades from the cursor end, in ascending order, limited to 50 entries.
// the backwards iteration is a limitation of the Kraken API which requires us to have an intermediary getTradeHistoryAdapter() method
func (k *krakenExchange) getTradeHistoryFromEndAscLimit50(tradingPair model.TradingPair, maybeCursorStartExclusive *string, maybeCursorEndInclusive *string) (*api.TradeHistoryResult, error) {
	var startCursorLogString, endCursorLogString = "(nil)", "(nil)"
	input := map[string]string{}
	if maybeCursorStartExclusive != nil {
		input["start"] = *maybeCursorStartExclusive
		startCursorLogString = *maybeCursorStartExclusive
	}
	if maybeCursorEndInclusive != nil {
		input["end"] = *maybeCursorEndInclusive
		endCursorLogString = *maybeCursorEndInclusive
	}
	log.Printf("fetching trade history from end ascending with a limit of 50 tradingPair=%s, maybeCursorStartExclusive=%s, maybeCursorEndInclusive=%s\n", tradingPair.String(), startCursorLogString, endCursorLogString)

	resp, e := k.nextAPI().Query("TradesHistory", input)
	if e != nil {
		return nil, e
	}
	krakenResp := resp.(map[string]interface{})
	krakenTrades := krakenResp["trades"].(map[string]interface{})

	res := api.TradeHistoryResult{Trades: []model.Trade{}}
	for _txid, v := range krakenTrades {
		m := v.(map[string]interface{})
		_time := m["time"].(float64)
		ts := model.MakeTimestamp(int64(_time) * 1000)
		_type := m["type"].(string)
		_ordertype := m["ordertype"].(string)
		_price := m["price"].(string)
		_vol := m["vol"].(string)
		_cost := m["cost"].(string)
		_fee := m["fee"].(string)
		_pair := m["pair"].(string)
		var pair *model.TradingPair
		pair, e = model.TradingPairFromString(4, k.assetConverter, _pair)
		if e != nil {
			return nil, fmt.Errorf("error parsing trading pair '%s' in krakenExchange#getTradeHistoryFromEndAscLimit50: %s", _pair, e)
		}
		orderConstraints := k.GetOrderConstraints(pair)
		// for now use the max precision between price and volume for fee and cost
		feeCostPrecision := orderConstraints.PricePrecision
		if orderConstraints.VolumePrecision > feeCostPrecision {
			feeCostPrecision = orderConstraints.VolumePrecision
		}

		if *pair == tradingPair {
			res.Trades = append(res.Trades, model.Trade{
				Order: model.Order{
					Pair:        pair,
					OrderAction: model.OrderActionFromString(_type),
					OrderType:   model.OrderTypeFromString(_ordertype),
					Price:       model.MustNumberFromString(_price, orderConstraints.PricePrecision),
					Volume:      model.MustNumberFromString(_vol, orderConstraints.VolumePrecision),
					Timestamp:   ts,
				},
				TransactionID: model.MakeTransactionID(_txid),
				Cost:          model.MustNumberFromString(_cost, feeCostPrecision),
				Fee:           model.MustNumberFromString(_fee, feeCostPrecision),
				// OrderID unavailable?
			})
		}
	}

	// sort to be in ascending order
	sort.Sort(model.TradesByTsID(res.Trades))

	// set correct value for cursor
	if len(res.Trades) > 0 {
		// use transaction IDs for updates to cursor
		// TODO this should use timestamp in seconds based on email communication with kraken team
		res.Cursor = res.Trades[len(res.Trades)-1].TransactionID.String()
	} else if maybeCursorStartExclusive != nil {
		res.Cursor = *maybeCursorStartExclusive
	} else {
		res.Cursor = nil
	}

	return &res, nil
}

// GetLatestTradeCursor impl.
func (k *krakenExchange) GetLatestTradeCursor() (interface{}, error) {
	timeNowSecs := time.Now().Unix()
	latestTradeCursor := fmt.Sprintf("%d", timeNowSecs)
	return latestTradeCursor, nil
}

// GetTrades impl.
func (k *krakenExchange) GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*api.TradesResult, error) {
	if maybeCursor != nil {
		mc := maybeCursor.(int64)
		return k.getTrades(pair, &mc)
	}
	return k.getTrades(pair, nil)
}

func (k *krakenExchange) getTrades(pair *model.TradingPair, maybeCursor *int64) (*api.TradesResult, error) {
	pairStr, e := pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	var tradesResp *krakenapi.TradesResponse
	if maybeCursor != nil {
		tradesResp, e = k.nextAPI().Trades(pairStr, *maybeCursor)
	} else {
		tradesResp, e = k.nextAPI().Trades(pairStr, -1)
	}
	if e != nil {
		return nil, e
	}

	orderConstraints := k.GetOrderConstraints(pair)
	tradesResult := &api.TradesResult{
		Cursor: tradesResp.Last,
		Trades: []model.Trade{},
	}
	for _, tInfo := range tradesResp.Trades {
		action, e := getAction(tInfo)
		if e != nil {
			return nil, e
		}
		orderType, e := getOrderType(tInfo)
		if e != nil {
			return nil, e
		}

		tradesResult.Trades = append(tradesResult.Trades, model.Trade{
			Order: model.Order{
				Pair:        pair,
				OrderAction: action,
				OrderType:   orderType,
				Price:       model.NumberFromFloat(tInfo.PriceFloat, orderConstraints.PricePrecision),
				Volume:      model.NumberFromFloat(tInfo.VolumeFloat, orderConstraints.VolumePrecision),
				Timestamp:   model.MakeTimestamp(tInfo.Time),
			},
			// TransactionID unavailable
			// don't know if OrderID is available
			// Cost unavailable
			// Fee unavailable
		})
	}

	// sort to be in ascending order
	sort.Sort(model.TradesByTsID(tradesResult.Trades))
	// cursor is already set using the result from the kraken go sdk, so no need to set again here

	return tradesResult, nil
}

func getAction(tInfo krakenapi.TradeInfo) (model.OrderAction, error) {
	if tInfo.Buy {
		return model.OrderActionBuy, nil
	} else if tInfo.Sell {
		return model.OrderActionSell, nil
	}

	// return OrderActionBuy as nil value
	return model.OrderActionBuy, errors.New("unidentified trade action")
}

func getOrderType(tInfo krakenapi.TradeInfo) (model.OrderType, error) {
	if tInfo.Market {
		return model.OrderTypeMarket, nil
	} else if tInfo.Limit {
		return model.OrderTypeLimit, nil
	}
	return -1, errors.New("unidentified trade action")
}

// GetWithdrawInfo impl.
func (k *krakenExchange) GetWithdrawInfo(
	asset model.Asset,
	amountToWithdraw *model.Number,
	address string,
) (*api.WithdrawInfo, error) {
	krakenAsset, e := k.assetConverter.ToString(asset)
	if e != nil {
		return nil, e
	}

	withdrawKey, e := k.withdrawKeys.getKey(asset, address)
	if e != nil {
		return nil, e
	}
	resp, e := k.nextAPI().Query(
		"WithdrawInfo",
		map[string]string{
			"asset":  krakenAsset,
			"key":    withdrawKey,
			"amount": amountToWithdraw.AsString(),
		},
	)
	if e != nil {
		return nil, e
	}

	return parseWithdrawInfoResponse(resp, amountToWithdraw)
}

func parseWithdrawInfoResponse(resp interface{}, amountToWithdraw *model.Number) (*api.WithdrawInfo, error) {
	switch m := resp.(type) {
	case map[string]interface{}:
		info, e := parseWithdrawInfo(m)
		if e != nil {
			return nil, e
		}
		if info.limit != nil && info.limit.AsFloat() < amountToWithdraw.AsFloat() {
			return nil, api.MakeErrWithdrawAmountAboveLimit(amountToWithdraw, info.limit)
		}
		if info.fee != nil && info.fee.AsFloat() >= amountToWithdraw.AsFloat() {
			return nil, api.MakeErrWithdrawAmountInvalid(amountToWithdraw, info.fee)
		}

		return &api.WithdrawInfo{AmountToReceive: info.amount}, nil
	default:
		return nil, fmt.Errorf("could not parse response type from WithdrawInfo: %s", reflect.TypeOf(m))
	}
}

type withdrawInfo struct {
	limit  *model.Number
	fee    *model.Number
	amount *model.Number
}

func parseWithdrawInfo(m map[string]interface{}) (*withdrawInfo, error) {
	// limit
	limit, e := networking.ParseNumber(m, "limit", "WithdrawInfo")
	if e != nil {
		return nil, e
	}

	// fee
	fee, e := networking.ParseNumber(m, "fee", "WithdrawInfo")
	if e != nil {
		if !strings.HasPrefix(e.Error(), networking.PrefixFieldNotFound) {
			return nil, e
		}
		// fee may be missing in which case it's null
		fee = nil
	}

	// amount
	amount, e := networking.ParseNumber(m, "amount", "WithdrawInfo")
	if e != nil {
		return nil, e
	}

	return &withdrawInfo{
		limit:  limit,
		fee:    fee,
		amount: amount,
	}, nil
}

// PrepareDeposit impl.
func (k *krakenExchange) PrepareDeposit(asset model.Asset, amount *model.Number) (*api.PrepareDepositResult, error) {
	krakenAsset, e := k.assetConverter.ToString(asset)
	if e != nil {
		return nil, e
	}

	dm, e := k.getDepositMethods(krakenAsset)
	if e != nil {
		return nil, e
	}

	if dm.limit != nil && dm.limit.AsFloat() < amount.AsFloat() {
		return nil, api.MakeErrDepositAmountAboveLimit(amount, dm.limit)
	}

	// get any unused address on the account or generate a new address if no existing unused address
	generateNewAddress := false
	for {
		addressList, e := k.getDepositAddress(krakenAsset, dm.method, generateNewAddress)
		if e != nil {
			if strings.Contains(e.Error(), "EFunding:Too many addresses") {
				return nil, api.MakeErrTooManyDepositAddresses()
			}
			return nil, e
		}
		// TODO 2 - filter addresses that may be "in progress" - save suggested address on account before using and filter using that list
		// discard addresses that have been used up
		addressList = keepOnlyNew(addressList)

		if len(addressList) > 0 {
			earliestAddress := addressList[len(addressList)-1]
			return &api.PrepareDepositResult{
				Fee:      dm.fee,
				Address:  earliestAddress.address,
				ExpireTs: earliestAddress.expireTs,
			}, nil
		}

		// error if we just tried to generate a new address which failed
		if generateNewAddress {
			return nil, fmt.Errorf("attempt to generate a new address failed")
		}

		// retry the loop by attempting to generate a new address
		generateNewAddress = true
	}
}

func keepOnlyNew(addressList []depositAddress) []depositAddress {
	ret := []depositAddress{}
	for _, a := range addressList {
		if a.isNew {
			ret = append(ret, a)
		}
	}
	return ret
}

type depositMethod struct {
	method     string
	limit      *model.Number
	fee        *model.Number
	genAddress bool
}

func (k *krakenExchange) getDepositMethods(asset string) (*depositMethod, error) {
	resp, e := k.nextAPI().Query(
		"DepositMethods",
		map[string]string{"asset": asset},
	)
	if e != nil {
		return nil, e
	}

	switch arr := resp.(type) {
	case []interface{}:
		switch m := arr[0].(type) {
		case map[string]interface{}:
			return parseDepositMethods(m)
		default:
			return nil, fmt.Errorf("could not parse inner response type of returned []interface{} from DepositMethods: %s", reflect.TypeOf(m))
		}
	default:
		return nil, fmt.Errorf("could not parse response type from DepositMethods: %s", reflect.TypeOf(arr))
	}
}

type depositAddress struct {
	address  string
	expireTs int64
	isNew    bool
}

func (k *krakenExchange) getDepositAddress(asset string, method string, genAddress bool) ([]depositAddress, error) {
	input := map[string]string{
		"asset":  asset,
		"method": method,
	}
	if genAddress {
		// only set "new" if it's supposed to be 'true'. If you set it to 'false' then it will be treated as true by Kraken :(
		input["new"] = "true"
	}
	resp, e := k.nextAPI().Query("DepositAddresses", input)
	if e != nil {
		return []depositAddress{}, e
	}

	addressList := []depositAddress{}
	switch arr := resp.(type) {
	case []interface{}:
		for _, elem := range arr {
			switch m := elem.(type) {
			case map[string]interface{}:
				da, e := parseDepositAddress(m)
				if e != nil {
					return []depositAddress{}, e
				}
				addressList = append(addressList, *da)
			default:
				return []depositAddress{}, fmt.Errorf("could not parse inner response type of returned []interface{} from DepositAddresses: %s", reflect.TypeOf(m))
			}
		}
	default:
		return []depositAddress{}, fmt.Errorf("could not parse response type from DepositAddresses: %s", reflect.TypeOf(arr))
	}
	return addressList, nil
}

func parseDepositAddress(m map[string]interface{}) (*depositAddress, error) {
	// address
	address, e := networking.ParseString(m, "address", "DepositAddresses")
	if e != nil {
		return nil, e
	}

	// expiretm
	expireN, e := networking.ParseNumber(m, "expiretm", "DepositAddresses")
	if e != nil {
		return nil, e
	}
	expireTs := int64(expireN.AsFloat())

	// new
	isNew, e := networking.ParseBool(m, "new", "DepositAddresses")
	if e != nil {
		if !strings.HasPrefix(e.Error(), networking.PrefixFieldNotFound) {
			return nil, e
		}
		// new may be missing in which case it's false
		isNew = false
	}

	return &depositAddress{
		address:  address,
		expireTs: expireTs,
		isNew:    isNew,
	}, nil
}

func parseDepositMethods(m map[string]interface{}) (*depositMethod, error) {
	// method
	method, e := networking.ParseString(m, "method", "DepositMethods")
	if e != nil {
		return nil, e
	}

	// limit
	var limit *model.Number
	limB, e := networking.ParseBool(m, "limit", "DepositMethods")
	if e != nil {
		// limit is special as it can be a boolean or a number
		limit, e = networking.ParseNumber(m, "limit", "DepositMethods")
		if e != nil {
			return nil, e
		}
	} else {
		if limB {
			return nil, fmt.Errorf("invalid value for 'limit' as a response from DepositMethods: boolean value of 'limit' should never be 'true' as it should be a number in that case")
		}
		limit = nil
	}

	// fee
	fee, e := networking.ParseNumber(m, "fee", "DepositMethods")
	if e != nil {
		if !strings.HasPrefix(e.Error(), networking.PrefixFieldNotFound) {
			return nil, e
		}
		// fee may be missing in which case it's null
		fee = nil
	}

	// gen-address
	genAddress, e := networking.ParseBool(m, "gen-address", "DepositMethods")
	if e != nil {
		return nil, e
	}

	return &depositMethod{
		method:     method,
		limit:      limit,
		fee:        fee,
		genAddress: genAddress,
	}, nil
}

// WithdrawFunds impl.
func (k *krakenExchange) WithdrawFunds(
	asset model.Asset,
	amountToWithdraw *model.Number,
	address string,
) (*api.WithdrawFunds, error) {
	krakenAsset, e := k.assetConverter.ToString(asset)
	if e != nil {
		return nil, e
	}

	withdrawKey, e := k.withdrawKeys.getKey(asset, address)
	if e != nil {
		return nil, e
	}
	resp, e := k.nextAPI().Query(
		"Withdraw",
		map[string]string{
			"asset":  krakenAsset,
			"key":    withdrawKey,
			"amount": amountToWithdraw.AsString(),
		},
	)
	if e != nil {
		return nil, e
	}

	return parseWithdrawResponse(resp)
}

func parseWithdrawResponse(resp interface{}) (*api.WithdrawFunds, error) {
	switch m := resp.(type) {
	case map[string]interface{}:
		refid, e := networking.ParseString(m, "refid", "Withdraw")
		if e != nil {
			return nil, e
		}
		return &api.WithdrawFunds{
			WithdrawalID: refid,
		}, nil
	default:
		return nil, fmt.Errorf("could not parse response type from Withdraw: %s", reflect.TypeOf(m))
	}
}

// krakenPrecisionMatrix describes the price and volume precision and min base volume for each trading pair
// taken from this URL: https://support.kraken.com/hc/en-us/articles/360001389366-Price-and-volume-decimal-precision
var krakenPrecisionMatrix = map[model.TradingPair]model.OrderConstraints{
	*model.MakeTradingPair(model.XLM, model.USD): *model.MakeOrderConstraints(6, 8, 30.0),
	*model.MakeTradingPair(model.XLM, model.BTC): *model.MakeOrderConstraints(8, 8, 30.0),
	*model.MakeTradingPair(model.BTC, model.USD): *model.MakeOrderConstraints(1, 8, 0.002),
	*model.MakeTradingPair(model.ETH, model.USD): *model.MakeOrderConstraints(2, 8, 0.02),
	*model.MakeTradingPair(model.ETH, model.BTC): *model.MakeOrderConstraints(5, 8, 0.02),
	*model.MakeTradingPair(model.XRP, model.USD): *model.MakeOrderConstraints(5, 8, 30.0),
	*model.MakeTradingPair(model.XRP, model.BTC): *model.MakeOrderConstraints(8, 8, 30.0),
}
