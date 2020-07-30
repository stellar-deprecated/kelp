package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

const actionSell = "sell"
const actionBuy = "buy "

// sellSideStrategy is a strategy to sell a specific currency on SDEX on a single side by reading prices from an exchange
type sellSideStrategy struct {
	sdex                *SDEX
	orderConstraints    *model.OrderConstraints
	ieif                *IEIF
	assetBase           *hProtocol.Asset
	assetQuote          *hProtocol.Asset
	levelsProvider      api.LevelProvider
	priceTolerance      float64
	amountTolerance     float64
	divideAmountByPrice bool
	action              string

	// uninitialized
	desiredLevels []api.Level // levels for current iteration
	maxAssetBase  float64
	maxAssetQuote float64
}

// ensure it implements SideStrategy
var _ api.SideStrategy = &sellSideStrategy{}

// makeSellSideStrategy is a factory method for sellSideStrategy
func makeSellSideStrategy(
	sdex *SDEX,
	orderConstraints *model.OrderConstraints,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
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
		orderConstraints:    orderConstraints,
		ieif:                ieif,
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
func (s *sellSideStrategy) PruneExistingOffers(offers []hProtocol.Offer) ([]build.TransactionMutator, []hProtocol.Offer) {
	// figure out which offers we want to prune
	shouldPrune := computeOffersToPrune(offers, s.desiredLevels)

	pruneOps := []txnbuild.Operation{}
	updatedOffers := []hProtocol.Offer{}
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
		log.Printf("offer | %s | level=%d | curPriceQuote=%.8f | curAmtBase=%.8f | pruning=%v\n", s.action, i+1, curPrice, curAmount, isPruning)
	}
	return api.ConvertOperation2TM(pruneOps), updatedOffers
}

// computeOffersToPrune returns a list of bools representing whether we should prune the offer at that position or not
func computeOffersToPrune(offers []hProtocol.Offer, levels []api.Level) []bool {
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

	// load desiredLevels only once here
	// invoke before checking remaining capacity because we want to always execute GetLevels so it can update internal state (specifically applies to sellTwap)
	newLevels, e := s.levelsProvider.GetLevels(s.maxAssetBase, s.maxAssetQuote)
	if e != nil {
		log.Printf("levels couldn't be loaded: %s\n", e)
		return e
	}
	log.Printf("levels returned (side = %s): %v\n", s.action, newLevels)

	// don't place orders if we have nothing to sell or if we cannot buy the asset in exchange
	nothingToSell := maxAssetBase == 0
	lineFull := maxAssetQuote == trustQuote
	if nothingToSell || lineFull {
		s.desiredLevels = []api.Level{}
		log.Printf("no capacity to place sell orders (nothingToSell = %v, lineFull = %v)\n", nothingToSell, lineFull)
		return nil
	}

	// set desiredLevels only once here
	s.desiredLevels = newLevels
	return nil
}

// computePrecedingLevels returns the levels priced better than the lowest existing offer, up to the max preceding levels allowed
func computePrecedingLevels(offers []hProtocol.Offer, levels []api.Level) []api.Level {
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
		// for now we want to maintain the amount's precision here so we're not using number.Divide
		targetAmount = model.NumberFromFloat(targetAmount.AsFloat()/targetPrice.AsFloat(), targetAmount.Precision())
	}

	return targetPrice, targetAmount, nil
}

