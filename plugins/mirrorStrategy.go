package plugins

import (
	"fmt"
	"log"
	"sync"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

type exchangeAPIKeysToml []struct {
	Key    string `valid:"-" toml:"KEY"`
	Secret string `valid:"-" toml:"SECRET"`
}

func (t *exchangeAPIKeysToml) toExchangeAPIKeys() []api.ExchangeAPIKey {
	apiKeys := []api.ExchangeAPIKey{}
	for _, apiKey := range *t {
		apiKeys = append(apiKeys, api.ExchangeAPIKey{
			Key:    apiKey.Key,
			Secret: apiKey.Secret,
		})
	}
	return apiKeys
}

// mirrorConfig contains the configuration params for this strategy
type mirrorConfig struct {
	Exchange        string              `valid:"-" toml:"EXCHANGE"`
	ExchangeBase    string              `valid:"-" toml:"EXCHANGE_BASE"`
	ExchangeQuote   string              `valid:"-" toml:"EXCHANGE_QUOTE"`
	OrderbookDepth  int32               `valid:"-" toml:"ORDERBOOK_DEPTH"`
	VolumeDivideBy  float64             `valid:"-" toml:"VOLUME_DIVIDE_BY"`
	PerLevelSpread  float64             `valid:"-" toml:"PER_LEVEL_SPREAD"`
	MinBaseVolume   float64             `valid:"-" toml:"MIN_BASE_VOLUME"`
	OffsetTrades    bool                `valid:"-" toml:"OFFSET_TRADES"`
	ExchangeAPIKeys exchangeAPIKeysToml `valid:"-" toml:"EXCHANGE_API_KEYS"`
}

// String impl.
func (c mirrorConfig) String() string {
	return utils.StructString(c, map[string]func(interface{}) interface{}{
		"EXCHANGE_API_KEYS": utils.Hide,
	})
}

// assetSurplus holds information about how many units of an asset needs to be offset on the exchange
// negative values mean we have eagerly offset an asset, likely because of minBaseVolume requirements of the backingExchange
type assetSurplus struct {
	total     *model.Number // total value in base asset units that are pending to be offset
	committed *model.Number // base asset units that are already committed to being offset
}

// makeAssetSurplus is a factory method
func makeAssetSurplus() *assetSurplus {
	return &assetSurplus{
		total:     model.NumberConstants.Zero,
		committed: model.NumberConstants.Zero,
	}
}

// mirrorStrategy is a strategy to mirror the orderbook of a given exchange
type mirrorStrategy struct {
	sdex               *SDEX
	ieif               *IEIF
	baseAsset          *horizon.Asset
	quoteAsset         *horizon.Asset
	primaryConstraints *model.OrderConstraints
	backingPair        *model.TradingPair
	backingConstraints *model.OrderConstraints
	orderbookDepth     int32
	perLevelSpread     float64
	volumeDivideBy     float64
	exchange           api.Exchange
	offsetTrades       bool
	mutex              *sync.Mutex
	baseSurplus        map[model.OrderAction]*assetSurplus // baseSurplus keeps track of any surplus we have of the base asset that needs to be offset on the backing exchange

	// uninitialized
	maxBackingBase  *model.Number
	maxBackingQuote *model.Number
}

// ensure this implements api.Strategy
var _ api.Strategy = &mirrorStrategy{}

// ensure this implements api.FillHandler
var _ api.FillHandler = &mirrorStrategy{}

// makeMirrorStrategy is a factory method
func makeMirrorStrategy(sdex *SDEX, ieif *IEIF, pair *model.TradingPair, baseAsset *horizon.Asset, quoteAsset *horizon.Asset, config *mirrorConfig, simMode bool) (api.Strategy, error) {
	var exchange api.Exchange
	var e error
	if config.OffsetTrades {
		exchangeAPIKeys := config.ExchangeAPIKeys.toExchangeAPIKeys()
		exchange, e = MakeTradingExchange(config.Exchange, exchangeAPIKeys, simMode)
		if e != nil {
			return nil, e
		}

		if config.MinBaseVolume == 0.0 {
			return nil, fmt.Errorf("need to specify non-zero MIN_BASE_VOLUME config param in mirror strategy config file")
		}
	} else {
		exchange, e = MakeExchange(config.Exchange, simMode)
		if e != nil {
			return nil, e
		}
	}

	// we have two sets of (tradingPair, orderConstraints): the primaryExchange and the backingExchange
	primaryConstraints := sdex.GetOrderConstraints(pair)
	// backingPair is taken from the mirror strategy config not from the passed in trading pair
	backingPair := &model.TradingPair{
		Base:  exchange.GetAssetConverter().MustFromString(config.ExchangeBase),
		Quote: exchange.GetAssetConverter().MustFromString(config.ExchangeQuote),
	}
	backingConstraints := exchange.GetOrderConstraints(backingPair)
	if config.OffsetTrades {
		backingConstraints.MinBaseVolume = *model.NumberFromFloat(config.MinBaseVolume, backingConstraints.VolumePrecision)
	}
	log.Printf("primaryPair='%s', primaryConstraints=%s\n", pair, primaryConstraints)
	log.Printf("backingPair='%s', backingConstraints=%s\n", backingPair, backingConstraints)
	return &mirrorStrategy{
		sdex:               sdex,
		ieif:               ieif,
		baseAsset:          baseAsset,
		quoteAsset:         quoteAsset,
		primaryConstraints: primaryConstraints,
		backingPair:        backingPair,
		backingConstraints: backingConstraints,
		orderbookDepth:     config.OrderbookDepth,
		perLevelSpread:     config.PerLevelSpread,
		volumeDivideBy:     config.VolumeDivideBy,
		exchange:           exchange,
		offsetTrades:       config.OffsetTrades,
		mutex:              &sync.Mutex{},
		baseSurplus: map[model.OrderAction]*assetSurplus{
			model.OrderActionBuy:  makeAssetSurplus(),
			model.OrderActionSell: makeAssetSurplus(),
		},
	}, nil
}

// PruneExistingOffers deletes any extra offers
func (s *mirrorStrategy) PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer, []horizon.Offer) {
	return []build.TransactionMutator{}, buyingAOffers, sellingAOffers
}

