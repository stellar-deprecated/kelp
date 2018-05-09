package kraken

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/api/orderbook"
)

const pricePrecision = 5
const volPrecision = 5

// AddOrder impl.
func (k krakenExchange) AddOrder(order *orderbook.Order) (*orderbook.TransactionID, error) {
	pairStr, e := order.Pair.ToString(k.assetConverter, k.delimiter)
	if e != nil {
		return nil, e
	}
	if order.Price.Precision() != pricePrecision {
		return nil, fmt.Errorf("price has unexpected precision: %d, expected %d", order.Price.Precision(), pricePrecision)
	}
	if order.Volume.Precision() != volPrecision {
		return nil, fmt.Errorf("volume has unexpected precision: %d, expected %d", order.Volume.Precision(), volPrecision)
	}

	args := map[string]string{
		"price": order.Price.AsString(),
	}
	// validate should not be present if it's false, otherwise Kraken treats it as true
	if k.isSimulated {
		args["validate"] = "true"
	}
	resp, e := k.api.AddOrder(
		pairStr,
		order.OrderAction.String(),
		order.OrderType.String(),
		order.Volume.AsString(),
		args,
	)
	if e != nil {
		return nil, e
	}

	// expected case for production orders
	if len(resp.TransactionIds) == 1 {
		return orderbook.MakeTransactionID(resp.TransactionIds[0]), nil
	}

	if len(resp.TransactionIds) > 1 {
		return nil, fmt.Errorf("there was more than 1 transctionId: %s", resp.TransactionIds)
	}

	if k.isSimulated {
		return nil, nil
	}
	return nil, fmt.Errorf("no transactionIds returned from order creation")
}
