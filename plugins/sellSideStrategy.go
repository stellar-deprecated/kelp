package plugins

import (
	"fmt"
	"log"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

const actionSell = "sell"
const actionBuy = "buy "

// sellSideStrategy is a strategy to sell a specific currency on SDEX on a single side by reading prices from an exchange
type sellSideStrategy struct {
	sdex                *SDEX
	assetBase           *horizon.Asset
	assetQuote          *horizon.Asset
	levelsProvider      api.LevelProvider
	priceTolerance      float64
	amountTolerance     float64
	divideAmountByPrice bool
	action              string

	// uninitialized
	currentLevels []api.Level // levels for current iteration
	maxAssetBase  float64
	maxAssetQuote float64
}

// ensure it implements SideStrategy
var _ api.SideStrategy = &sellSideStrategy{}

// makeSellSideStrategy is a factory method for sellSideStrategy
func makeSellSideStrategy(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	levelsProvider api.LevelProvider,
	priceTolerance float64,
	amountTolerance float64,
	divideAmountByPrice bool,
) api.SideStrategy {
	action := actionSell
	if divideAmountByPrice {
		action = actionBuy
	}
	return &sellSideStrategy{
		sdex:                sdex,
		assetBase:           assetBase,
		assetQuote:          assetQuote,
		levelsProvider:      levelsProvider,
		priceTolerance:      priceTolerance,
		amountTolerance:     amountTolerance,
		divideAmountByPrice: divideAmountByPrice,
		action:              action,
	}
}

// PruneExistingOffers impl
func (s *sellSideStrategy) PruneExistingOffers(offers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer) {
	// figure out which offers we want to prune
	shouldPrune := computeOffersToPrune(offers, s.currentLevels)

	pruneOps := []build.TransactionMutator{}
	updatedOffers := []horizon.Offer{}
	for i, offer := range offers {
		isPruning := shouldPrune[i]
		if isPruning {
			pOp := s.sdex.DeleteOffer(offer)
			pruneOps = append(pruneOps, &pOp)
		} else {
			updatedOffers = append(updatedOffers, offer)
		}

		curAmount := utils.AmountStringAsFloat(offer.Amount)
		curPrice := utils.GetPrice(offer)
		if s.divideAmountByPrice {
			curAmount = curAmount * curPrice
			curPrice = 1 / curPrice
		}
		// base and quote here refers to the bot's base and quote, not the base and quote of the sellSideStrategy
		log.Printf("offer | %s | level=%d | curPriceQuote=%.7f | curAmtBase=%.7f | pruning=%v\n", s.action, i+1, curPrice, curAmount, isPruning)
	}
	return pruneOps, updatedOffers
}

// computeOffersToPrune returns a list of bools representing whether we should prune the offer at that position or not
func computeOffersToPrune(offers []horizon.Offer, levels []api.Level) []bool {
	numToPrune := len(offers) - len(levels)
	if numToPrune <= 0 {
		return make([]bool, len(offers))
	}

	offerIdx := 0
	levelIdx := 0
	shouldPrune := make([]bool, len(offers))
	for numToPrune > 0 {
		if offerIdx == len(offers) || levelIdx == len(levels) {
			// prune remaining offers (from the back as a convention)
			for i := 0; i < numToPrune; i++ {
				shouldPrune[len(offers)-1-i] = true
			}
			return shouldPrune
		}

		offerPrice := float64(offers[offerIdx].PriceR.N) / float64(offers[offerIdx].PriceR.D)
		levelPrice := levels[levelIdx].Price.AsFloat()
		if offerPrice < levelPrice {
			shouldPrune[offerIdx] = true
			numToPrune--
			offerIdx++
		} else if offerPrice == levelPrice {
			shouldPrune[offerIdx] = false
			offerIdx++
			// do not increment levelIdx because we could have two offers or levels at the same price. This will resolve in the next iteration automatically.
		} else {
			levelIdx++
		}
	}
	return shouldPrune
}

// PreUpdate impl
func (s *sellSideStrategy) PreUpdate(maxAssetBase float64, maxAssetQuote float64, trustBase float64, trustQuote float64) error {
	s.maxAssetBase = maxAssetBase
	s.maxAssetQuote = maxAssetQuote

	// don't place orders if we have nothing to sell or if we cannot buy the asset in exchange
	nothingToSell := maxAssetBase == 0
	lineFull := maxAssetQuote == trustQuote
	if nothingToSell || lineFull {
		s.currentLevels = []api.Level{}
		log.Printf("no capacity to place sell orders (nothingToSell = %v, lineFull = %v)\n", nothingToSell, lineFull)
		return nil
	}

	// load currentLevels only once here
	var e error
	s.currentLevels, e = s.levelsProvider.GetLevels(s.maxAssetBase, s.maxAssetQuote)
	if e != nil {
		log.Printf("levels couldn't be loaded: %s\n", e)
		return e
	}
	return nil
}

// computePrecedingLevels returns the levels priced better than the lowest existing offer, up to the max preceding levels allowed
func computePrecedingLevels(offers []horizon.Offer, levels []api.Level) []api.Level {
	if len(offers) == 0 {
		// we want to place all levels as create offers
		return levels
	}
	if len(offers) >= len(levels) {
		// we have enough offers to modify to reach our goal
		// our logic is not sophisticated so we default to the modify all offers behavior here
		return []api.Level{}
	}

	// the number of new levels we can place is capped by the number of existing offers we have
	maxPrecedingLevels := len(levels) - len(offers)
	// we only want to create new offers that are priced lower than the lowest existing offer
	cutoffPrice := float64(offers[0].PriceR.N) / float64(offers[0].PriceR.D)

	precedingLevels := []api.Level{}
	for i, level := range levels {
		if i >= maxPrecedingLevels {
			break
		}

		if level.Price.AsFloat() >= cutoffPrice {
			break
		}
		precedingLevels = append(precedingLevels, level)
	}
	return precedingLevels
}

func (s *sellSideStrategy) computeTargets(level api.Level) (targetPrice *model.Number, targetAmount *model.Number, e error) {
	targetPrice = &level.Price
	targetAmount = &level.Amount

	if targetPrice.AsFloat() == 0 {
		return nil, nil, fmt.Errorf("targetPrice is 0")
	}
	if targetAmount.AsFloat() == 0 {
		return nil, nil, fmt.Errorf("targetAmount is 0")
	}

	if s.divideAmountByPrice {
		targetAmount = model.NumberFromFloat(targetAmount.AsFloat()/targetPrice.AsFloat(), targetAmount.Precision())
	}

	return targetPrice, targetAmount, nil
}

func (s *sellSideStrategy) createPrecedingOffers(
	precedingLevels []api.Level,
) (
	int, // numLevelsConsumed
	bool, // hitCapacityLimit
	[]build.TransactionMutator, // ops
	*model.Number, // newTopOffer
	error, // e
) {
	hitCapacityLimit := false
	ops := []build.TransactionMutator{}
	var newTopOffer *model.Number

	for i := 0; i < len(precedingLevels); i++ {
		if hitCapacityLimit {
			// we consider the ith level consumed because we don't want to create an offer for it anyway since we hit the capacity limit
			return (i + 1), true, ops, newTopOffer, nil
		}

		targetPrice, targetAmount, e := s.computeTargets(precedingLevels[i])
		if e != nil {
			return 0, false, nil, nil, fmt.Errorf("could not compute targets: %s", e)
		}

		var offerPrice *model.Number
		var op *build.ManageOfferBuilder
		offerPrice, hitCapacityLimit, op, e = s.createSellLevel(i, *targetPrice, *targetAmount)
		if e != nil {
			return 0, false, nil, nil, fmt.Errorf("unable to create new preceding offer: %s", e)
		}

		if op != nil {
			ops = append(ops, op)
		}

		// update top offer, newTopOffer is minOffer because this is a sell strategy, and the lowest price is the best (top) price on the orderbook
		if newTopOffer == nil || offerPrice.AsFloat() < newTopOffer.AsFloat() {
			newTopOffer = offerPrice
		}
	}

	// hitCapacityLimit can be updated after the check inside the for loop
	return len(precedingLevels), hitCapacityLimit, ops, newTopOffer, nil
}

// UpdateWithOps impl
func (s *sellSideStrategy) UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error) {
	deleteOps := []build.TransactionMutator{}

	// first we want to re-create any offers that precede our existing offers and are additions to the existing offers that we have
	precedingLevels := computePrecedingLevels(offers, s.currentLevels)
	var hitCapacityLimit bool
	var numLevelsConsumed int
	numLevelsConsumed, hitCapacityLimit, ops, newTopOffer, e = s.createPrecedingOffers(precedingLevels)
	if e != nil {
		return nil, nil, fmt.Errorf("unable to create preceding offers: %s", e)
	}
	// pad the offers so it lines up correctly with numLevelsConsumed.
	// alternatively we could chop off the beginning of s.currentLevels but then that affects the logging of levels downstream
	for i := 0; i < numLevelsConsumed; i++ {
		offers = append([]horizon.Offer{horizon.Offer{}}, offers...)
	}

	// next we want to adjust our remaining offers to be in line with what is desired, creating new offers that may not exist at the end of our existing offers
	for i := numLevelsConsumed; i < len(s.currentLevels); i++ {
		isModify := i < len(offers)
		// we only want to delete offers after we hit the capacity limit which is why we perform this check in the beginning
		if hitCapacityLimit {
			if isModify {
				delOp := s.sdex.DeleteOffer(offers[i])
				log.Printf("deleting offer because we previously hit the capacity limit, offerId=%d\n", offers[i].ID)
				deleteOps = append(deleteOps, delOp)
				continue
			} else {
				// we can break because we would never see a modify operation happen after a non-modify operation
				break
			}
		}

		// hitCapacityLimit can be updated below
		targetPrice, targetAmount, e := s.computeTargets(s.currentLevels[i])
		if e != nil {
			return nil, nil, fmt.Errorf("could not compute targets: %s", e)
		}

		var offerPrice *model.Number
		var op *build.ManageOfferBuilder
		if isModify {
			offerPrice, hitCapacityLimit, op, e = s.modifySellLevel(offers, i, *targetPrice, *targetAmount)
		} else {
			offerPrice, hitCapacityLimit, op, e = s.createSellLevel(i, *targetPrice, *targetAmount)
		}
		if e != nil {
			return nil, nil, fmt.Errorf("unable to update existing offers or create new offers: %s", e)
		}
		if op != nil {
			reducedOrderSize := isModify && targetAmount.AsFloat() < utils.AmountStringAsFloat(offers[i].Amount)
			hitCapacityLimitModify := isModify && hitCapacityLimit
			if reducedOrderSize || hitCapacityLimitModify {
				// prepend operations that reduce the size of an existing order because they decrease our liabilities
				ops = append([]build.TransactionMutator{op}, ops...)
			} else {
				ops = append(ops, op)
			}
		}

		// update top offer, newTopOffer is minOffer because this is a sell strategy, and the lowest price is the best (top) price on the orderbook
		if newTopOffer == nil || offerPrice.AsFloat() < newTopOffer.AsFloat() {
			newTopOffer = offerPrice
		}
	}

	// prepend deleteOps because we want to delete offers first so we "free" up our liabilities capacity to place the new/modified offers
	ops = append(deleteOps, ops...)

	return ops, newTopOffer, nil
}