// PreUpdate changes the strategy's state in prepration for the update
func (s *mirrorStrategy) PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error {
	if s.offsetTrades {
		return s.recordBalances()
	}
	return nil
}

func (s *mirrorStrategy) recordBalances() error {
	balanceMap, e := s.exchange.GetAccountBalances([]interface{}{s.backingPair.Base, s.backingPair.Quote})
	if e != nil {
		return fmt.Errorf("unable to fetch balances for assets: %s", e)
	}

	// save asset balances from backing exchange to be used when placing offers in offset mode
	if baseBalance, ok := balanceMap[s.backingPair.Base]; ok {
		s.maxBackingBase = &baseBalance
	} else {
		return fmt.Errorf("unable to fetch balance for base asset: %s", string(s.backingPair.Base))
	}

	if quoteBalance, ok := balanceMap[s.backingPair.Quote]; ok {
		s.maxBackingQuote = &quoteBalance
	} else {
		return fmt.Errorf("unable to fetch balance for quote asset: %s", string(s.backingPair.Quote))
	}

	return nil
}

// UpdateWithOps builds the operations we want performed on the account
func (s *mirrorStrategy) UpdateWithOps(
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) ([]build.TransactionMutator, error) {
	ob, e := s.exchange.GetOrderBook(s.backingPair, s.orderbookDepth)
	if e != nil {
		return nil, e
	}

	// limit bids and asks to max 50 operations each because of Stellar's limit of 100 ops/tx
	bids := ob.Bids()
	if len(bids) > 50 {
		bids = bids[:50]
	}
	asks := ob.Asks()
	if len(asks) > 50 {
		asks = asks[:50]
	}

	sellBalanceCoordinator := balanceCoordinator{
		placedUnits:      model.NumberConstants.Zero,
		backingBalance:   s.maxBackingBase,
		backingAssetType: "base",
		isBackingBuy:     false,
	}
	buyOps, e := s.updateLevels(
		buyingAOffers,
		bids,
		s.sdex.ModifyBuyOffer,
		s.sdex.CreateBuyOffer,
		(1 - s.perLevelSpread),
		true,
		sellBalanceCoordinator, // we sell on the backing exchange to offset trades that are bought on the primary exchange
	)
	if e != nil {
		return nil, e
	}
	log.Printf("num. buyOps in this update: %d\n", len(buyOps))

	buyBalanceCoordinator := balanceCoordinator{
		placedUnits:      model.NumberConstants.Zero,
		backingBalance:   s.maxBackingQuote,
		backingAssetType: "quote",
		isBackingBuy:     true,
	}
	sellOps, e := s.updateLevels(
		sellingAOffers,
		asks,
		s.sdex.ModifySellOffer,
		s.sdex.CreateSellOffer,
		(1 + s.perLevelSpread),
		false,
		buyBalanceCoordinator, // we buy on the backing exchange to offset trades that are sold on the primary exchange
	)
	if e != nil {
		return nil, e
	}
	log.Printf("num. sellOps in this update: %d\n", len(sellOps))

	ops := []build.TransactionMutator{}
	if len(ob.Bids()) > 0 && len(sellingAOffers) > 0 && ob.Bids()[0].Price.AsFloat() >= utils.PriceAsFloat(sellingAOffers[0].Price) {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	} else {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	}

	return ops, nil
}

