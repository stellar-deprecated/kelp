package kraken

import (
	"github.com/Beldur/kraken-go-api-client"

	"github.com/lightyeario/kelp/model/assets"
	"github.com/lightyeario/kelp/support/exchange/api/dates"
	"github.com/lightyeario/kelp/support/exchange/api/number"
	"github.com/lightyeario/kelp/support/exchange/api/orderbook"
)

// GetOrderBook impl.
func (k krakenExchange) GetOrderBook(pair *model.TradingPair, maxCount int32) (*orderbook.OrderBook, error) {
	pairStr, e := pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	krakenob, e := k.api.Depth(pairStr, int(maxCount))
	if e != nil {
		return nil, e
	}

	asks := k.readOrders(krakenob.Asks, pair, orderbook.ActionSell)
	bids := k.readOrders(krakenob.Bids, pair, orderbook.ActionBuy)
	ob := orderbook.MakeOrderBook(pair, asks, bids)
	return ob, nil
}

func (k krakenExchange) readOrders(obi []krakenapi.OrderBookItem, pair *model.TradingPair, orderAction orderbook.OrderAction) []orderbook.Order {
	orders := []orderbook.Order{}
	for _, item := range obi {
		orders = append(orders, orderbook.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   orderbook.TypeLimit,
			Price:       number.FromFloat(item.Price, k.precision),
			Volume:      number.FromFloat(item.Amount, k.precision),
			Timestamp:   dates.MakeTimestamp(item.Ts),
		})
	}
	return orders
}
