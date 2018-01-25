package kraken

import (
	"strconv"

	"github.com/lightyeario/kelp/support/exchange/orderbook"
)

// AddOrder impl.
func (k krakenExchange) AddOrder(order *orderbook.Order) (*orderbook.TransactionID, error) {
	pairStr, e := order.Pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}

	resp, e := k.api.AddOrder(
		pairStr,
		order.OrderAction.String(),
		order.OrderType.String(),
		order.Volume.AsString(),
		map[string]string{
			"price":    order.Price.AsString(),
			"validate": strconv.FormatBool(k.isSimulated),
		},
	)
	if e != nil {
		return nil, e
	}

	return orderbook.MakeTransactionID(resp.TransactionIds[0]), nil
}
