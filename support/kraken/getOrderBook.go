package kraken

import (
	"github.com/Beldur/kraken-go-api-client"

	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/orderbook"
)

// GetOrderBook impl.
func (k krakenExchange) GetOrderBook(pair assets.TradingPair, maxCount int32) (*orderbook.OrderBook, error) {
	pairStr, e := pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}
	krakenob, e := k.api.Depth(pairStr, int(maxCount))
	if e != nil {
		return nil, e
	}

	asks := readOrders(krakenob.Asks, orderbook.TypeAsk)
	bids := readOrders(krakenob.Bids, orderbook.TypeBid)
	ob := orderbook.MakeOrderBook(asks, bids)
	return ob, nil
}

func readOrders(obi []krakenapi.OrderBookItem, orderType orderbook.OrderType) []orderbook.Order {
	orders := []orderbook.Order{}
	for _, item := range obi {
		orders = append(orders, orderbook.Order{
			OrderType: orderType,
			Price:     number.FromFloat(item.Price),
			Volume:    number.FromFloat(item.Amount),
			Timestamp: dates.MakeTimestamp(item.Ts),
		})
	}
	return orders
}
