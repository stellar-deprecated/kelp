package plugins

import (
	"fmt"
	"log"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

// SubmitFilter allows you to filter out operations before submitting to the network
type SubmitFilter interface {
	Apply(
		ops []txnbuild.Operation,
		sellingOffers []hProtocol.Offer, // quoted quote/base
		buyingOffers []hProtocol.Offer, // quoted base/quote
	) ([]txnbuild.Operation, error)
}

type filterFn func(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, bool, error)

func filterOps(ops []txnbuild.Operation, fn filterFn) ([]txnbuild.Operation, error) {
	numKeep := 0
	numDropped := 0
	numTransformed := 0
	filteredOps := []txnbuild.Operation{}
	for _, op := range ops {
		var newOp txnbuild.Operation
		var keep bool
		switch o := op.(type) {
		case *txnbuild.ManageSellOffer:
			var e error
			newOp, keep, e = fn(o)
			if e != nil {
				return nil, fmt.Errorf("could not transform offer (pointer case): %s", e)
			}
		default:
			newOp = o
			keep = true
		}

		isNewOpNil := newOp == nil || fmt.Sprintf("%v", newOp) == "<nil>"
		if keep {
			if isNewOpNil {
				return nil, fmt.Errorf("we want to keep op but newOp was nil (programmer error?)")
			}
			filteredOps = append(filteredOps, newOp)
			numKeep++
		} else {
			if !isNewOpNil {
				// newOp can be a transformed op to change the op to an effectively "dropped" state
				filteredOps = append(filteredOps, newOp)
				numTransformed++
			} else {
				numDropped++
			}
		}
	}

	log.Printf("filter result: dropped %d, transformed %d, kept %d ops from original %d ops, len(filteredOps) = %d\n", numDropped, numTransformed, numKeep, len(ops), len(filteredOps))
	return filteredOps, nil
}
