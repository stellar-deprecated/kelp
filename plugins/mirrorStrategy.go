package plugins

import (
	"log"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// mirrorConfig contains the configuration params for this strategy
type mirrorConfig struct {
	EXCHANGE         string  `valid:"-"`
	EXCHANGE_BASE    string  `valid:"-"`
	EXCHANGE_QUOTE   string  `valid:"-"`
	ORDERBOOK_DEPTH  int32   `valid:"-"`
	VOLUME_DIVIDE_BY float64 `valid:"-"`
	PER_LEVEL_SPREAD float64 `valid:"-"`
}

// String impl.
func (c mirrorConfig) String() string {
	return utils.StructString(c, nil)
}

// mirrorStrategy is a strategy to mirror the orderbook of a given exchange
type mirrorStrategy struct {
	sdex          *SDEX
	orderbookPair *model.TradingPair
	baseAsset     *horizon.Asset
	quoteAsset    *horizon.Asset
	config        *mirrorConfig
	tradeAPI      api.TradeAPI
}

// ensure this implements Strategy
var _ api.Strategy = &mirrorStrategy{}

// makeMirrorStrategy is a factory method
func makeMirrorStrategy(sdex *SDEX, baseAsset *horizon.Asset, quoteAsset *horizon.Asset, config *mirrorConfig) api.Strategy {
	exchange := MakeExchange(config.EXCHANGE)
	orderbookPair := &model.TradingPair{
		Base:  exchange.GetAssetConverter().MustFromString(config.EXCHANGE_BASE),
		Quote: exchange.GetAssetConverter().MustFromString(config.EXCHANGE_QUOTE),
	}
	return &mirrorStrategy{
		sdex:          sdex,
		orderbookPair: orderbookPair,
		baseAsset:     baseAsset,
		quoteAsset:    quoteAsset,
		config:        config,
		tradeAPI:      api.TradeAPI(exchange),
	}
}

// PruneExistingOffers deletes any extra offers
func (s mirrorStrategy) PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer, []horizon.Offer) {
	return []build.TransactionMutator{}, buyingAOffers, sellingAOffers
}

// PreUpdate changes the strategy's state in prepration for the update
func (s *mirrorStrategy) PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error {
	return nil
}

// UpdateWithOps builds the operations we want performed on the account
func (s mirrorStrategy) UpdateWithOps(
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) ([]build.TransactionMutator, error) {
	ob, e := s.tradeAPI.GetOrderBook(s.orderbookPair, s.config.ORDERBOOK_DEPTH)
	if e != nil {
		return nil, e
	}

	buyOps, e := s.updateLevels(
		buyingAOffers,
		ob.Bids(),
		s.sdex.ModifyBuyOffer,
		s.sdex.CreateBuyOffer,
		(1 - s.config.PER_LEVEL_SPREAD),
		true,
	)
	if e != nil {
		return nil, e
	}
	log.Printf("num. buyOps in this update: %d\n", len(buyOps))
	sellOps, e := s.updateLevels(
		sellingAOffers,
		ob.Asks(),
		s.sdex.ModifySellOffer,
		s.sdex.CreateSellOffer,
		(1 + s.config.PER_LEVEL_SPREAD),
		false,
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
) ([]build.TransactionMutator, error) {
	ops := []build.TransactionMutator{}
	var e error
	if len(newOrders) >= len(oldOffers) {
		for i := 0; i < len(oldOffers); i++ {
			ops, e = s.doModifyOffer(oldOffers[i], newOrders[i], priceMultiplier, s.config.VOLUME_DIVIDE_BY, modifyOffer, ops, hackPriceInvertForBuyOrderChangeCheck)
			if e != nil {
				return nil, e
			}
		}

		// create offers for remaining new bids
		for i := len(oldOffers); i < len(newOrders); i++ {
			price := model.NumberFromFloat(newOrders[i].Price.AsFloat()*priceMultiplier, utils.SdexPrecision).AsFloat()
			vol := model.NumberFromFloat(newOrders[i].Volume.AsFloat()/s.config.VOLUME_DIVIDE_BY, utils.SdexPrecision).AsFloat()
			incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(true)
			var mo *build.ManageOfferBuilder
			mo, e = createOffer(*s.baseAsset, *s.quoteAsset, price, vol, incrementalNativeAmountRaw)
			if e != nil {
				return nil, e
			}
			if mo != nil {
				ops = append(ops, *mo)
				// update the cached liabilities if we create a valid operation to create an offer
				s.sdex.AddLiabilities(*s.baseAsset, *s.quoteAsset, vol, vol*price, incrementalNativeAmountRaw)
			}
		}
	} else {
		for i := 0; i < len(newOrders); i++ {
			ops, e = s.doModifyOffer(oldOffers[i], newOrders[i], priceMultiplier, s.config.VOLUME_DIVIDE_BY, modifyOffer, ops, hackPriceInvertForBuyOrderChangeCheck)
			if e != nil {
				return nil, e
			}
		}

		// delete remaining prior offers
		for i := len(newOrders); i < len(oldOffers); i++ {
			op := s.sdex.DeleteOffer(oldOffers[i])
			ops = append(ops, op)
		}
	}
	return ops, nil
}

func (s *mirrorStrategy) doModifyOffer(
	oldOffer horizon.Offer,
	newOrder model.Order,
	priceMultiplier float64,
	volumeDivideBy float64,
	modifyOffer func(offer horizon.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*build.ManageOfferBuilder, error),
	ops []build.TransactionMutator,
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
) ([]build.TransactionMutator, error) {
	price := newOrder.Price.AsFloat() * priceMultiplier
	vol := newOrder.Volume.AsFloat() / volumeDivideBy

	oldPrice := model.MustNumberFromString(oldOffer.Price, 6)
	oldVol := model.MustNumberFromString(oldOffer.Amount, 6)
	if hackPriceInvertForBuyOrderChangeCheck {
		// we want to multiply oldVol by the original oldPrice so we can get the correct oldVol, since ModifyBuyOffer multiplies price * vol
		oldVol = model.NumberFromFloat(oldVol.AsFloat()*oldPrice.AsFloat(), 6)
		oldPrice = model.NumberFromFloat(1/oldPrice.AsFloat(), 6)
	}
	newPrice := model.NumberFromFloat(price, 6)
	newVol := model.NumberFromFloat(vol, 6)
	epsilon := 0.0001
	sameOrderParams := utils.FloatEquals(oldPrice.AsFloat(), newPrice.AsFloat(), epsilon) && utils.FloatEquals(oldVol.AsFloat(), newVol.AsFloat(), epsilon)
	if sameOrderParams {
		return ops, nil
	}

	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(false)
	offerPrice := model.NumberFromFloat(price, utils.SdexPrecision).AsFloat()
	offerAmount := model.NumberFromFloat(vol, utils.SdexPrecision).AsFloat()
	mo, e := modifyOffer(
		oldOffer,
		offerPrice,
		offerAmount,
		incrementalNativeAmountRaw,
	)
	if e != nil {
		return nil, e
	}
	if mo != nil {
		ops = append(ops, *mo)
		// update the cached liabilities if we create a valid operation to modify the offer
		s.sdex.AddLiabilities(oldOffer.Selling, oldOffer.Buying, offerAmount, offerAmount*offerPrice, incrementalNativeAmountRaw)
	}
	return ops, nil
}

// PostUpdate changes the strategy's state after the update has taken place
func (s *mirrorStrategy) PostUpdate() error {
	return nil
}
