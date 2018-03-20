package sideStrategy

import (
	"fmt"
	"math"
	"strconv"

	"github.com/lightyeario/kelp/support/exchange/number"

	"github.com/lightyeario/kelp/alfonso/priceFeed"
	"github.com/lightyeario/kelp/alfonso/strategy/level"
	kelp "github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

// SellConfig contains the configuration params for this SideStrategy
type SellConfig struct {
	EXCHANGE         string        `valid:"-"`
	EXCHANGE_BASE    string        `valid:"-"`
	EXCHANGE_QUOTE   string        `valid:"-"`
	USE_BID_PRICE    string        `valid:"-"`
	PRICE_TOLERANCE  float64       `valid:"-"`
	AMOUNT_TOLERANCE float64       `valid:"-"`
	AMOUNT_OF_A_BASE float64       `valid:"-"` // the size of order
	LEVELS           []level.Level `valid:"-"`
}

// SellSideStrategy is a strategy to sell a specific currency on SDEX on a single side by reading prices from an exchange
type SellSideStrategy struct {
	txButler   *kelp.TxButler
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
	config     *SellConfig
	pf         priceFeed.FeedPair

	// uninitialized
	centerPrice   float64
	maxAssetBase  float64
	maxAssetQuote float64
}

// ensure it implements SideStrategy
var _ SideStrategy = &SellSideStrategy{}

// MakeSellSideStrategy is a factory method for SellSideStrategy
func MakeSellSideStrategy(
	txButler *kelp.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *SellConfig,
) SideStrategy {
	useBidPrice, e := strconv.ParseBool(config.USE_BID_PRICE)
	if e != nil {
		log.Panic("could not parse USE_BID_PRICE as a bool value: ", config.USE_BID_PRICE)
		return nil
	}

	exchangeFeedPairURL := fmt.Sprintf("%s/%s/%s/%v", config.EXCHANGE, config.EXCHANGE_BASE, config.EXCHANGE_QUOTE, useBidPrice)
	return &SellSideStrategy{
		txButler:   txButler,
		assetBase:  assetBase,
		assetQuote: assetQuote,
		config:     config,
		pf: *priceFeed.MakeFeedPair(
			"exchange", // hardcode "exchange" here because this pricefeed is from an exchange, at least for now
			exchangeFeedPairURL,
			"fixed", // this is a fixed value of 1 because the exchange priceFeed is a ratio of both assets
			"1.0",
		),
	}
}

// PruneExistingOffers impl
func (s *SellSideStrategy) PruneExistingOffers(offers []horizon.Offer) []horizon.Offer {
	for i := len(s.config.LEVELS); i < len(offers); i++ {
		s.txButler.DeleteOffer(offers[i])
	}
	if len(offers) > len(s.config.LEVELS) {
		offers = offers[:len(s.config.LEVELS)]
	}
	return offers
}

// PreUpdate impl
func (s *SellSideStrategy) PreUpdate(
	maxAssetBase float64,
	maxAssetQuote float64,
	offers []horizon.Offer,
) error {
	s.maxAssetBase = maxAssetBase
	s.maxAssetQuote = maxAssetQuote

	var e error
	s.centerPrice, e = s.pf.GetCenterPrice()
	if e != nil {
		log.Error("Center price couldn't be loaded! ", e)
	} else {
		log.Info("Center price: ", s.centerPrice, "        v0.2")
	}
	return e
}

// UpdateWithOps impl
func (s *SellSideStrategy) UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *number.Number, e error) {
	newTopOffer = nil
	for i := len(s.config.LEVELS) - 1; i >= 0; i-- {
		op := s.updateSellLevel(offers, i)
		if op != nil {
			offer, e := number.FromString(op.MO.Price.String(), 7)
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
func (s *SellSideStrategy) PostUpdate() error {
	return nil
}

// Selling Base
func (s *SellSideStrategy) updateSellLevel(offers []horizon.Offer, index int) *build.ManageOfferBuilder {
	spread := s.centerPrice * s.config.LEVELS[index].SPREAD
	targetPrice := s.centerPrice + spread
	targetAmount := s.config.LEVELS[index].AMOUNT * s.config.AMOUNT_OF_A_BASE
	targetAmount = math.Min(targetAmount, s.maxAssetBase)

	if len(offers) <= index {
		// no existing offer at this index
		log.Info("create sell: target:", targetPrice, " ta:", targetAmount)
		return s.txButler.CreateSellOffer(*s.assetBase, *s.assetQuote, targetPrice, targetAmount)
	}

	highestPrice := targetPrice + targetPrice*s.config.PRICE_TOLERANCE
	lowestPrice := targetPrice - targetPrice*s.config.PRICE_TOLERANCE
	minAmount := targetAmount - targetAmount*s.config.AMOUNT_TOLERANCE
	maxAmount := targetAmount + targetAmount*s.config.AMOUNT_TOLERANCE

	//check if existing offer needs to be modified
	curPrice := kelp.GetPrice(offers[index])
	curAmount := kelp.AmountStringAsFloat(offers[index].Amount)

	// existing offer not within tolerances
	priceTrigger := (curPrice > highestPrice) || (curPrice < lowestPrice)
	amountTrigger := (curAmount < minAmount) || (curAmount > maxAmount)
	if priceTrigger || amountTrigger {
		log.Info("mod sell curPrice: ", curPrice, ", highPrice: ", highestPrice, ", lowPrice: ", lowestPrice, ", curAmt: ", curAmount, ", minAmt: ", minAmount, ", maxAmt: ", maxAmount)
		return s.txButler.ModifySellOffer(offers[index], targetPrice, targetAmount)
	}
	return nil
}
