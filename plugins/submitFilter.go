package plugins

import (
	"fmt"
	"log"
	"strconv"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/support/utils"
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

type filterCounter struct {
	idx         int
	kept        uint8
	dropped     uint8
	transformed uint8
	ignored     uint8
}

func (f *filterCounter) add(other filterCounter) {
	f.idx += other.idx
	f.kept += other.kept
	f.dropped += other.dropped
	f.transformed += other.transformed
	f.ignored += other.ignored
}

// build a list of the existing offers that have a corresponding operation so we ignore these offers and only consider the operation version
func ignoreOfferIDs(ops []txnbuild.Operation) map[int64]bool {
	ignoreOfferIDs := map[int64]bool{}
	for _, op := range ops {
		switch o := op.(type) {
		case *txnbuild.ManageSellOffer:
			ignoreOfferIDs[o.OfferID] = true
		default:
			continue
		}
	}
	return ignoreOfferIDs
}

// TODO - simplify filterOps by separating out logic to convert into a single list of operations from transforming the operations
/*
What filterOps() does and why:

Solving the "existing offers problem":
Problem: We need to run the existing offers against the filter as well since they may no longer be compliant.
Solution: Do a merge of two "sorted" lists (operations list, offers list) to create a new list of operations.
	When sorted by price, this will ensure that we delete any spurious existing offers to meet the filter's
	needs. This also serves the purpose of "interleaving" the operations related to the offers and ops.

Solving the "ordering problem":
Problem: The incoming operations list combines both buy and sell operations. We want to run it though the filter
	without modifying the order of the buy or sell segments, or modify operations within the segments since that
	ordering is dictated by the strategy logic.
Solution: Since both these segments of buy/sell offers are contiguous, i.e. buy offers are all together and sell
	offers are all together, we can identify the "cutover point" in each list of operations and offers, and then
	advance the iteration index to the next segment for both segments in both lists by converting the remaining
	offers and operations to delete operations. This will not affect the order of operations, but any new delete
	operations created should be placed at the beginning of the respective buy and sell segments as is a requirement
	on sdex (see sellSideStrategy.go for details on why we need to start off with the delete operations).

Possible Question: Why do we not reuse the same logic that is in sellSideStrategy.go to "delete remaining offers"?
Answer: The logic that could possibly be reused is minimal -- it's just a for loop. The logic that converts offers
	to the associated delete operation is reused, which is the main crux of the "business logic" that we want to
	avoid rewriting. The logic in sellSideStrategy.go also only works on offers, here we work on offers and ops.

Solving the "increase price problem":
Problem: If we increase the price off a sell offer (or decrease price of a buy offer) then we will see the offer
	with an incorrect price before we see the update to the offer. This will result in an incorrect calculation,
	since we will later on see the updated offer and make adjustments, which would result in runtime complexity
	worse than O(N).
Solution: We first "dedupe" the offers and operations, by removing any offers that have a corresponding operation
	update based on offerID. This has an additional overhead on runtime complexity of O(N).

Solving the "no update operations problem":
Problem: if our trading strategy produces no operations for a given update cycle, indicating that the state of the
	orderbook is correct, then we will not enter the for-loop which is conditioned on operations. This would result
	in control going straight to the post-operations logic which should correctly consider the existing offers. This
	logic would be the same as what happens inside the for loop and we should ensure there is no repetition.
Solution: Refactor the code inside the for loop to clearly allow for reuse of functions and evaluation of existing
	offers outside the for loop.
*/
func filterOps(
	filterName string,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	sellingOffers []hProtocol.Offer,
	buyingOffers []hProtocol.Offer,
	ops []txnbuild.Operation,
	fn filterFn,
) ([]txnbuild.Operation, error) {
	ignoreOfferIds := ignoreOfferIDs(ops)
	opCounter := filterCounter{}
	buyCounter := filterCounter{}
	sellCounter := filterCounter{}
	filteredOps := []txnbuild.Operation{}

	for opCounter.idx < len(ops) {
		op := ops[opCounter.idx]

		switch o := op.(type) {
		case *txnbuild.ManageSellOffer:
			offerList, offerCounter, e := selectBuySellList(
				baseAsset,
				quoteAsset,
				o,
				sellingOffers,
				buyingOffers,
				&sellCounter,
				&buyCounter,
			)
			if e != nil {
				return nil, fmt.Errorf("unable to pick between whether the op was a buy or sell op: %s", e)
			}

			opToTransform, originalOffer, filterCounterToIncrement, isIgnoredOffer, e := selectOpOrOffer(
				offerList,
				offerCounter,
				o,
				&opCounter,
				ignoreOfferIds,
			)
			if e != nil {
				return nil, fmt.Errorf("error while picking op or offer: %s", e)
			}
			filterCounterToIncrement.idx++
			if isIgnoredOffer {
				filterCounterToIncrement.ignored++
				continue
			}

			newOpToAppend, newOpToPrepend, filterCounterToIncrement, incrementValues, e := runInnerFilterFn(
				*opToTransform, // pass copy
				fn,
				originalOffer,
				*o, // pass copy
				offerCounter,
				&opCounter,
			)
			if e != nil {
				return nil, fmt.Errorf("error while running inner filter function: %s", e)
			}
			if newOpToAppend != nil {
				filteredOps = append(filteredOps, newOpToAppend)
			}
			if newOpToPrepend != nil {
				filteredOps = append([]txnbuild.Operation{newOpToPrepend}, filteredOps...)
			}
			if filterCounterToIncrement != nil && incrementValues != nil {
				filterCounterToIncrement.add(*incrementValues)
			}
		default:
			filteredOps = append(filteredOps, o)
			opCounter.kept++
			opCounter.idx++
		}
	}

	// convert all remaining buy and sell offers to delete offers
	filteredOps, e := handleRemainingOffers(
		&sellCounter,
		sellingOffers,
		&opCounter,
		ignoreOfferIds,
		filteredOps,
		fn,
	)
	if e != nil {
		return nil, fmt.Errorf("error when handling remaining sell offers: %s", e)
	}
	filteredOps, e = handleRemainingOffers(
		&buyCounter,
		buyingOffers,
		&opCounter,
		ignoreOfferIds,
		filteredOps,
		fn,
	)
	if e != nil {
		return nil, fmt.Errorf("error when handling remaining buy offers: %s", e)
	}

	log.Printf("filter \"%s\" result A: dropped %d, transformed %d, kept %d ops from the %d ops passed in\n", filterName, opCounter.dropped, opCounter.transformed, opCounter.kept, len(ops))
	log.Printf("filter \"%s\" result B: dropped %d, transformed %d, kept %d, ignored %d sell offers (corresponding op update) from original %d sell offers\n", filterName, sellCounter.dropped, sellCounter.transformed, sellCounter.kept, sellCounter.ignored, len(sellingOffers))
	log.Printf("filter \"%s\" result C: dropped %d, transformed %d, kept %d, ignored %d buy offers (corresponding op update) from original %d buy offers\n", filterName, buyCounter.dropped, buyCounter.transformed, buyCounter.kept, buyCounter.ignored, len(buyingOffers))
	log.Printf("filter \"%s\" result D: len(filteredOps) = %d\n", filterName, len(filteredOps))
	return filteredOps, nil
}

func selectBuySellList(
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	mso *txnbuild.ManageSellOffer,
	sellingOffers []hProtocol.Offer,
	buyingOffers []hProtocol.Offer,
	sellCounter *filterCounter,
	buyCounter *filterCounter,
) ([]hProtocol.Offer, *filterCounter, error) {
	isSellOp, e := utils.IsSelling(baseAsset, quoteAsset, mso.Selling, mso.Buying)
	if e != nil {
		return nil, nil, fmt.Errorf("could not check whether the ManageSellOffer was selling or buying: %s", e)
	}

	if isSellOp {
		return sellingOffers, sellCounter, nil
	}
	return buyingOffers, buyCounter, nil
}

func selectOpOrOffer(
	offerList []hProtocol.Offer,
	offerCounter *filterCounter,
	mso *txnbuild.ManageSellOffer,
	opCounter *filterCounter,
	ignoreOfferIds map[int64]bool,
) (
	opToTransform *txnbuild.ManageSellOffer,
	originalOfferAsOp *txnbuild.ManageSellOffer,
	c *filterCounter,
	isIgnoredOffer bool,
	err error,
) {
	if offerCounter.idx >= len(offerList) {
		return mso, nil, opCounter, false, nil
	}

	existingOffer := offerList[offerCounter.idx]
	if _, ignoreOffer := ignoreOfferIds[existingOffer.ID]; ignoreOffer {
		// we want to only compare against valid offers so skip this offer by returning ignored = true
		return nil, nil, offerCounter, true, nil
	}

	offerPrice := float64(existingOffer.PriceR.N) / float64(existingOffer.PriceR.D)
	opPrice, e := strconv.ParseFloat(mso.Price, 64)
	if e != nil {
		return nil, nil, nil, false, fmt.Errorf("could not parse price as float64: %s", e)
	}

	// use the existing offer if the price is the same so we don't recreate an offer unnecessarily
	if opPrice < offerPrice {
		return mso, nil, opCounter, false, nil
	}

	offerAsOp := convertOffer2MSO(existingOffer)
	offerAsOpCopy := *offerAsOp
	return offerAsOp, &offerAsOpCopy, offerCounter, false, nil
}

func runInnerFilterFn(
	opToTransform txnbuild.ManageSellOffer, // passed by value so it doesn't change
	fn filterFn,
	originalOfferAsOp *txnbuild.ManageSellOffer,
	originalMSO txnbuild.ManageSellOffer, // passed by value so it doesn't change
	offerCounter *filterCounter,
	opCounter *filterCounter,
) (
	newOpToAppend *txnbuild.ManageSellOffer,
	newOpToPrepend *txnbuild.ManageSellOffer,
	filterCounterToIncrement *filterCounter,
	incrementValues *filterCounter,
	err error,
) {
	var keep bool
	var newOp *txnbuild.ManageSellOffer
	var e error

	// delete operations should never be dropped
	if opToTransform.Amount == "0" {
		newOp, keep = &opToTransform, true
	} else {
		newOp, keep, e = fn(&opToTransform)
		if e != nil {
			return nil, nil, nil, nil, fmt.Errorf("error in inner filter fn: %s", e)
		}
	}

	isNewOpNil := newOp == nil || fmt.Sprintf("%v", newOp) == "<nil>"
	if keep && isNewOpNil {
		return nil, nil, nil, nil, fmt.Errorf("we want to keep op but newOp was nil (programmer error?)")
	} else if keep {
		if originalOfferAsOp != nil && originalOfferAsOp.Price == newOp.Price && originalOfferAsOp.Amount == newOp.Amount {
			// do not append to filteredOps because this is an existing offer that we want to keep as-is
			return nil, nil, offerCounter, &filterCounter{kept: 1}, nil
		} else if originalOfferAsOp != nil {
			// we were dealing with an existing offer that was modified
			return newOp, nil, offerCounter, &filterCounter{transformed: 1}, nil
		} else {
			// we were dealing with an operation
			opModified := originalMSO.Price != newOp.Price || originalMSO.Amount != newOp.Amount
			if opModified {
				return newOp, nil, opCounter, &filterCounter{transformed: 1}, nil
			}
			return newOp, nil, opCounter, &filterCounter{kept: 1}, nil
		}
	} else if isNewOpNil {
		// newOp will never be nil for an original offer since we will return the original non-nil offer
		return nil, nil, opCounter, &filterCounter{dropped: 1}, nil
	} else {
		// newOp can be a transformed op to change the op to an effectively "dropped" state
		// prepend this so we always have delete commands at the beginning of the operation list
		if originalOfferAsOp != nil {
			// we are dealing with an existing offer that needs dropping
			return nil, newOp, offerCounter, &filterCounter{dropped: 1}, nil
		} else {
			// we are dealing with an operation that had updated an offer which now needs dropping
			// using the transformed counter here because we are changing the actual intent of the operation
			// from an update existing offer to delete existing offer logic
			return nil, newOp, opCounter, &filterCounter{transformed: 1}, nil
		}
	}
}

func handleRemainingOffers(
	offerCounter *filterCounter,
	offers []hProtocol.Offer,
	opCounter *filterCounter,
	ignoreOfferIds map[int64]bool,
	filteredOps []txnbuild.Operation,
	fn filterFn,
) ([]txnbuild.Operation, error) {
	for offerCounter.idx < len(offers) {
		if _, ignoreOffer := ignoreOfferIds[offers[offerCounter.idx].ID]; ignoreOffer {
			// we want to only compare against valid offers so ignore this offer
			offerCounter.ignored++
			offerCounter.idx++
			continue
		}

		originalOfferAsOp := convertOffer2MSO(offers[offerCounter.idx])
		newOpToAppend, newOpToPrepend, filterCounterToIncrement, incrementValues, e := runInnerFilterFn(
			*originalOfferAsOp, // pass copy
			fn,
			originalOfferAsOp,
			*originalOfferAsOp, // pass copy
			offerCounter,
			opCounter,
		)
		if e != nil {
			return nil, fmt.Errorf("error while running inner filter function for remaining offers: %s", e)
		}
		if newOpToAppend != nil {
			filteredOps = append(filteredOps, newOpToAppend)
		}
		if newOpToPrepend != nil {
			filteredOps = append([]txnbuild.Operation{newOpToPrepend}, filteredOps...)
		}
		if filterCounterToIncrement != nil && incrementValues != nil {
			filterCounterToIncrement.add(*incrementValues)
		}
		offerCounter.idx++
	}
	return filteredOps, nil
}

func convertOffer2MSO(offer hProtocol.Offer) *txnbuild.ManageSellOffer {
	return &txnbuild.ManageSellOffer{
		Selling:       utils.Asset2Asset(offer.Selling),
		Buying:        utils.Asset2Asset(offer.Buying),
		Amount:        offer.Amount,
		Price:         offer.Price,
		OfferID:       offer.ID,
		SourceAccount: &txnbuild.SimpleAccount{AccountID: offer.Seller},
	}
}
