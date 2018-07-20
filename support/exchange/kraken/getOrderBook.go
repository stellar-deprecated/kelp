package kraken

import (
	"github.com/Beldur/kraken-go-api-client"

	"github.com/lightyeario/kelp/model"
)

// GetOrderBook impl.
func (k krakenExchange) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {
	pairStr, e := pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	krakenob, e := k.api.Depth(pairStr, int(maxCount))
	if e != nil {
		return nil, e
	}

	asks := k.readOrders(krakenob.Asks, pair, model.OrderActionSell)
	bids := k.readOrders(krakenob.Bids, pair, model.OrderActionBuy)
	ob := model.MakeOrderBook(pair, asks, bids)
	return ob, nil
}

func (k krakenExchange) readOrders(obi []krakenapi.OrderBookItem, pair *model.TradingPair, orderAction model.OrderAction) []model.Order {
	orders := []model.Order{}
	for _, item := range obi {
		orders = append(orders, model.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   model.OrderTypeLimit,
			Price:       model.FromFloat(item.Price, k.precision),
			Volume:      model.FromFloat(item.Amount, k.precision),
			Timestamp:   model.MakeTimestamp(item.Ts),
		})
	}
	return orders
}