func (s *mirrorStrategy) updateLevels(
	oldOffers []horizon.Offer,
	newOrders []model.Order,
	modifyOffer func(offer horizon.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error),
	createOffer func(baseAsset horizon.Asset, quoteAsset horizon.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error),
	priceMultiplier float64,
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
	bc balanceCoordinator,
) ([]build.TransactionMutator, error) {
	ops := []build.TransactionMutator{}
	deleteOps := []build.TransactionMutator{}
	if len(newOrders) >= len(oldOffers) {
		for i := 0; i < len(oldOffers); i++ {
			modifyOp, deleteOp, e := s.doModifyOffer(oldOffers[i], newOrders[i], priceMultiplier, modifyOffer, hackPriceInvertForBuyOrderChangeCheck)
			if e != nil {
				return nil, e
			}
			if modifyOp != nil {
				if s.offsetTrades && !bc.checkBalance(newOrders[i].Volume, newOrders[i].Price) {
					continue
				}
				ops = append(ops, modifyOp)
			}
			if deleteOp != nil {
				deleteOps = append(deleteOps, deleteOp)
			}
		}

		// create offers for remaining new bids
		for i := len(oldOffers); i < len(newOrders); i++ {
			price := newOrders[i].Price.Scale(priceMultiplier)
			vol := newOrders[i].Volume.Scale(1.0 / s.volumeDivideBy)
			incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(true)

			if vol.AsFloat() < s.backingConstraints.MinBaseVolume.AsFloat() {
				log.Printf("skip level creation, baseVolume (%s) < minBaseVolume (%s) of backing exchange\n", vol.AsString(), s.backingConstraints.MinBaseVolume.AsString())
				continue
			}

			if s.offsetTrades && !bc.checkBalance(vol, price) {
				continue
			}

			mo, e := createOffer(*s.baseAsset, *s.quoteAsset, price.AsFloat(), vol.AsFloat(), incrementalNativeAmountRaw)
			if e != nil {
				return nil, e
			}
			if mo != nil {
				ops = append(ops, *mo)
				// update the cached liabilities if we create a valid operation to create an offer
				if hackPriceInvertForBuyOrderChangeCheck {
					s.ieif.AddLiabilities(*s.quoteAsset, *s.baseAsset, vol.Multiply(*price).AsFloat(), vol.AsFloat(), incrementalNativeAmountRaw)
				} else {
					s.ieif.AddLiabilities(*s.baseAsset, *s.quoteAsset, vol.AsFloat(), vol.Multiply(*price).AsFloat(), incrementalNativeAmountRaw)
				}
			}
		}
	} else {
		for i := 0; i < len(newOrders); i++ {
			modifyOp, deleteOp, e := s.doModifyOffer(oldOffers[i], newOrders[i], priceMultiplier, modifyOffer, hackPriceInvertForBuyOrderChangeCheck)
			if e != nil {
				return nil, e
			}
			if modifyOp != nil {
				if s.offsetTrades && !bc.checkBalance(newOrders[i].Volume, newOrders[i].Price) {
					continue
				}
				ops = append(ops, modifyOp)
			}
			if deleteOp != nil {
				deleteOps = append(deleteOps, deleteOp)
			}
		}

		// delete remaining prior offers
		for i := len(newOrders); i < len(oldOffers); i++ {
			deleteOp := s.sdex.DeleteOffer(oldOffers[i])
			deleteOps = append(deleteOps, deleteOp)
		}
	}

	// prepend deleteOps because we want to delete offers first so we "free" up our liabilities capacity to place the new/modified offers
	allOps := append(deleteOps, ops...)
	log.Printf("prepended %d deleteOps\n", len(deleteOps))

	return allOps, nil
}

