package kraken

import (
	"errors"
	"strconv"

	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/trades"
)

// GetTradeHistory impl.
func (k krakenExchange) GetTradeHistory(maybeCursorStart interface{}, maybeCursorEnd interface{}) (*exchange.TradeHistoryResult, error) {
	var mcs *int64
	if maybeCursorStart != nil {
		i := maybeCursorStart.(int64)
		mcs = &i
	}

	var mce *int64
	if maybeCursorEnd != nil {
		i := maybeCursorEnd.(int64)
		mce = &i
	}

	return k.getTradeHistory(mcs, mce)
}

func (k krakenExchange) getTradeHistory(maybeCursorStart *int64, maybeCursorEnd *int64) (*exchange.TradeHistoryResult, error) {
	input := map[string]string{}
	if maybeCursorStart != nil {
		input["start"] = strconv.FormatInt(*maybeCursorStart, 10)
	}
	if maybeCursorEnd != nil {
		input["end"] = strconv.FormatInt(*maybeCursorEnd, 10)
	}

	resp, e := k.api.Query("TradesHistory", input)
	if e != nil {
		return nil, e
	}
	krakenResp := resp.(map[string]interface{})
	krakenTrades := krakenResp["trades"].(map[string]interface{})

	res := exchange.TradeHistoryResult{Trades: []trades.Trade{}}
	for _, v := range krakenTrades {
		m := v.(map[string]interface{})
		_txid := m["ordertxid"].(string)
		_time := m["time"].(float64)
		ts := dates.MakeTimestamp(int64(_time))
		_type := m["type"].(string)
		_ordertype := m["ordertype"].(string)
		tradeType, e := getTradeTypeFromStrings(_type, _ordertype)
		if e != nil {
			return nil, e
		}
		_price := m["price"].(string)
		_vol := m["vol"].(string)
		_cost := m["cost"].(string)
		_fee := m["fee"].(string)
		_pair := m["pair"].(string)
		pair, e := assets.FromString(k.assetConverter, _pair)
		if e != nil {
			return nil, e
		}

		res.Trades = append(res.Trades, trades.Trade{
			TransactionID: &_txid,
			Timestamp:     ts,
			Pair:          pair,
			Type:          tradeType,
			Price:         number.MustFromString(_price),
			Volume:        number.MustFromString(_vol),
			Cost:          number.MustFromString(_cost),
			Fee:           number.MustFromString(_fee),
		})
	}
	return &res, nil
}

func getTradeTypeFromStrings(_type string, _ordertype string) (*trades.TradeType, error) {
	var tradeType *trades.TradeType
	if _type == "buy" {
		if _ordertype == "market" {
			tradeType = trades.BuyMarket
		} else if _ordertype == "limit" {
			tradeType = trades.BuyLimit
		} else {
			return nil, errors.New("unidentified buy trade type")
		}
	} else if _type == "sell" {
		if _ordertype == "market" {
			tradeType = trades.SellMarket
		} else if _ordertype == "limit" {
			tradeType = trades.SellLimit
		} else {
			return nil, errors.New("unidentified sell trade type")
		}
	} else {
		return nil, errors.New("unidentified trade type")
	}
	return tradeType, nil
}
