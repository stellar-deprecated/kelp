package plugins

import (
	"fmt"
	"log"
	"math"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

type orderConstraintsFilter struct {
	oc         *model.OrderConstraints
	baseAsset  horizon.Asset
	quoteAsset horizon.Asset
}

var _ SubmitFilter = &orderConstraintsFilter{}

// MakeOrderConstraintsFilter makes a submit filter based on the passed in orderConstraints
func MakeOrderConstraintsFilter(
	oc *model.OrderConstraints,
	baseAsset horizon.Asset,
	quoteAsset horizon.Asset,
) SubmitFilter {
	return &orderConstraintsFilter{
		oc:         oc,
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
	}
}

// Apply impl.
func (f *orderConstraintsFilter) Apply(
	ops []build.TransactionMutator,
	sellingOffers []horizon.Offer,
	buyingOffers []horizon.Offer,
) ([]build.TransactionMutator, error) {
	numKeep := 0
	numDropped := 0
	filteredOps := []build.TransactionMutator{}

	for _, op := range ops {
		var keep bool
		var e error
		var opPtr *build.ManageOfferBuilder

		switch o := op.(type) {
		case *build.ManageOfferBuilder:
			keep, e = f.shouldKeepOffer(o)
			if e != nil {
				return nil, fmt.Errorf("could not transform offer (pointer case): %s", e)
			}
			opPtr = o
		case build.ManageOfferBuilder:
			keep, e = f.shouldKeepOffer(&o)
			if e != nil {
				return nil, fmt.Errorf("could not check transform offer (non-pointer case): %s", e)
			}
			opPtr = &o
		default:
			keep = true
		}

		if keep {
			filteredOps = append(filteredOps, opPtr)
			numKeep++
		} else {
			numDropped++
			// figure out how to convert the offer to a dropped state
			if opPtr.MO.OfferId == 0 {
				// new offers can be dropped, so don't add to filteredOps
			} else if opPtr.MO.Amount != 0 {
				// modify offers should be converted to delete offers
				opCopy := *opPtr
				opCopy.MO.Amount = 0
				filteredOps = append(filteredOps, opCopy)
			} else {
				return nil, fmt.Errorf("unable to drop manageOffer operation (probably a delete op that should not have reached here): offerID=%d, amountRaw=%.8f", opPtr.MO.OfferId, float64(opPtr.MO.Amount))
			}
		}
	}

	log.Printf("dropped %d, kept %d ops in orderConstraintsFilter from original %d ops, len(filteredOps) = %d\n", numDropped, numKeep, len(ops), len(filteredOps))
	return filteredOps, nil
}

func (f *orderConstraintsFilter) shouldKeepOffer(op *build.ManageOfferBuilder) (bool, error) {
	// delete operations should never be dropped
	if op.MO.Amount == 0 {
		return true, nil
	}

	isSell, e := utils.IsSelling(f.baseAsset, f.quoteAsset, op.MO.Selling, op.MO.Buying)
	if e != nil {
		return false, fmt.Errorf("error when running the isSelling check: %s", e)
	}

	sellPrice := float64(op.MO.Price.N) / float64(op.MO.Price.D)
	if isSell {
		baseAmount := float64(op.MO.Amount) / math.Pow(10, 7)
		quoteAmount := baseAmount * sellPrice
		if baseAmount < f.oc.MinBaseVolume.AsFloat() {
			log.Printf("orderConstraintsFilter: selling, keep = (baseAmount) %.8f < %s (MinBaseVolume): keep = false\n", baseAmount, f.oc.MinBaseVolume.AsString())
			return false, nil
		}
		if f.oc.MinQuoteVolume != nil && quoteAmount < f.oc.MinQuoteVolume.AsFloat() {
			log.Printf("orderConstraintsFilter: selling, keep = (quoteAmount) %.8f < %s (MinQuoteVolume): keep = false\n", quoteAmount, f.oc.MinQuoteVolume.AsString())
			return false, nil
		}
		log.Printf("orderConstraintsFilter: selling, baseAmount=%.8f, quoteAmount=%.8f, keep = true\n", baseAmount, quoteAmount)
		return true, nil
	}

	// buying
	quoteAmount := float64(op.MO.Amount) / math.Pow(10, 7)
	baseAmount := quoteAmount * sellPrice
	if baseAmount < f.oc.MinBaseVolume.AsFloat() {
		log.Printf("orderConstraintsFilter:  buying, keep = (baseAmount) %.8f < %s (MinBaseVolume): keep = false\n", baseAmount, f.oc.MinBaseVolume.AsString())
		return false, nil
	}
	if f.oc.MinQuoteVolume != nil && quoteAmount < f.oc.MinQuoteVolume.AsFloat() {
		log.Printf("orderConstraintsFilter:  buying, keep = (quoteAmount) %.8f < %s (MinQuoteVolume): keep = false\n", quoteAmount, f.oc.MinQuoteVolume.AsString())
		return false, nil
	}
	log.Printf("orderConstraintsFilter:  buying, baseAmount=%.8f, quoteAmount=%.8f, keep = true\n", baseAmount, quoteAmount)
	return true, nil
}
