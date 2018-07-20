package kraken

import (
	"fmt"

	"github.com/lightyeario/kelp/model"
)

// GetOpenOrders impl.
func (k krakenExchange) GetOpenOrders() (map[model.TradingPair][]model.OpenOrder, error) {
	openOrdersResponse, e := k.api.OpenOrders(map[string]string{})
	if e != nil {
		return nil, e
	}

	m := map[model.TradingPair][]model.OpenOrder{}
	for ID, o := range openOrdersResponse.Open {
		// for some reason the open orders API returns the normal codes for assets
		pair, e := model.TradingPairFromString(3, model.Display, o.Description.AssetPair)
		if e != nil {
			return nil, e
		}
		if _, ok := m[*pair]; !ok {
			m[*pair] = []model.OpenOrder{}
		}
		if _, ok := m[model.TradingPair{Base: pair.Quote, Quote: pair.Base}]; ok {
			return nil, fmt.Errorf("open orders are listed with repeated base/quote pairs for %s", *pair)
		}

		m[*pair] = append(m[*pair], model.OpenOrder{
			Order: model.Order{
				Pair:        pair,
				OrderAction: model.OrderActionFromString(o.Description.Type),
				OrderType:   model.OrderTypeFromString(o.Description.OrderType),
				Price:       model.MustFromString(o.Description.PrimaryPrice, k.precision),
				Volume:      model.MustFromString(o.Volume, k.precision),
				Timestamp:   model.MakeTimestamp(int64(o.OpenTime)),
			},
			ID:             ID,
			StartTime:      model.MakeTimestamp(int64(o.StartTime)),
			ExpireTime:     model.MakeTimestamp(int64(o.ExpireTime)),
			VolumeExecuted: model.FromFloat(o.VolumeExecuted, k.precision),
		})
	}
	return m, nil
}