// doModifyOffer returns a new modifyOp, deleteOp, error
func (s *mirrorStrategy) doModifyOffer(
	oldOffer horizon.Offer,
	newOrder model.Order,
	priceMultiplier float64,
	modifyOffer func(offer horizon.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error),
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
) (build.TransactionMutator, build.TransactionMutator, error) {
	price := newOrder.Price.Scale(priceMultiplier)
	vol := newOrder.Volume.Scale(1.0 / s.volumeDivideBy)
	oldPrice := model.MustNumberFromString(oldOffer.Price, s.primaryConstraints.PricePrecision)
	oldVol := model.MustNumberFromString(oldOffer.Amount, s.primaryConstraints.VolumePrecision)
	if hackPriceInvertForBuyOrderChangeCheck {
		// we want to multiply oldVol by the original oldPrice so we can get the correct oldVol, since ModifyBuyOffer multiplies price * vol
		oldVol = oldVol.Multiply(*oldPrice)
		oldPrice = model.InvertNumber(oldPrice)
	}
	epsilon := 0.0001
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(false)
	sameOrderParams := oldPrice.EqualsPrecisionNormalized(*price, epsilon) && oldVol.EqualsPrecisionNormalized(*vol, epsilon)
	if sameOrderParams {
		// update the cached liabilities if we keep the existing offer
		if hackPriceInvertForBuyOrderChangeCheck {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, oldVol.Multiply(*oldPrice).AsFloat(), oldVol.AsFloat(), incrementalNativeAmountRaw)
		} else {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, oldVol.AsFloat(), oldVol.Multiply(*oldPrice).AsFloat(), incrementalNativeAmountRaw)
		}
		return nil, nil, nil
	}

	// convert the precision from the backing exchange to the primary exchange
	offerPrice := model.NumberByCappingPrecision(price, s.primaryConstraints.PricePrecision)
	offerAmount := model.NumberByCappingPrecision(vol, s.primaryConstraints.VolumePrecision)
	if s.offsetTrades && offerAmount.AsFloat() < s.backingConstraints.MinBaseVolume.AsFloat() {
		log.Printf("deleting level, baseVolume (%f) on backing exchange dropped below minBaseVolume of backing exchange (%f)\n",
			offerAmount.AsFloat(), s.backingConstraints.MinBaseVolume.AsFloat())
		deleteOp := s.sdex.DeleteOffer(oldOffer)
		return nil, deleteOp, nil
	}
	mo, e := modifyOffer(
		oldOffer,
		offerPrice.AsFloat(),
		offerAmount.AsFloat(),
		incrementalNativeAmountRaw,
	)
	if e != nil {
		return nil, nil, e
	}
	if mo != nil {
		// update the cached liabilities if we create a valid operation to modify the offer
		if hackPriceInvertForBuyOrderChangeCheck {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, offerAmount.Multiply(*offerPrice).AsFloat(), offerAmount.AsFloat(), incrementalNativeAmountRaw)
		} else {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, offerAmount.AsFloat(), offerAmount.Multiply(*offerPrice).AsFloat(), incrementalNativeAmountRaw)
		}
		return *mo, nil, nil
	}

	// since mo is nil we want to delete this offer
	deleteOp := s.sdex.DeleteOffer(oldOffer)
	return nil, deleteOp, nil
}

// PostUpdate changes the strategy's state after the update has taken place
func (s *mirrorStrategy) PostUpdate() error {
	return nil
}

// GetFillHandlers impl
func (s *mirrorStrategy) GetFillHandlers() ([]api.FillHandler, error) {
	if s.offsetTrades {
		return []api.FillHandler{s}, nil
	}
	return nil, nil
}

func (s *mirrorStrategy) baseVolumeToOffset(trade model.Trade, newOrderAction model.OrderAction) (newVolume *model.Number, ok bool) {
	uncommittedBase := s.baseSurplus[newOrderAction].total.Subtract(*s.baseSurplus[newOrderAction].committed)

	if uncommittedBase.AsFloat() < s.backingConstraints.MinBaseVolume.Scale(0.5).AsFloat() {
		log.Printf("offset-skip | tradeID=%s | tradeBaseAmt=%f | tradeQuoteAmt=%f | tradePriceQuote=%f | minBaseVolume=%f | newOrderAction=%s | baseSurplusTotal=%f | baseSurplusCommitted=%f\n",
			trade.TransactionID.String(),
			trade.Volume.AsFloat(),
			trade.Volume.Multiply(*trade.Price).AsFloat(),
			trade.Price.AsFloat(),
			s.backingConstraints.MinBaseVolume.AsFloat(),
			newOrderAction.String(),
			s.baseSurplus[newOrderAction].total.AsFloat(),
			s.baseSurplus[newOrderAction].committed.AsFloat())
		return nil, false
	}

	if uncommittedBase.AsFloat() > s.backingConstraints.MinBaseVolume.AsFloat() {
		newVolume = uncommittedBase
	} else {
		// we want to offset the MinBaseVolume and take a deficit in the baseSurplus on success
		newVolume = &s.backingConstraints.MinBaseVolume
	}
	return model.NumberByCappingPrecision(newVolume, s.backingConstraints.VolumePrecision), true
}

