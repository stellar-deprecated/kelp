package kraken

import (
	"errors"

	"github.com/lightyeario/kelp/support/exchange/orderbook"

	"github.com/Beldur/kraken-go-api-client"

	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/trades"
)

// GetTrades impl.
func (k krakenExchange) GetTrades(pair *assets.TradingPair, maybeCursor interface{}) (*exchange.TradesResult, error) {
	if maybeCursor != nil {
		mc := maybeCursor.(int64)
		return k.getTrades(pair, &mc)
	}
	return k.getTrades(pair, nil)
}

func (k krakenExchange) getTrades(pair *assets.TradingPair, maybeCursor *int64) (*exchange.TradesResult, error) {
	pairStr, e := pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	var tradesResp *krakenapi.TradesResponse
	if maybeCursor != nil {
		tradesResp, e = k.api.Trades(pairStr, *maybeCursor)
	} else {
		tradesResp, e = k.api.Trades(pairStr, -1)
	}
	if e != nil {
		return nil, e
	}

	tradesResult := &exchange.TradesResult{
		Cursor: tradesResp.Last,
		Trades: []trades.Trade{},
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

		tradesResult.Trades = append(tradesResult.Trades, trades.Trade{
			Order: orderbook.Order{
				Pair:        pair,
				OrderAction: action,
				OrderType:   orderType,
				Price:       number.FromFloat(tInfo.PriceFloat),
				Volume:      number.FromFloat(tInfo.VolumeFloat),
				Timestamp:   dates.MakeTimestamp(tInfo.Time),
			},
			// TransactionID unavailable
			// Cost unavailable
			// Fee unavailable
		})
	}
	return tradesResult, nil
}

func getAction(tInfo krakenapi.TradeInfo) (orderbook.OrderAction, error) {
	if tInfo.Buy {
		return orderbook.ActionBuy, nil
	} else if tInfo.Sell {
		return orderbook.ActionSell, nil
	}

	// return ActionBuy as nil value
	return orderbook.ActionBuy, errors.New("unidentified trade action")
}

func getOrderType(tInfo krakenapi.TradeInfo) (orderbook.OrderType, error) {
	if tInfo.Market {
		return orderbook.TypeMarket, nil
	} else if tInfo.Limit {
		return orderbook.TypeLimit, nil
	}
	return -1, errors.New("unidentified trade action")
}
