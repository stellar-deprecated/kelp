package strategy

import (

	//"context"
	"math"

	"github.com/lightyeario/kelp/alfonso/priceFeed"
	"github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

/*
Prices are (amount of B)/(amount of A)
1BTC/5000xlm=price

Amounts are in A
*/

// Level represents a layer in the orderbook
type Level struct {
	SPREAD float64 `valid:"-"`
	AMOUNT float64 `valid:"-"`
}

// SimpleConfig contains the configuration params for this strategy
type SimpleConfig struct {
	PRICE_TOLERANCE  float64 `valid:"-"`
	AMOUNT_TOLERANCE float64 `valid:"-"`
	AMOUNT_OF_A_BASE float64 `valid:"-"` // the size of order to keep on either side
	DATA_TYPE_A      string  `valid:"-"`
	DATA_FEED_A_URL  string  `valid:"-"`
	DATA_TYPE_B      string  `valid:"-"`
	DATA_FEED_B_URL  string  `valid:"-"`
	LEVELS           []Level `valid:"-"`
}

// SimpleStrategy is a simple market maker strategy that puts buy and sell orders
// on each side of a "centered" price based on a configuration file
type SimpleStrategy struct {
	txButler *kelp.TxButler
	assetA   horizon.Asset
	assetB   horizon.Asset
	config   *SimpleConfig
	pf       priceFeed.FeedPair

	// uninitialized
	centerPrice           float64
	lastPlacedCenterPrice float64
	maxAssetA             float64
	maxAssetB             float64
}

// ensure that SimpleStrategy implements Strategy
var _ Strategy = &SimpleStrategy{}

// MakeSimpleStrategy is a factory method
func MakeSimpleStrategy(txButler *kelp.TxButler, assetA horizon.Asset, assetB horizon.Asset, config *SimpleConfig) Strategy {
	return &SimpleStrategy{
		txButler: txButler,
		assetA:   assetA,
		assetB:   assetB,
		config:   config,
		pf: *priceFeed.MakeFeedPair(
			config.DATA_TYPE_A,
			config.DATA_FEED_A_URL,
			config.DATA_TYPE_B,
			config.DATA_FEED_B_URL,
		),
	}
}

// PruneExistingOffers deletes any extra offers
func (s SimpleStrategy) PruneExistingOffers(offers []horizon.Offer) []horizon.Offer {
	for i := len(s.config.LEVELS); i < len(offers); i++ {
		s.txButler.DeleteOffer(offers[i])
	}
	if len(offers) > len(s.config.LEVELS) {
		return offers[:len(s.config.LEVELS)]
	}
	return offers
}

// PreUpdate changes the strategy's state in prepration for the update
func (s *SimpleStrategy) PreUpdate(maxAssetA float64, maxAssetB float64) error {
	s.maxAssetA = maxAssetA
	s.maxAssetB = maxAssetB

	var e error
	s.centerPrice, e = s.pf.GetCenterPrice()
	if e != nil {
		log.Error("Center price couldn't be loaded! ", e)
	} else {
		log.Info("Center price: ", s.centerPrice, "        v0.2")
	}
	return e
}

// UpdateWithOps builds the operations we want performed on the account
func (s SimpleStrategy) UpdateWithOps(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, error) {
	sellOps := s.updateLevels(bindOffersToUpdateLevel(s.updateSellLevel, sellingAOffers))
	buyOps := s.updateLevels(bindOffersToUpdateLevel(s.updateBuyLevel, buyingAOffers))

	ops := []build.TransactionMutator{}
	if s.centerPrice < s.lastPlacedCenterPrice {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	} else {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	}

	return ops, nil
}

// PostUpdate changes the strategy's state after the update has taken place
func (s *SimpleStrategy) PostUpdate() error {
	s.lastPlacedCenterPrice = s.centerPrice
	return nil
}

func bindOffersToUpdateLevel(
	updateLevel func(offers []horizon.Offer, index int) *build.ManageOfferBuilder,
	offers []horizon.Offer,
) func(index int) *build.ManageOfferBuilder {
	return func(index int) *build.ManageOfferBuilder {
		return updateLevel(offers, index)
	}
}

func (s *SimpleStrategy) updateLevels(
	updateLevel func(index int) *build.ManageOfferBuilder,
) []build.TransactionMutator {
	ops := []build.TransactionMutator{}
	for i := len(s.config.LEVELS) - 1; i >= 0; i-- {
		op := updateLevel(i)
		if op != nil {
			ops = append(ops, op)
		}
	}
	return ops
}

// Buying A
func (s *SimpleStrategy) updateBuyLevel(buyingAOffers []horizon.Offer, index int) *build.ManageOfferBuilder {
	spread := s.centerPrice * s.config.LEVELS[index].SPREAD
	targetPrice := s.centerPrice - spread

	targetAmount := s.config.LEVELS[index].AMOUNT * s.config.AMOUNT_OF_A_BASE
	targetAmount = math.Min(targetAmount, s.maxAssetB/targetPrice)

	if len(buyingAOffers) <= index {
		// no existing offer at this index
		log.Info("create buy: target:", targetPrice, " ta:", targetAmount)

		return s.txButler.CreateBuyOffer(s.assetA, s.assetB, targetPrice, targetAmount)
	}

	highestPrice := targetPrice + targetPrice*s.config.PRICE_TOLERANCE
	lowestPrice := targetPrice - targetPrice*s.config.PRICE_TOLERANCE
	//check if existing offer needs to be modified
	curPrice := kelp.GetInvertedPrice(buyingAOffers[index])
	curAmount := kelp.AmountStringAsFloat(buyingAOffers[index].Amount) * kelp.PriceAsFloat(buyingAOffers[index].Price)

	minAmount := targetAmount - targetAmount*s.config.AMOUNT_TOLERANCE
	maxAmount := targetAmount + targetAmount*s.config.AMOUNT_TOLERANCE

	if (curPrice > highestPrice) ||
		(curPrice < lowestPrice) ||
		(curAmount < minAmount) ||
		(curAmount > maxAmount) {
		// existing offer not within tolerances
		log.Info("mod buy:", curPrice, " a:", curAmount)
		return s.txButler.ModifyBuyOffer(buyingAOffers[index], targetPrice, targetAmount)
	}
	return nil
}

// Selling A
func (s *SimpleStrategy) updateSellLevel(sellingAOffers []horizon.Offer, index int) *build.ManageOfferBuilder {
	spread := s.centerPrice * s.config.LEVELS[index].SPREAD
	targetPrice := s.centerPrice + spread
	targetAmount := s.config.LEVELS[index].AMOUNT * s.config.AMOUNT_OF_A_BASE
	targetAmount = math.Min(targetAmount, s.maxAssetA)

	if len(sellingAOffers) <= index {
		// no existing offer at this index
		log.Info("create sell: target:", targetPrice, " ta:", targetAmount)
		return s.txButler.CreateSellOffer(s.assetA, s.assetB, targetPrice, targetAmount)
	}

	highestPrice := targetPrice + targetPrice*s.config.PRICE_TOLERANCE
	lowestPrice := targetPrice - targetPrice*s.config.PRICE_TOLERANCE

	//check if existing offer needs to be modified
	curPrice := kelp.GetPrice(sellingAOffers[index])
	curAmount := kelp.AmountStringAsFloat(sellingAOffers[index].Amount)

	minAmount := targetAmount - targetAmount*s.config.AMOUNT_TOLERANCE
	maxAmount := targetAmount + targetAmount*s.config.AMOUNT_TOLERANCE

	if (curPrice > highestPrice) ||
		(curPrice < lowestPrice) ||
		(curAmount < minAmount) ||
		(curAmount > maxAmount) {
		// existing offer not within tolerances
		log.Info("mod sell curPrice", curPrice, " hp:", highestPrice, " lp:", lowestPrice, " ca:", curAmount, " minA:", minAmount, " maxA:", maxAmount)

		return s.txButler.ModifySellOffer(sellingAOffers[index], targetPrice, targetAmount)
	}
	return nil
}