// HandleFill impl
func (s *mirrorStrategy) HandleFill(trade model.Trade) error {
	// we should only ever have one active fill handler to avoid inconsistent R/W on baseSurplus
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newOrderAction := trade.OrderAction.Reverse()
	// increase the baseSurplus for the additional amount that needs to be offset because of the incoming trade
	s.baseSurplus[newOrderAction].total = s.baseSurplus[newOrderAction].total.Add(*trade.Volume)

	newVolume, ok := s.baseVolumeToOffset(trade, newOrderAction)
	if !ok {
		return nil
	}
	// commit the newVolume that we are trying to use so the next handler does not double-count this amount
	s.baseSurplus[newOrderAction].committed = s.baseSurplus[newOrderAction].committed.Add(*newVolume)

	newOrder := model.Order{
		Pair:        s.backingPair, // we want to offset trades on the backing exchange so use the backing exchange's trading pair
		OrderAction: newOrderAction,
		OrderType:   model.OrderTypeLimit,
		Price:       model.NumberByCappingPrecision(trade.Price, s.backingConstraints.PricePrecision),
		Volume:      newVolume,
		Timestamp:   nil,
	}
	log.Printf("offset-attempt | tradeID=%s | tradeBaseAmt=%f | tradeQuoteAmt=%f | tradePriceQuote=%f | newOrderAction=%s | baseSurplusTotal=%f | baseSurplusCommitted=%f | minBaseVolume=%f | newOrderBaseAmt=%f | newOrderQuoteAmt=%f | newOrderPriceQuote=%f\n",
		trade.TransactionID.String(),
		trade.Volume.AsFloat(),
		trade.Volume.Multiply(*trade.Price).AsFloat(),
		trade.Price.AsFloat(),
		newOrderAction.String(),
		s.baseSurplus[newOrderAction].total.AsFloat(),
		s.baseSurplus[newOrderAction].committed.AsFloat(),
		s.backingConstraints.MinBaseVolume.AsFloat(),
		newOrder.Volume.AsFloat(),
		newOrder.Volume.Multiply(*newOrder.Price).AsFloat(),
		newOrder.Price.AsFloat())
	transactionID, e := s.exchange.AddOrder(&newOrder)
	if e != nil {
		return fmt.Errorf("error when offsetting trade (newOrder=%s): %s", newOrder, e)
	}
	if transactionID == nil {
		return fmt.Errorf("error when offsetting trade (newOrder=%s): transactionID was <nil>", newOrder)
	}

	// update the baseSurplus on success
	s.baseSurplus[newOrderAction].total = s.baseSurplus[newOrderAction].total.Subtract(*newVolume)
	s.baseSurplus[newOrderAction].committed = s.baseSurplus[newOrderAction].committed.Subtract(*newVolume)

	log.Printf("offset-success | tradeID=%s | tradeBaseAmt=%f | tradeQuoteAmt=%f | tradePriceQuote=%f | newOrderAction=%s | baseSurplusTotal=%f | baseSurplusCommitted=%f | minBaseVolume=%f | newOrderBaseAmt=%f | newOrderQuoteAmt=%f | newOrderPriceQuote=%f | transactionID=%s\n",
		trade.TransactionID.String(),
		trade.Volume.AsFloat(),
		trade.Volume.Multiply(*trade.Price).AsFloat(),
		trade.Price.AsFloat(),
		newOrderAction.String(),
		s.baseSurplus[newOrderAction].total.AsFloat(),
		s.baseSurplus[newOrderAction].committed.AsFloat(),
		s.backingConstraints.MinBaseVolume.AsFloat(),
		newOrder.Volume.AsFloat(),
		newOrder.Volume.Multiply(*newOrder.Price).AsFloat(),
		newOrder.Price.AsFloat(),
		transactionID)
	return nil
}

// balanceCoordinator coordinates the balances from the backing exchange with orders placed on the primary exchange
type balanceCoordinator struct {
	placedUnits      *model.Number
	backingBalance   *model.Number
	backingAssetType string
	isBackingBuy     bool
}

func (b *balanceCoordinator) checkBalance(vol *model.Number, price *model.Number) bool {
	additionalUnits := vol
	if b.isBackingBuy {
		additionalUnits = vol.Multiply(*price)
	}

	newPlacedUnits := b.placedUnits.Add(*additionalUnits)
	if newPlacedUnits.AsFloat() > b.backingBalance.AsFloat() {
		log.Printf("skip level creation, not enough balance of %s asset on backing exchange: %s (needs at least %s)\n", b.backingAssetType, b.backingBalance.AsString(), newPlacedUnits.AsString())
		return false
	}

	b.placedUnits = newPlacedUnits
	return true
}
