package kraken

import (
	"github.com/lightyeario/kelp/model"
	"github.com/stellar/go/support/log"
)

// CancelOrder impl.
func (k krakenExchange) CancelOrder(txID *model.TransactionID) (model.CancelOrderResult, error) {
	resp, e := k.api.CancelOrder(txID.String())
	if e != nil {
		return model.CancelResultFailed, e
	}

	if resp.Count > 1 {
		log.Info("warning: count from a cancelled order is greater than 1", resp.Count)
	}

	// TODO 2 - need to figure out whether count = 0 could also mean that it is pending cancellation
	if resp.Count == 0 {
		return model.CancelResultFailed, nil
	}
	// resp.Count == 1 here

	if resp.Pending {
		return model.CancelResultPending, nil
	}
	return model.CancelResultCancelSuccessful, nil
}
