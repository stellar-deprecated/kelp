package kraken

import (
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/orderbook"
)

// save this map for now to avoid if branches in GetOpenOrders below
var orderTypeMap = map[string]orderbook.OrderType{
	"buy":  orderbook.TypeBid,
	"sell": orderbook.TypeAsk,
}

// GetOpenOrders impl.
func (k krakenExchange) GetOpenOrders() (map[assets.TradingPair][]orderbook.OpenOrder, error) {
	openOrdersResponse, e := k.api.OpenOrders(map[string]string{})
	if e != nil {
		return nil, e
	}

	// TODO 2 - not sure if the trading pair is ordered correctly with the orderTypeMap above for buy/sell
	m := map[assets.TradingPair][]orderbook.OpenOrder{}
	for _, o := range openOrdersResponse.Open {
		pair, e := k.parsePair(o.Description.AssetPair)
		if e != nil {
			return nil, e
		}
		if _, ok := m[*pair]; !ok {
			m[*pair] = []orderbook.OpenOrder{}
		}

		m[*pair] = append(m[*pair], orderbook.OpenOrder{
			Order: orderbook.Order{
				OrderType: orderTypeMap[o.Description.Type],
				Price:     number.FromFloat(o.Price),
				Volume:    number.MustFromString(o.Volume),
				Timestamp: dates.MakeTimestamp(int64(o.OpenTime)),
			},
			ID:             o.ReferenceID, // TODO 2 - is this correct, or should it be o.UserRef?
			StartTime:      dates.MakeTimestamp(int64(o.StartTime)),
			ExpireTime:     dates.MakeTimestamp(int64(o.ExpireTime)),
			VolumeExecuted: number.FromFloat(o.VolumeExecuted),
		})
	}
	return m, nil
}
