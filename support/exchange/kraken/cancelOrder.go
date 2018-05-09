package kraken

import (
	"github.com/lightyeario/kelp/support/exchange/api/orderbook"
	"github.com/lightyeario/kelp/support/exchange/api/trades"
	"github.com/stellar/go/support/log"
)

// CancelOrder impl.
func (k krakenExchange) CancelOrder(txID *orderbook.TransactionID) (trades.CancelOrderResult, error) {
	resp, e := k.api.CancelOrder(txID.String())
	if e != nil {
		return trades.CancelResultFailed, e
	}

	if resp.Count > 1 {
		log.Info("warning: count from a cancelled order is greater than 1", resp.Count)
	}

	// TODO 2 - need to figure out whether count = 0 could also mean that it is pending cancellation
	if resp.Count == 0 {
		return trades.CancelResultFailed, nil
	}
	// resp.Count == 1 here

	if resp.Pending {
		return trades.CancelResultPending, nil
	}
	return trades.CancelResultCancelSuccessful, nil
}
