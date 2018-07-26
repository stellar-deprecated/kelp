package plugins

import (
	"math"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
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
		log.Warnf("no capacity to place sell orders (nothingToSell = %v, lineFull = %v)\n", nothingToSell, lineFull)
		return nil
	}

	// load currentLevels only once here
	var e error
	s.currentLevels, e = s.levelsProvider.GetLevels(s.maxAssetBase, s.maxAssetQuote)
	if e != nil {
		log.Error("levels couldn't be loaded: ", e)
		return e
	}
	return nil
}

// UpdateWithOps impl
func (s *sellSideStrategy) UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error) {
	newTopOffer = nil
	for i := len(s.currentLevels) - 1; i >= 0; i-- {
		op := s.updateSellLevel(offers, i)
		if op != nil {
			offer, e := model.NumberFromString(op.MO.Price.String(), 7)
			if e != nil {
				return nil, nil, e
			}

			// newTopOffer is minOffer because this is a sell strategy, and the lowest price is the best (top) price on the orderbook
			if newTopOffer == nil || offer.AsFloat() < newTopOffer.AsFloat() {
				newTopOffer = offer
			}

			ops = append(ops, op)
		}
	}
	return ops, newTopOffer, nil
}

// PostUpdate impl
func (s *sellSideStrategy) PostUpdate() error {
	return nil
}

// Selling Base
func (s *sellSideStrategy) updateSellLevel(offers []horizon.Offer, index int) *build.ManageOfferBuilder {
	targetPrice := s.currentLevels[index].Price
	targetAmount := s.currentLevels[index].Amount
	if s.divideAmountByPrice {
		targetAmount /= targetPrice
	}
	targetAmount = math.Min(targetAmount, s.maxAssetBase)

	if len(offers) <= index {
		// no existing offer at this index
		log.Info("create sell: target:", targetPrice, " ta:", targetAmount)
		return s.sdex.CreateSellOffer(*s.assetBase, *s.assetQuote, targetPrice, targetAmount)
	}

	highestPrice := targetPrice + targetPrice*s.priceTolerance
	lowestPrice := targetPrice - targetPrice*s.priceTolerance
	minAmount := targetAmount - targetAmount*s.amountTolerance
	maxAmount := targetAmount + targetAmount*s.amountTolerance

	//check if existing offer needs to be modified
	curPrice := utils.GetPrice(offers[index])
	curAmount := utils.AmountStringAsFloat(offers[index].Amount)

	// existing offer not within tolerances
	priceTrigger := (curPrice > highestPrice) || (curPrice < lowestPrice)
	amountTrigger := (curAmount < minAmount) || (curAmount > maxAmount)
	if priceTrigger || amountTrigger {
		log.Info("mod sell curPrice: ", curPrice, ", highPrice: ", highestPrice, ", lowPrice: ", lowestPrice, ", curAmt: ", curAmount, ", minAmt: ", minAmount, ", maxAmt: ", maxAmount)
		return s.sdex.ModifySellOffer(offers[index], targetPrice, targetAmount)
	}
	return nil
}
