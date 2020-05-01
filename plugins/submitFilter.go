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

// filterFn returns a non-nil op to indicate the op that we want to append to the update. the newOp can do one of the following:
//     - modify an existing offer
//     - create a new offer
//     - update an offer that was created by a previous filterFn
// filterFn has no knowledge of whether the passed in op is an existing offer or a new op and therefore is not responsible for
// deleting existing offers.
// If the newOp returned is nil and it was spawned from an existingOffer then the filterOps function here will automatically delete
// the existing offer. i.e. if filterFn returns a nil newOp value then we will "drop" that operation or delete the existing offer.
type filterFn func(op *txnbuild.ManageSellOffer) (newOp *txnbuild.ManageSellOffer, e error)

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

func makeOfferMap(offers []hProtocol.Offer) map[int64]hProtocol.Offer {
	offerMap := map[int64]hProtocol.Offer{}
	for _, o := range offers {
		offerMap[o.ID] = o
	}
	return offerMap
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
	offerMap := makeOfferMap(append(sellingOffers, buyingOffers...))
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

			opToTransform, filterCounterToIncrement, isIgnoredOffer, e := selectOpOrOffer(
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
				// don't increment anything here becuase it will be addressed with the op that updated the offer
				continue
			}
			originalOfferAsOp := fetchOfferAsOpByID(opToTransform.OfferID, offerMap)

			newOpToPrepend, newOpToAppend, incrementValues, e := runInnerFilterFn(
				*opToTransform, // pass copy
				fn,
				originalOfferAsOp,
				*o, // pass copy
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
			if originalOfferAsOp != nil {
				offerCounter.add(incrementValues)

				// if this was a selection of an operation that had a corresponding offer than increment opCounter's ignored field
				if *filterCounterToIncrement == opCounter {
					opCounter.ignored++
				}
			} else {
				opCounter.add(incrementValues)
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
		ignoreOfferIds,
		filteredOps,
		fn,
	)
	if e != nil {
		return nil, fmt.Errorf("error when handling remaining buy offers: %s", e)
	}

	log.Printf("filter \"%s\" result A: dropped %d, transformed %d, kept %d, ignored %d (handled by offer counter) ops from the %d ops passed in\n", filterName, opCounter.dropped, opCounter.transformed, opCounter.kept, opCounter.ignored, len(ops))
	log.Printf("filter \"%s\" result B: dropped %d, transformed %d, kept %d from original %d sell offers\n", filterName, sellCounter.dropped, sellCounter.transformed, sellCounter.kept, len(sellingOffers))
	log.Printf("filter \"%s\" result C: dropped %d, transformed %d, kept %d from original %d buy offers\n", filterName, buyCounter.dropped, buyCounter.transformed, buyCounter.kept, len(buyingOffers))
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
	c *filterCounter,
	isIgnoredOffer bool,
	err error,
) {
	if offerCounter.idx >= len(offerList) {
		return mso, opCounter, false, nil
	}

	nextOffer := offerList[offerCounter.idx]
	if _, ignoreOffer := ignoreOfferIds[nextOffer.ID]; ignoreOffer {
		// we want to only compare against valid offers so ignore this offer
		return nil, offerCounter, true, nil
	}

	offerPrice := float64(nextOffer.PriceR.N) / float64(nextOffer.PriceR.D)
	opPrice, e := strconv.ParseFloat(mso.Price, 64)
	if e != nil {
		return nil, nil, false, fmt.Errorf("could not parse price as float64: %s", e)
	}

	// use the existing offer if the price is the same so we don't recreate an offer unnecessarily
	offerAsOp := convertOffer2MSO(nextOffer)
	if opPrice < offerPrice {
		return mso, opCounter, false, nil
	}
	return offerAsOp, offerCounter, false, nil
}

// fetchOfferAsOpByID returns the offer as an op if it exists otherwise nil
func fetchOfferAsOpByID(offerID int64, offerMap map[int64]hProtocol.Offer) *txnbuild.ManageSellOffer {
	if offer, exists := offerMap[offerID]; exists {
		return convertOffer2MSO(offer)
	}
	return nil
}

func runInnerFilterFn(
	opToTransform txnbuild.ManageSellOffer, // passed by value so it doesn't change
	fn filterFn,
	originalOfferAsOp *txnbuild.ManageSellOffer,
	originalMSO txnbuild.ManageSellOffer, // passed by value so it doesn't change
) (
	newOpToPrepend *txnbuild.ManageSellOffer,
	newOpToAppend *txnbuild.ManageSellOffer,
	incrementValues filterCounter,
	err error,
) {
	var newOp *txnbuild.ManageSellOffer
	var e error

	// delete operations should never be dropped
	if opToTransform.Amount == "0" {
		newOp = &opToTransform
	} else {
		newOp, e = fn(&opToTransform)
		if e != nil {
			return nil, nil, filterCounter{}, fmt.Errorf("error in inner filter fn: %s", e)
		}
	}

	keep := newOp != nil && fmt.Sprintf("%v", newOp) != "<nil>"
	if keep {
		if originalOfferAsOp != nil && originalOfferAsOp.Price == newOp.Price && originalOfferAsOp.Amount == newOp.Amount {
			// do not append to filteredOps because this is an existing offer that we want to keep as-is
			return nil, nil, filterCounter{kept: 1}, nil
		} else if originalOfferAsOp != nil {
			// we were dealing with an existing offer that was modified
			return nil, newOp, filterCounter{transformed: 1}, nil
		} else {
			// we were dealing with an operation
			opModified := originalMSO.Price != newOp.Price || originalMSO.Amount != newOp.Amount
			if opModified {
				return nil, newOp, filterCounter{transformed: 1}, nil
			}
			return nil, newOp, filterCounter{kept: 1}, nil
		}
	} else {
		if originalOfferAsOp != nil {
			// if newOp is nil for an original offer it means we want to explicitly delete that offer
			opCopy := *originalOfferAsOp
			opCopy.Amount = "0"
			return &opCopy, nil, filterCounter{dropped: 1}, nil
		} else {
			// if newOp is nil and it is not an original offer it means we want to drop the operation.
			return nil, nil, filterCounter{dropped: 1}, nil
		}
	}
}

func handleRemainingOffers(
	offerCounter *filterCounter,
	offers []hProtocol.Offer,
	ignoreOfferIds map[int64]bool,
	filteredOps []txnbuild.Operation,
	fn filterFn,
) ([]txnbuild.Operation, error) {
	for offerCounter.idx < len(offers) {
		if _, ignoreOffer := ignoreOfferIds[offers[offerCounter.idx].ID]; ignoreOffer {
			// don't increment anything here becuase it was already addressed with the op that updated the offer
			// so just move on to the next one
			offerCounter.idx++
			continue
		}

		originalOfferAsOp := convertOffer2MSO(offers[offerCounter.idx])
		newOpToPrepend, newOpToAppend, incrementValues, e := runInnerFilterFn(
			*originalOfferAsOp, // pass copy
			fn,
			originalOfferAsOp,
			*originalOfferAsOp, // pass copy
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
		offerCounter.add(incrementValues)
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
