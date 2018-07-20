package kraken

import (
	"errors"

	"github.com/Beldur/kraken-go-api-client"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/exchange/api"
	"github.com/lightyeario/kelp/support/exchange/api/trades"
)

// GetTrades impl.
func (k krakenExchange) GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*api.TradesResult, error) {
	if maybeCursor != nil {
		mc := maybeCursor.(int64)
		return k.getTrades(pair, &mc)
	}
	return k.getTrades(pair, nil)
}

func (k krakenExchange) getTrades(pair *model.TradingPair, maybeCursor *int64) (*api.TradesResult, error) {
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

	tradesResult := &api.TradesResult{
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
			Order: model.Order{
				Pair:        pair,
				OrderAction: action,
				OrderType:   orderType,
				Price:       model.FromFloat(tInfo.PriceFloat, k.precision),
				Volume:      model.FromFloat(tInfo.VolumeFloat, k.precision),
				Timestamp:   model.MakeTimestamp(tInfo.Time),
			},
			// TransactionID unavailable
			// Cost unavailable
			// Fee unavailable
		})
	}
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
