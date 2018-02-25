package strategy

import (
	"fmt"
	"log"
	"strconv"

	"github.com/lightyeario/kelp/alfonso/priceFeed"
	kelp "github.com/lightyeario/kelp/support"
	"github.com/lightyeario/kelp/support/exchange/assets"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// SellConfig contains the configuration params for this strategy
type SellConfig struct {
	EXCHANGE         string  `valid:"-"`
	EXCHANGE_BASE    string  `valid:"-"`
	EXCHANGE_QUOTE   string  `valid:"-"`
	USE_BID_PRICE    string  `valid:"-"`
	PRICE_TOLERANCE  float64 `valid:"-"`
	AMOUNT_TOLERANCE float64 `valid:"-"`
	AMOUNT_OF_A_BASE float64 `valid:"-"` // the size of order
	LEVELS           []Level `valid:"-"`
}

// SellStrategy is a strategy to sell a specific currency on SDEX by reading prices form an exchange
type SellStrategy struct {
	txButler   *kelp.TxButler
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
	config     *SellConfig
	pf         priceFeed.FeedPair

	// uninitialized
	centerPrice           float64
	lastPlacedCenterPrice *float64
	maxAssetBase          float64
	maxAssetQuote         float64
}

// ensure it implements Strategy
var _ Strategy = &SellStrategy{}

// MakeSellStrategy is a factory method for SellStrategy
func MakeSellStrategy(
	txButler *kelp.TxButler,
	exchange string,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *SellConfig,
) Strategy {
	// check valid config codes
	assetOrderMatched := config.EXCHANGE_BASE == assetBase.Code && config.EXCHANGE_QUOTE == assetQuote.Code
	assetOrderReversed := config.EXCHANGE_BASE == assetQuote.Code && config.EXCHANGE_QUOTE == assetBase.Code
	if !assetOrderMatched && !assetOrderReversed {
		log.Panic("strategy's config does not have the same base/quote as is specified in the bot config")
	}

	// convert to exchange's codes
	exchangeAssetConverter := kelp.ExchangeFactory(config.EXCHANGE).GetAssetConverter()
	baseExchangeAssetStr, e := exchangeAssetConverter.ToString(assets.Display.MustFromString(config.EXCHANGE_BASE))
	if e != nil {
		log.Panic("could not fetch exchange's asset code for string: ", config.EXCHANGE_BASE, e)
	}
	quoteExchangeAssetStr, e := exchangeAssetConverter.ToString(assets.Display.MustFromString(config.EXCHANGE_QUOTE))
	if e != nil {
		log.Panic("could not fetch exchange's asset code for string: ", config.EXCHANGE_QUOTE, e)
	}

	useBidPrice, e := strconv.ParseBool(config.USE_BID_PRICE)
	if e != nil {
		log.Panic("could not parse USE_BID_PRICE as a bool value: ", config.USE_BID_PRICE)
	}

	// build
	exchangeFeedPairURL := fmt.Sprintf("%s/%s/%s/%v", config.EXCHANGE, baseExchangeAssetStr, quoteExchangeAssetStr, useBidPrice)
	return &SellStrategy{
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
func (s *SellStrategy) PruneExistingOffers(offers []horizon.Offer) []horizon.Offer {
	return nil
}

// PreUpdate impl
func (s *SellStrategy) PreUpdate(
	maxAssetA float64,
	maxAssetB float64,
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) error {
	return nil
}

// UpdateWithOps impl
func (s *SellStrategy) UpdateWithOps(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, error) {
	return nil, nil
}

// PostUpdate impl
func (s *SellStrategy) PostUpdate() error {
	return nil
}