// PostUpdate impl
func (s *sellSideStrategy) PostUpdate() error {
	return nil
}

// computeRemainderAmount returns sellingAmount, buyingAmount, error
func (s *sellSideStrategy) computeRemainderAmount(incrementalSellAmount float64, incrementalBuyAmount float64, price float64, incrementalNativeAmountRaw float64) (float64, float64, error) {
	availableSellingCapacity, e := s.sdex.AvailableCapacity(*s.assetBase, incrementalNativeAmountRaw)
	if e != nil {
		return 0, 0, e
	}
	availableBuyingCapacity, e := s.sdex.AvailableCapacity(*s.assetQuote, incrementalNativeAmountRaw)
	if e != nil {
		return 0, 0, e
	}

	if availableSellingCapacity.Selling >= incrementalSellAmount && availableBuyingCapacity.Buying >= incrementalBuyAmount {
		return 0, 0, fmt.Errorf("error: (programmer?) unable to create offer but available capacities were more than the attempted offer amounts, sellingCapacity=%.7f, incrementalSellAmount=%.7f, buyingCapacity=%.7f, incrementalBuyAmount=%.7f",
			availableSellingCapacity.Selling, incrementalSellAmount, availableBuyingCapacity.Buying, incrementalBuyAmount)
	}

	if availableSellingCapacity.Selling <= 0 || availableBuyingCapacity.Buying <= 0 {
		log.Printf("computed remainder amount, no capacity available: availableSellingCapacity=%.7f, availableBuyingCapacity=%.7f\n", availableSellingCapacity.Selling, availableBuyingCapacity.Buying)
		return 0, 0, nil
	}

	// return the smaller amount between the buying and selling capacities that will max out either one
	if availableSellingCapacity.Selling*price < availableBuyingCapacity.Buying {
		sellingAmount := availableSellingCapacity.Selling
		buyingAmount := availableSellingCapacity.Selling * price
		log.Printf("computed remainder amount, constrained by selling capacity, returning sellingAmount=%.7f, buyingAmount=%.7f\n", sellingAmount, buyingAmount)
		return sellingAmount, buyingAmount, nil
	} else if availableBuyingCapacity.Buying/price < availableBuyingCapacity.Selling {
		sellingAmount := availableBuyingCapacity.Buying / price
		buyingAmount := availableBuyingCapacity.Buying
		log.Printf("computed remainder amount, constrained by buying capacity, returning sellingAmount=%.7f, buyingAmount=%.7f\n", sellingAmount, buyingAmount)
		return sellingAmount, buyingAmount, nil
	}
	return 0, 0, fmt.Errorf("error: (programmer?) unable to constrain by either buying capacity or selling capacity, sellingCapacity=%.7f, buyingCapacity=%.7f, price=%.7f",
		availableSellingCapacity.Selling, availableBuyingCapacity.Buying, price)
}

