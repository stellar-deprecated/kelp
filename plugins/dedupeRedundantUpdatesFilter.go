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
			return nil, false, nil
		}
		return op, true, nil
	}
	ops, e := filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, innerFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}

	return ops, nil
}
