package strategy

import (
	"github.com/lightyeario/kelp/support/exchange/orderbook"

	"github.com/lightyeario/kelp/support"
	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/lightyeario/kelp/support/kraken"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// MirrorConfig contains the configuration params for this strategy
type MirrorConfig struct {
	EXCHANGE         string  `valid:"-"`
	EXCHANGE_BASE    string  `valid:"-"`
	EXCHANGE_QUOTE   string  `valid:"-"`
	ORDERBOOK_DEPTH  int32   `valid:"-"`
	VOLUME_DIVIDE_BY float64 `valid:"-"`
	// TODO 2 need to account for these tolerances in the strategy impl.
	PRICE_TOLERANCE  float64 `valid:"-"`
	AMOUNT_TOLERANCE float64 `valid:"-"`
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
	var exchange exchange.Exchange
	switch config.EXCHANGE {
	case "kraken":
		exchange = kraken.MakeKrakenExchange()
	}

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

	// TODO 2 confirm Bids is correct
	buyOps := s.updateLevels(buyingAOffers, ob.Bids(), s.txButler.ModifyBuyOffer, s.txButler.CreateBuyOffer)
	// TODO 2 confirm Asks is correct
	sellOps := s.updateLevels(sellingAOffers, ob.Asks(), s.txButler.ModifySellOffer, s.txButler.CreateSellOffer)

	// TODO 2 confirm this should be Bids here (intention is for it to be the new buy orders)
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
) []build.TransactionMutator {
	ops := []build.TransactionMutator{}
	if len(newOrders) >= len(oldOffers) {
		offset := len(newOrders) - len(oldOffers)
		for i := len(newOrders) - 1; (i - offset) >= 0; i-- {
			vol := newOrders[i].Volume.AsFloat() / s.config.VOLUME_DIVIDE_BY
			mo := modifyOffer(oldOffers[i-offset], newOrders[i].Price.AsFloat(), vol)
			if mo != nil {
				ops = append(ops, *mo)
			}
		}

		// create offers for remaining new bids
		for i := offset - 1; i >= 0; i-- {
			vol := newOrders[i].Volume.AsFloat() / s.config.VOLUME_DIVIDE_BY
			mo := createOffer(*s.baseAsset, *s.quoteAsset, newOrders[i].Price.AsFloat(), vol)
			if mo != nil {
				ops = append(ops, *mo)
			}
		}
	} else {
		offset := len(oldOffers) - len(newOrders)
		for i := len(oldOffers) - 1; (i - offset) >= 0; i-- {
			vol := newOrders[i-offset].Volume.AsFloat() / s.config.VOLUME_DIVIDE_BY
			mo := modifyOffer(oldOffers[i], newOrders[i-offset].Price.AsFloat(), vol)
			if mo != nil {
				ops = append(ops, *mo)
			}
		}

		// delete remaining prior offers
		for i := offset - 1; i >= 0; i-- {
			s.txButler.DeleteOffer(oldOffers[i])
		}
	}
	return ops
}

// PostUpdate changes the strategy's state after the update has taken place
func (s *MirrorStrategy) PostUpdate() error {
	return nil
}