// createSellLevel returns offerPrice, hitCapacityLimit, op, error.
func (s *sellSideStrategy) createSellLevel(index int, targetPrice model.Number, targetAmount model.Number) (*model.Number, bool, *build.ManageOfferBuilder, error) {
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(true)
	targetPrice = *model.NumberByCappingPrecision(&targetPrice, utils.SdexPrecision)
	targetAmount = *model.NumberByCappingPrecision(&targetAmount, utils.SdexPrecision)

	hitCapacityLimit, op, e := s.placeOrderWithRetry(
		targetPrice.AsFloat(),
		targetAmount.AsFloat(),
		incrementalNativeAmountRaw,
		func(price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
			priceLogged := price
			amountLogged := amount
			if s.divideAmountByPrice {
				priceLogged = 1 / price
				amountLogged = amount * price
			}
			log.Printf("%s | create | level=%d | priceQuote=%.7f | amtBase=%.7f\n", s.action, index+1, priceLogged, amountLogged)
			return s.sdex.CreateSellOffer(*s.assetBase, *s.assetQuote, price, amount, incrementalNativeAmountRaw)
		},
		*s.assetBase,
		*s.assetQuote,
	)
	return &targetPrice, hitCapacityLimit, op, e
}

// modifySellLevel returns offerPrice, hitCapacityLimit, op, error.
func (s *sellSideStrategy) modifySellLevel(offers []horizon.Offer, index int, targetPrice model.Number, targetAmount model.Number) (*model.Number, bool, *build.ManageOfferBuilder, error) {
	highestPrice := targetPrice.AsFloat() + targetPrice.AsFloat()*s.priceTolerance
	lowestPrice := targetPrice.AsFloat() - targetPrice.AsFloat()*s.priceTolerance
	minAmount := targetAmount.AsFloat() - targetAmount.AsFloat()*s.amountTolerance
	maxAmount := targetAmount.AsFloat() + targetAmount.AsFloat()*s.amountTolerance

	//check if existing offer needs to be modified
	curPrice := utils.GetPrice(offers[index])
	curAmount := utils.AmountStringAsFloat(offers[index].Amount)

	// existing offer not within tolerances
	priceTrigger := (curPrice > highestPrice) || (curPrice < lowestPrice)
	amountTrigger := (curAmount < minAmount) || (curAmount > maxAmount)
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(false)
	if !priceTrigger && !amountTrigger {
		// always add back the current offer in the cached liabilities when we don't modify it
		s.sdex.AddLiabilities(offers[index].Selling, offers[index].Buying, curAmount, curAmount*curPrice, incrementalNativeAmountRaw)
		offerPrice := model.NumberFromFloat(curPrice, utils.SdexPrecision)
		return offerPrice, false, nil, nil
	}

	targetPrice = *model.NumberByCappingPrecision(&targetPrice, utils.SdexPrecision)
	targetAmount = *model.NumberByCappingPrecision(&targetAmount, utils.SdexPrecision)
	hitCapacityLimit, op, e := s.placeOrderWithRetry(
		targetPrice.AsFloat(),
		targetAmount.AsFloat(),
		incrementalNativeAmountRaw,
		func(price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
			priceLogged := price
			amountLogged := amount
			curPriceLogged := curPrice
			lowestPriceLogged := lowestPrice
			highestPriceLogged := highestPrice
			curAmountLogged := curAmount
			minAmountLogged := minAmount
			maxAmountLogged := maxAmount
			if s.divideAmountByPrice {
				priceLogged = 1 / price
				amountLogged = amount * price
				curPriceLogged = 1 / curPrice
				curAmountLogged = curAmount * curPrice
				minAmountLogged = minAmount * curPrice
				maxAmountLogged = maxAmount * curPrice
				// because we flip prices, the low and high need to be swapped here
				lowestPriceLogged = 1 / highestPrice
				highestPriceLogged = 1 / lowestPrice
			}
			log.Printf("%s | modify | level=%d | targetPriceQuote=%.7f | targetAmtBase=%.7f | curPriceQuote=%.7f | lowPriceQuote=%.7f | highPriceQuote=%.7f | curAmtBase=%.7f | minAmtBase=%.7f | maxAmtBase=%.7f\n",
				s.action, index+1, priceLogged, amountLogged, curPriceLogged, lowestPriceLogged, highestPriceLogged, curAmountLogged, minAmountLogged, maxAmountLogged)
			return s.sdex.ModifySellOffer(offers[index], price, amount, incrementalNativeAmountRaw)
		},
		offers[index].Selling,
		offers[index].Buying,
	)
	return &targetPrice, hitCapacityLimit, op, e
}

