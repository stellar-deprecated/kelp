package plugins

import (
	"fmt"
	"log"
	"strconv"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

type orderConstraintsFilter struct {
	oc         *model.OrderConstraints
	baseAsset  hProtocol.Asset
	quoteAsset hProtocol.Asset
}

var _ SubmitFilter = &orderConstraintsFilter{}

// MakeFilterOrderConstraints makes a submit filter based on the passed in orderConstraints
func MakeFilterOrderConstraints(
	oc *model.OrderConstraints,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
) SubmitFilter {
	return &orderConstraintsFilter{
		oc:         oc,
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
	}
}

// Apply impl.
func (f *orderConstraintsFilter) Apply(
	ops []txnbuild.Operation,
	sellingOffers []hProtocol.Offer,
	buyingOffers []hProtocol.Offer,
) ([]txnbuild.Operation, error) {
	numKeep := 0
	numDropped := 0
	filteredOps := []txnbuild.Operation{}

	for _, op := range ops {
		var keep bool
		var e error
		var opPtr *txnbuild.ManageSellOffer

		switch o := op.(type) {
		case *txnbuild.ManageSellOffer:
			keep, e = f.shouldKeepOffer(o)
			if e != nil {
				return nil, fmt.Errorf("could not transform offer (pointer case): %s", e)
			}
			opPtr = o
		default:
			keep = true
		}

		if keep {
			filteredOps = append(filteredOps, opPtr)
			numKeep++
		} else {
			numDropped++
			// figure out how to convert the offer to a dropped state
			if opPtr.OfferID == 0 {
				// new offers can be dropped, so don't add to filteredOps
			} else if opPtr.Amount != "0" {
				// modify offers should be converted to delete offers
				opCopy := *opPtr
				opCopy.Amount = "0"
				filteredOps = append(filteredOps, &opCopy)
			} else {
				return nil, fmt.Errorf("unable to drop manageOffer operation (probably a delete op that should not have reached here): offerID=%d, amountRaw=%s", opPtr.OfferID, opPtr.Amount)
			}
		}
	}

	log.Printf("orderConstraintsFilter: dropped %d, kept %d ops from original %d ops, len(filteredOps) = %d\n", numDropped, numKeep, len(ops), len(filteredOps))
	return filteredOps, nil
}

func (f *orderConstraintsFilter) shouldKeepOffer(op *txnbuild.ManageSellOffer) (bool, error) {
	// delete operations should never be dropped
	amountFloat, e := strconv.ParseFloat(op.Amount, 64)
	if e != nil {
		return false, fmt.Errorf("could not convert amount (%s) to float: %s", op.Amount, e)
	}
	if op.Amount == "0" || amountFloat == 0.0 {
		log.Printf("orderConstraintsFilter: keeping delete operation with amount = %s\n", op.Amount)
		return true, nil
	}

	isSell, e := utils.IsSelling(f.baseAsset, f.quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return false, fmt.Errorf("error when running the isSelling check for offer '%+v': %s", *op, e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return false, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	if isSell {
		baseAmount := amountFloat
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
	quoteAmount := amountFloat
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
