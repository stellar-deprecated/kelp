package plugins

import (
	"fmt"
	"log"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// sellSideStrategy is a strategy to sell a specific currency on SDEX on a single side by reading prices from an exchange
type sellSideStrategy struct {
	sdex                *SDEX
	assetBase           *horizon.Asset
	assetQuote          *horizon.Asset
	levelsProvider      api.LevelProvider
	priceTolerance      float64
	amountTolerance     float64
	divideAmountByPrice bool

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
	return &sellSideStrategy{
		sdex:                sdex,
		assetBase:           assetBase,
		assetQuote:          assetQuote,
		levelsProvider:      levelsProvider,
		priceTolerance:      priceTolerance,
		amountTolerance:     amountTolerance,
		divideAmountByPrice: divideAmountByPrice,
	}
}

// PruneExistingOffers impl
func (s *sellSideStrategy) PruneExistingOffers(offers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer) {
	pruneOps := []build.TransactionMutator{}
	for i := len(s.currentLevels); i < len(offers); i++ {
		pOp := s.sdex.DeleteOffer(offers[i])
		pruneOps = append(pruneOps, &pOp)
	}
	if len(offers) > len(s.currentLevels) {
		offers = offers[:len(s.currentLevels)]
	}
	return pruneOps, offers
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

// UpdateWithOps impl
func (s *sellSideStrategy) UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error) {
	deleteOps := []build.TransactionMutator{}
	newTopOffer = nil
	hitCapacityLimit := false
	for i := 0; i < len(s.currentLevels); i++ {
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
		targetPrice := s.currentLevels[i].Price
		targetAmount := s.currentLevels[i].Amount
		if s.divideAmountByPrice {
			targetAmount = *model.NumberFromFloat(targetAmount.AsFloat()/targetPrice.AsFloat(), targetAmount.Precision())
		}

		var offerPrice *model.Number
		var op *build.ManageOfferBuilder
		var e error
		if isModify {
			offerPrice, hitCapacityLimit, op, e = s.modifySellLevel(offers, i, targetPrice, targetAmount)
		} else {
			offerPrice, hitCapacityLimit, op, e = s.createSellLevel(targetPrice, targetAmount)
		}
		if e != nil {
			return nil, nil, e
		}
		if op != nil {
			ops = append(ops, op)
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
func (s *sellSideStrategy) createSellLevel(targetPrice model.Number, targetAmount model.Number) (*model.Number, bool, *build.ManageOfferBuilder, error) {
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(true)
	targetPrice = *model.NumberByCappingPrecision(&targetPrice, utils.SdexPrecision)
	targetAmount = *model.NumberByCappingPrecision(&targetAmount, utils.SdexPrecision)

	hitCapacityLimit, op, e := s.placeOrderWithRetry(
		targetPrice.AsFloat(),
		targetAmount.AsFloat(),
		incrementalNativeAmountRaw,
		func(price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error) {
			log.Printf("sell,create,p=%.7f,a=%.7f\n", price, amount)
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
			log.Printf("sell,modify,tp=%.7f,ta=%.7f,curPrice=%.7f,highPrice=%.7f,lowPrice=%.7f,curAmt=%.7f,minAmt=%.7f,maxAmt=%.7f\n",
				price, amount, curPrice, highestPrice, lowestPrice, curAmount, minAmount, maxAmount)
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