// placeOrderWithRetry returns hitCapacityLimit, op, error
func (s *sellSideStrategy) placeOrderWithRetry(
	targetPrice float64,
	targetAmount float64,
	incrementalNativeAmountRaw float64,
	placeOffer func(price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error),
	assetBase horizon.Asset,
	assetQuote horizon.Asset,
) (bool, *build.ManageOfferBuilder, error) {
	op, e := placeOffer(targetPrice, targetAmount, incrementalNativeAmountRaw)
	if e != nil {
		return false, nil, e
	}
	incrementalSellAmount := targetAmount
	incrementalBuyAmount := targetAmount * targetPrice
	// op is nil only when we hit capacity limits
	if op != nil {
		// update the cached liabilities if we create a valid operation to create an offer
		s.sdex.AddLiabilities(assetBase, assetQuote, incrementalSellAmount, incrementalBuyAmount, incrementalNativeAmountRaw)
		return false, op, nil
	}

	// place an order for the remainder between our intended amount and our remaining capacity
	newSellingAmount, newBuyingAmount, e := s.computeRemainderAmount(incrementalSellAmount, incrementalBuyAmount, targetPrice, incrementalNativeAmountRaw)
	if e != nil {
		return true, nil, e
	}
	if newSellingAmount == 0 || newBuyingAmount == 0 {
		return true, nil, nil
	}

	op, e = placeOffer(targetPrice, newSellingAmount, incrementalNativeAmountRaw)
	if e != nil {
		return true, nil, e
	}

	if op != nil {
		// update the cached liabilities if we create a valid operation to create an offer
		s.sdex.AddLiabilities(assetBase, assetQuote, newSellingAmount, newBuyingAmount, incrementalNativeAmountRaw)
		return true, op, nil
	}
	return true, nil, fmt.Errorf("error: (programmer?) unable to place offer with the new (reduced) selling and buying amounts, oldSellingAmount=%.7f, newSellingAmount=%.7f, oldBuyingAmount=%.7f, newBuyingAmount=%.7f",
		incrementalSellAmount, newSellingAmount, incrementalBuyAmount, newBuyingAmount)
}

// GetFillHandlers impl
func (s *sellSideStrategy) GetFillHandlers() ([]api.FillHandler, error) {
	return s.levelsProvider.GetFillHandlers()
}
