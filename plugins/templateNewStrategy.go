package plugins

import (
	"fmt"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// templateNewConfig contains the configuration params for this templateNewStrategy
type templateNewConfig struct {
	// parameters specific to this template new trading strategy
	FeedTypeA string `valid:"-" toml:"FEED_TYPE_A"`
	FeedURLA  string `valid:"-" toml:"FEED_URL_A"`
	FeedTypeB string `valid:"-" toml:"FEED_TYPE_B"`
	FeedURLB  string `valid:"-" toml:"FEED_URL_B"`
	FeedTypeC string `valid:"-" toml:"FEED_TYPE_C"`
	FeedURLC  string `valid:"-" toml:"FEED_URL_C"`
	// common parameters below
	PriceTolerance         float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance        float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	RateOffsetPercent      float64 `valid:"-" toml:"RATE_OFFSET_PERCENT"`
	RateOffset             float64 `valid:"-" toml:"RATE_OFFSET"`
	RateOffsetPercentFirst bool    `valid:"-" toml:"RATE_OFFSET_PERCENT_FIRST"`
}

// String impl.
func (c templateNewConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makeTemplateNewStrategy is a factory method for SellTwapStrategy
func makeTemplateNewStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	config *templateNewConfig,
) (api.Strategy, error) {
	// instantiate the three price feeds using the strings from the configs
	priceFeedA, e := MakePriceFeed(config.FeedTypeA, config.FeedURLA)
	if e != nil {
		return nil, fmt.Errorf("error when making the feed A: %s", e)
	}
	priceFeedB, e := MakePriceFeed(config.FeedTypeB, config.FeedURLB)
	if e != nil {
		return nil, fmt.Errorf("error when making the feed B: %s", e)
	}
	priceFeedC, e := MakePriceFeed(config.FeedTypeC, config.FeedURLC)
	if e != nil {
		return nil, fmt.Errorf("error when making the feed C: %s", e)
	}
	// build the remaining dependencies that are used to build the templateNewLevelProvider
	orderConstraints := sdex.GetOrderConstraints(pair)
	offset := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
	}
	// build the templateNewLevelProvider, to be used as the sell side
	sellLevelProvider, e := makeTemplateNewLevelProvider(
		priceFeedA,
		priceFeedB,
		priceFeedC,
		offset,
		orderConstraints,
	)
	if e != nil {
		return nil, fmt.Errorf("error when making a templateNewLevelProvider: %s", e)
	}

	// since we are using the levelProvider framework, we need to inject our levelProvider into the sellSideStrategy.
	// This also works on the buy side with the same levelProvider implementation (separate instance if it holds state), see comments below.
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,  // pass in the base asset as an argument to the baseAsset parameter
		assetQuote, // pass in the quote asset as an argument to the quoteAsset parameter
		sellLevelProvider,
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)

	// make a delete strategy as an example for a one-sided strategy
	// if you are looking for an example of a strategy that uses both sides with the same levelProvider, look at buysellStrategy.go, only the initialization of the buy side needs to change
	// All side strategies are written in the "context" of the sell side for simplicity which allows you to return prices in increasing order for deeper levels. The downside of this is that
	// the Kelp strategy framework requires you to switch sides of base/quote here for the buy side.
	buySideStrategy := makeDeleteSideStrategy(
		sdex,
		assetQuote, // pass in the quote asset as an argument to the baseAsset parameter
		assetBase,  // pass in the base asset as an argument to the quoteAsset parameter
	)

	// always pass the base asset as base to the compose strategy. Here we are combining the two side strategies into one single strategy
	return makeComposeStrategy(
		assetBase,        // always the base asset
		assetQuote,       // always the quote asset
		buySideStrategy,  // buy side
		sellSideStrategy, // sell side
	), nil

	// Note: it is more complicated if you do not use the levelProvider framework. The mirror strategy is an example of this.
	// The reason it is more complex is because you have to worry about a lot of details such as placing offer amounts for insufficient funds,
	// handling both the buy side and sell side in your logic, etc. All of this and more is taken care of for you when using the level provider
	// framework. All new strategies created in Kelp typically use the level provider framework, as demonstrated above.
}
