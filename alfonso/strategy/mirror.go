package strategy

import (
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/lightyeario/kelp/support/exchange/orderbook"

	"github.com/lightyeario/kelp/support"
	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

// MirrorConfig contains the configuration params for this strategy
type MirrorConfig struct {
	EXCHANGE         string  `valid:"-"`
	EXCHANGE_BASE    string  `valid:"-"`
	EXCHANGE_QUOTE   string  `valid:"-"`
	ORDERBOOK_DEPTH  int32   `valid:"-"`
	VOLUME_DIVIDE_BY float64 `valid:"-"`
	PER_LEVEL_SPREAD float64 `valid:"-"`
}

// MirrorStrategy is a strategy to mirror the orderbook of a given exchange
type MirrorStrategy struct {
	txButler      *kelp.TxButler
	orderbookPair *assets.TradingPair
	baseAsset     *horizon.Asset
	quoteAsset    *horizon.Asset
	config        *MirrorConfig
	exchange      exchange.Exchange

	// uninitialized
	maxAssetA float64
	maxAssetB float64
}

// ensure this implements Strategy
var _ Strategy = &MirrorStrategy{}

// MakeMirrorStrategy is a factory method
func MakeMirrorStrategy(txButler *kelp.TxButler, baseAsset *horizon.Asset, quoteAsset *horizon.Asset, config *MirrorConfig) Strategy {
	exchange := kelp.ExchangeFactory(config.EXCHANGE)
	orderbookPair := &assets.TradingPair{
		Base:  exchange.GetAssetConverter().MustFromString(config.EXCHANGE_BASE),
		Quote: exchange.GetAssetConverter().MustFromString(config.EXCHANGE_QUOTE),
	}
	return &MirrorStrategy{
		txButler:      txButler,
		orderbookPair: orderbookPair,
		baseAsset:     baseAsset,
		quoteAsset:    quoteAsset,
		config:        config,
		exchange:      exchange,
	}
}

// PruneExistingOffers deletes any extra offers
func (s MirrorStrategy) PruneExistingOffers(offers []horizon.Offer) []horizon.Offer {
	return offers
}

// PreUpdate changes the strategy's state in prepration for the update
func (s *MirrorStrategy) PreUpdate(
	maxAssetA float64,
	maxAssetB float64,
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) error {
	return nil
}

// UpdateWithOps builds the operations we want performed on the account
func (s MirrorStrategy) UpdateWithOps(
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) ([]build.TransactionMutator, error) {
	ob, e := s.exchange.GetOrderBook(s.orderbookPair, s.config.ORDERBOOK_DEPTH)
	if e != nil {
		return nil, e
	}

	buyOps := s.updateLevels(
		buyingAOffers,
		ob.Bids(),
		s.txButler.ModifyBuyOffer,
		s.txButler.CreateBuyOffer,
		(1 - s.config.PER_LEVEL_SPREAD),
		true,
	)
	log.Info("num. buyOps in this update: ", len(buyOps))
	sellOps := s.updateLevels(
		sellingAOffers,
		ob.Asks(),
		s.txButler.ModifySellOffer,
		s.txButler.CreateSellOffer,
		(1 + s.config.PER_LEVEL_SPREAD),
		false,
	)
	log.Info("num. sellOps in this update: ", len(sellOps))

	ops := []build.TransactionMutator{}
	if len(ob.Bids()) > 0 && len(sellingAOffers) > 0 && ob.Bids()[0].Price.AsFloat() >= kelp.PriceAsFloat(sellingAOffers[0].Price) {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	} else {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	}

	return ops, nil
}

func (s *MirrorStrategy) updateLevels(
	oldOffers []horizon.Offer,
	newOrders []orderbook.Order,
	modifyOffer func(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder,
	createOffer func(baseAsset horizon.Asset, quoteAsset horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder,
	priceMultiplier float64,
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
) []build.TransactionMutator {
	ops := []build.TransactionMutator{}
	if len(newOrders) >= len(oldOffers) {
		offset := len(newOrders) - len(oldOffers)
		for i := len(newOrders) - 1; (i - offset) >= 0; i-- {
			ops = doModifyOffer(oldOffers[i-offset], newOrders[i], priceMultiplier, s.config.VOLUME_DIVIDE_BY, modifyOffer, ops, hackPriceInvertForBuyOrderChangeCheck)
		}

		// create offers for remaining new bids
		for i := offset - 1; i >= 0; i-- {
			price := newOrders[i].Price.AsFloat() * priceMultiplier
			vol := newOrders[i].Volume.AsFloat() / s.config.VOLUME_DIVIDE_BY
			mo := createOffer(*s.baseAsset, *s.quoteAsset, price, vol)
			if mo != nil {
				ops = append(ops, *mo)
			}
		}
	} else {
		offset := len(oldOffers) - len(newOrders)
		for i := len(oldOffers) - 1; (i - offset) >= 0; i-- {
			ops = doModifyOffer(oldOffers[i], newOrders[i-offset], priceMultiplier, s.config.VOLUME_DIVIDE_BY, modifyOffer, ops, hackPriceInvertForBuyOrderChangeCheck)
		}

		// delete remaining prior offers
		for i := offset - 1; i >= 0; i-- {
			s.txButler.DeleteOffer(oldOffers[i])
		}
	}
	return ops
}

func doModifyOffer(
	oldOffer horizon.Offer,
	newOrder orderbook.Order,
	priceMultiplier float64,
	volumeDivideBy float64,
	modifyOffer func(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder,
	ops []build.TransactionMutator,
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
) []build.TransactionMutator {
	price := newOrder.Price.AsFloat() * priceMultiplier
	vol := newOrder.Volume.AsFloat() / volumeDivideBy

	oldPrice := number.MustFromString(oldOffer.Price, 6)
	oldVol := number.MustFromString(oldOffer.Amount, 6)
	if hackPriceInvertForBuyOrderChangeCheck {
		// we want to multiply oldVol by the original oldPrice so we can get the correct oldVol, since ModifyBuyOffer multiplies price * vol
		oldVol = number.FromFloat(oldVol.AsFloat()*oldPrice.AsFloat(), 6)
		oldPrice = number.FromFloat(1/oldPrice.AsFloat(), 6)
	}
	newPrice := number.FromFloat(price, 6)
	newVol := number.FromFloat(vol, 6)
	epsilon := 0.0001
	sameOrderParams := kelp.FloatEquals(oldPrice.AsFloat(), newPrice.AsFloat(), epsilon) && kelp.FloatEquals(oldVol.AsFloat(), newVol.AsFloat(), epsilon)
	//log.Info("oldPrice: ", oldPrice.AsString(), " | newPrice: ", newPrice.AsString(), " | oldVol: ", oldVol.AsString(), " | newVol: ", newVol.AsString(), " | sameOrderParams: ", sameOrderParams)
	if sameOrderParams {
		return ops
	}

	mo := modifyOffer(oldOffer, price, vol)
	if mo != nil {
		ops = append(ops, *mo)
	}
	return ops
}

// PostUpdate changes the strategy's state after the update has taken place
func (s *MirrorStrategy) PostUpdate() error {
	return nil
}
