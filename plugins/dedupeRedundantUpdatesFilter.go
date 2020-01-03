package plugins

import (
	"fmt"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

type dedupeRedundantUpdatesFilter struct {
	name       string
	baseAsset  hProtocol.Asset
	quoteAsset hProtocol.Asset
}

// MakeFilterDedupeRedundantUpdates makes a submit filter that dedupes offer updates that would result in the same offer being placed on the orderbook
func MakeFilterDedupeRedundantUpdates(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset) SubmitFilter {
	return &dedupeRedundantUpdatesFilter{
		name:       "dedupeRedundantUpdatesFilter",
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
	}
}

var _ SubmitFilter = &dedupeRedundantUpdatesFilter{}

func (f *dedupeRedundantUpdatesFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	offersMap := map[int64]hProtocol.Offer{}
	for _, offer := range append(sellingOffers, buyingOffers...) {
		offersMap[offer.ID] = offer
	}

	innerFn := func(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, bool, error) {
		existingOffer, hasOffer := offersMap[op.OfferID]
		isDupe := hasOffer && existingOffer.Amount == op.Amount && existingOffer.Price == op.Price

		if isDupe {
			// do not return an op because this is spawned from an offer (hasOffer = true) and we do not want to drop the offer nor do we want to create
			// a new operation to delete the offer and re-update the offer to the same values, so the returned Operation is nil with keep = false
			return nil, false, nil
		}

		// this an actual update of either the offer or a new operation, we dont want to make any changes to that so we return the original op with keep = true
		return op, true, nil
	}
	ops, e := filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, innerFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}

	return ops, nil
}
