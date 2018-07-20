package kraken

import (
	"github.com/Beldur/kraken-go-api-client"

	"github.com/lightyeario/kelp/model"
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

	asks := k.readOrders(krakenob.Asks, pair, orderbook.OrderActionSell)
	bids := k.readOrders(krakenob.Bids, pair, orderbook.OrderActionBuy)
	ob := orderbook.MakeOrderBook(pair, asks, bids)
	return ob, nil
}

func (k krakenExchange) readOrders(obi []krakenapi.OrderBookItem, pair *model.TradingPair, orderAction orderbook.OrderAction) []orderbook.Order {
	orders := []orderbook.Order{}
	for _, item := range obi {
		orders = append(orders, orderbook.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   orderbook.OrderTypeLimit,
			Price:       model.FromFloat(item.Price, k.precision),
			Volume:      model.FromFloat(item.Amount, k.precision),
			Timestamp:   model.MakeTimestamp(item.Ts),
		})
	}
	return orders
}