func (s *sellSideStrategy) createPrecedingOffers(
	precedingLevels []api.Level,
) (
	int, // numLevelsConsumed
	bool, // hitCapacityLimit
	[]txnbuild.Operation, // ops
	*model.Number, // newTopOffer
	error, // e
) {
	hitCapacityLimit := false
	ops := []txnbuild.Operation{}
	var newTopOffer *model.Number

	for i := 0; i < len(precedingLevels); i++ {
		if hitCapacityLimit {
			log.Printf("%s, hitCapacityLimit in preceding level loop, returning numLevelsConsumed=%d\n", s.action, i)
			return i, true, ops, newTopOffer, nil
		}

		targetPrice, targetAmount, e := s.computeTargets(precedingLevels[i])
		if e != nil {
			return 0, false, nil, nil, fmt.Errorf("could not compute targets: %s", e)
		}

		var offerPrice *model.Number
		var op *txnbuild.ManageSellOffer
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

	numLevelsConsumed := len(precedingLevels)
	newTopOfferPrice := "<nil>"
	if newTopOffer != nil {
		newTopOfferPrice = newTopOffer.AsString()
	}
	log.Printf("%s, done creating preceding offers (numLevelsConsumed=%d, hitCapacityLimit=%v, numOps=%d, newTopOfferPrice=%s)",
		s.action, numLevelsConsumed, hitCapacityLimit, len(ops), newTopOfferPrice,
	)

	// hitCapacityLimit can be updated after the check inside the for loop
	return numLevelsConsumed, hitCapacityLimit, ops, newTopOffer, nil
}

// UpdateWithOps impl
func (s *sellSideStrategy) UpdateWithOps(offers []hProtocol.Offer) (opsOld []build.TransactionMutator, newTopOffer *model.Number, e error) {
	var ops []txnbuild.Operation
	deleteOps := []txnbuild.Operation{}

	// first we want to re-create any offers that precede our existing offers and are additions to the existing offers that we have
	precedingLevels := computePrecedingLevels(offers, s.desiredLevels)
	var hitCapacityLimit bool
	var numLevelsConsumed int
	numLevelsConsumed, hitCapacityLimit, ops, newTopOffer, e = s.createPrecedingOffers(precedingLevels)
	if e != nil {
		return nil, nil, fmt.Errorf("unable to create preceding offers: %s", e)
	}

	// next we want to adjust our remaining offers to be in line with what is desired
	// either modifying the existing offers, or creating new offers at the end of our existing offers
	for i := numLevelsConsumed; i < len(s.desiredLevels); i++ {
		existingOffersIdx := i - numLevelsConsumed
		isModify := existingOffersIdx < len(offers)
		// we only want to delete offers after we hit the capacity limit which is why we perform this check in the beginning
		if hitCapacityLimit {
			if isModify {
				delOp := s.sdex.DeleteOffer(offers[existingOffersIdx])
				log.Printf("deleting offer because we previously hit the capacity limit, offerId=%d\n", offers[existingOffersIdx].ID)
				deleteOps = append(deleteOps, &delOp)
				continue
			} else {
				// we can break because we would never see a modify operation happen after a non-modify operation
				break
			}
		}

		// hitCapacityLimit can be updated below
		targetPrice, targetAmount, e := s.computeTargets(s.desiredLevels[i])
		if e != nil {
			return nil, nil, fmt.Errorf("could not compute targets: %s", e)
		}

		var offerPrice *model.Number
		var op *txnbuild.ManageSellOffer
		if isModify {
			offerPrice, hitCapacityLimit, op, e = s.modifySellLevel(offers, existingOffersIdx, i, *targetPrice, *targetAmount)
		} else {
			offerPrice, hitCapacityLimit, op, e = s.createSellLevel(i, *targetPrice, *targetAmount)
		}
		if e != nil {
			return nil, nil, fmt.Errorf("unable to update existing offers or create new offers: %s", e)
		}
		if op != nil {
			reducedOrderSize := isModify && targetAmount.AsFloat() < utils.AmountStringAsFloat(offers[existingOffersIdx].Amount)
			hitCapacityLimitModify := isModify && hitCapacityLimit
			if reducedOrderSize || hitCapacityLimitModify {
				// prepend operations that reduce the size of an existing order because they decrease our liabilities
				ops = append([]txnbuild.Operation{op}, ops...)
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

	return api.ConvertOperation2TM(ops), newTopOffer, nil
}

// PostUpdate impl
func (s *sellSideStrategy) PostUpdate() error {
	return nil
}

// computeRemainderAmount returns sellingAmount, buyingAmount, error
func (s *sellSideStrategy) computeRemainderAmount(incrementalSellAmount float64, incrementalBuyAmount float64, price float64, incrementalNativeAmountRaw float64) (float64, float64, error) {
	availableSellingCapacity, e := s.ieif.AvailableCapacity(*s.assetBase, incrementalNativeAmountRaw)
	if e != nil {
		return 0, 0, e
	}
	availableBuyingCapacity, e := s.ieif.AvailableCapacity(*s.assetQuote, incrementalNativeAmountRaw)
	if e != nil {
		return 0, 0, e
	}

	if availableSellingCapacity.Selling >= incrementalSellAmount && availableBuyingCapacity.Buying >= incrementalBuyAmount {
		return 0, 0, fmt.Errorf("error: (programmer?) unable to create offer but available capacities were more than the attempted offer amounts, sellingCapacity=%.8f, incrementalSellAmount=%.8f, buyingCapacity=%.8f, incrementalBuyAmount=%.8f",
			availableSellingCapacity.Selling, incrementalSellAmount, availableBuyingCapacity.Buying, incrementalBuyAmount)
	}

	if availableSellingCapacity.Selling <= 0 || availableBuyingCapacity.Buying <= 0 {
		log.Printf("computed remainder amount, no capacity available: availableSellingCapacity=%.8f, availableBuyingCapacity=%.8f\n", availableSellingCapacity.Selling, availableBuyingCapacity.Buying)
		return 0, 0, nil
	}

	// return the smaller amount between the buying and selling capacities that will max out either one
	if availableSellingCapacity.Selling*price < availableBuyingCapacity.Buying {
		sellingAmount := availableSellingCapacity.Selling
		buyingAmount := availableSellingCapacity.Selling * price
		log.Printf("computed remainder amount, constrained by selling capacity, returning sellingAmount=%.8f, buyingAmount=%.8f\n", sellingAmount, buyingAmount)
		return sellingAmount, buyingAmount, nil
	} else if availableBuyingCapacity.Buying/price < availableBuyingCapacity.Selling {
		sellingAmount := availableBuyingCapacity.Buying / price
		buyingAmount := availableBuyingCapacity.Buying
		log.Printf("computed remainder amount, constrained by buying capacity, returning sellingAmount=%.8f, buyingAmount=%.8f\n", sellingAmount, buyingAmount)
		return sellingAmount, buyingAmount, nil
	}
	return 0, 0, fmt.Errorf("error: (programmer?) unable to constrain by either buying capacity or selling capacity, sellingCapacity=%.8f, buyingCapacity=%.8f, price=%.8f",
		availableSellingCapacity.Selling, availableBuyingCapacity.Buying, price)
}

// createSellLevel returns offerPrice, hitCapacityLimit, op, error.
func (s *sellSideStrategy) createSellLevel(index int, targetPrice model.Number, targetAmount model.Number) (*model.Number, bool, *txnbuild.ManageSellOffer, error) {
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(true)
	targetPrice = *model.NumberByCappingPrecision(&targetPrice, s.orderConstraints.PricePrecision)
	targetAmount = *model.NumberByCappingPrecision(&targetAmount, s.orderConstraints.VolumePrecision)

	hitCapacityLimit, op, e := s.placeOrderWithRetry(
		targetPrice.AsFloat(),
		targetAmount.AsFloat(),
		incrementalNativeAmountRaw,
		func(price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
			priceLogged := price
			amountLogged := amount
			if s.divideAmountByPrice {
				priceLogged = 1 / price
				amountLogged = amount * price
			}
			log.Printf("%s | create | new level=%d | priceQuote=%.8f | amtBase=%.8f\n", s.action, index+1, priceLogged, amountLogged)
			return s.sdex.CreateSellOffer(*s.assetBase, *s.assetQuote, price, amount, incrementalNativeAmountRaw)
		},
		*s.assetBase,
		*s.assetQuote,
	)
	return &targetPrice, hitCapacityLimit, op, e
}

// modifySellLevel returns offerPrice, hitCapacityLimit, op, error.
func (s *sellSideStrategy) modifySellLevel(offers []hProtocol.Offer, index int, newIndex int, targetPrice model.Number, targetAmount model.Number) (*model.Number, bool, *txnbuild.ManageSellOffer, error) {
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
	var oversellTrigger bool
	sellingAsset := offers[index].Selling
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(false)
	var e error
	if sellingAsset == utils.NativeAsset {
		oversellTrigger, e = s.ieif.willOversellNative(curAmount + incrementalNativeAmountRaw)
		if e != nil {
			return nil, false, nil, fmt.Errorf("could not check oversellTrigger for native asset: %s", e)
		}
	} else {
		oversellTrigger, e = s.ieif.willOversell(sellingAsset, curAmount)
		if e != nil {
			return nil, false, nil, fmt.Errorf("could not check oversellTrigger for sellingAsset (%s): %s", utils.Asset2String(sellingAsset), e)
		}
	}
	if !priceTrigger && !amountTrigger && !oversellTrigger {
		// always add back the current offer in the cached liabilities when we don't modify it
		s.ieif.AddLiabilities(offers[index].Selling, offers[index].Buying, curAmount, curAmount*curPrice, incrementalNativeAmountRaw)
		log.Printf("%s | modify | unmodified original level = %d | newLevel number = %d\n", s.action, index+1, newIndex+1)
		offerPrice := model.NumberFromFloat(curPrice, s.orderConstraints.PricePrecision)
		return offerPrice, false, nil, nil
	}
	triggers := []string{}
	if priceTrigger {
		triggers = append(triggers, "price")
	}
	if amountTrigger {
		triggers = append(triggers, "amount")
	}
	if oversellTrigger {
		triggers = append(triggers, "oversell")
	}

	targetPrice = *model.NumberByCappingPrecision(&targetPrice, s.orderConstraints.PricePrecision)
	targetAmount = *model.NumberByCappingPrecision(&targetAmount, s.orderConstraints.VolumePrecision)
	hitCapacityLimit, op, e := s.placeOrderWithRetry(
		targetPrice.AsFloat(),
		targetAmount.AsFloat(),
		incrementalNativeAmountRaw,
		func(price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error) {
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
			log.Printf("%s | modify | old level=%d | new level = %d | triggers=%v | targetPriceQuote=%.8f | targetAmtBase=%.8f | curPriceQuote=%.8f | lowPriceQuote=%.8f | highPriceQuote=%.8f | curAmtBase=%.8f | minAmtBase=%.8f | maxAmtBase=%.8f\n",
				s.action, index+1, newIndex+1, triggers, priceLogged, amountLogged, curPriceLogged, lowestPriceLogged, highestPriceLogged, curAmountLogged, minAmountLogged, maxAmountLogged)
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
	placeOffer func(price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error),
	assetBase hProtocol.Asset,
	assetQuote hProtocol.Asset,
) (bool, *txnbuild.ManageSellOffer, error) {
	op, e := placeOffer(targetPrice, targetAmount, incrementalNativeAmountRaw)
	if e != nil {
		return false, nil, e
	}
	incrementalSellAmount := targetAmount
	incrementalBuyAmount := targetAmount * targetPrice
	// op is nil only when we hit capacity limits
	if op != nil {
		// update the cached liabilities if we create a valid operation to create an offer
		s.ieif.AddLiabilities(assetBase, assetQuote, incrementalSellAmount, incrementalBuyAmount, incrementalNativeAmountRaw)
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
		s.ieif.AddLiabilities(assetBase, assetQuote, newSellingAmount, newBuyingAmount, incrementalNativeAmountRaw)
		return true, op, nil
	}
	return true, nil, fmt.Errorf("error: (programmer?) unable to place offer with the new (reduced) selling and buying amounts, oldSellingAmount=%.8f, newSellingAmount=%.8f, oldBuyingAmount=%.8f, newBuyingAmount=%.8f",
		incrementalSellAmount, newSellingAmount, incrementalBuyAmount, newBuyingAmount)
}

// GetFillHandlers impl
func (s *sellSideStrategy) GetFillHandlers() ([]api.FillHandler, error) {
	return s.levelsProvider.GetFillHandlers()
}
