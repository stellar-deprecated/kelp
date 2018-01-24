package kraken

import (
	"errors"

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
		tradeType, e := getTradeType(tInfo)
		if e != nil {
			return nil, e
		}
		tradesResult.Trades = append(tradesResult.Trades, trades.Trade{
			Type:      tradeType,
			Price:     number.FromFloat(tInfo.PriceFloat),
			Volume:    number.FromFloat(tInfo.VolumeFloat),
			Timestamp: dates.MakeTimestamp(tInfo.Time),
		})
	}
	return tradesResult, nil
}

func getTradeType(tInfo krakenapi.TradeInfo) (*trades.TradeType, error) {
	var tradeType *trades.TradeType
	if tInfo.Buy {
		if tInfo.Market {
			tradeType = trades.BuyMarket
		} else if tInfo.Limit {
			tradeType = trades.BuyLimit
		} else {
			return nil, errors.New("unidentified buy trade type")
		}
	} else if tInfo.Sell {
		if tInfo.Market {
			tradeType = trades.SellMarket
		} else if tInfo.Limit {
			tradeType = trades.SellLimit
		} else {
			return nil, errors.New("unidentified sell trade type")
		}
	} else {
		return nil, errors.New("unidentified trade type")
	}
	return tradeType, nil
}
