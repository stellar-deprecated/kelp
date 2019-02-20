package plugins

import (
	"fmt"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// buySellConfig contains the configuration params for this strategy
type buySellConfig struct {
	PriceTolerance         float64       `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance        float64       `valid:"-" toml:"AMOUNT_TOLERANCE"`
	RateOffsetPercent      float64       `valid:"-" toml:"RATE_OFFSET_PERCENT"`
	RateOffset             float64       `valid:"-" toml:"RATE_OFFSET"`
	RateOffsetPercentFirst bool          `valid:"-" toml:"RATE_OFFSET_PERCENT_FIRST"`
	AmountOfABase          float64       `valid:"-" toml:"AMOUNT_OF_A_BASE"` // the size of order to keep on either side
	DataTypeA              string        `valid:"-" toml:"DATA_TYPE_A"`
	DataFeedAURL           string        `valid:"-" toml:"DATA_FEED_A_URL"`
	DataTypeB              string        `valid:"-" toml:"DATA_TYPE_B"`
	DataFeedBURL           string        `valid:"-" toml:"DATA_FEED_B_URL"`
	Levels                 []staticLevel `valid:"-" toml:"LEVELS"`
}

// String impl.
func (c buySellConfig) String() string {
	return utils.StructString(c, nil)
}

// makeBuySellStrategy is a factory method
func makeBuySellStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	ieif *IEIF,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *buySellConfig,
) (api.Strategy, error) {
	offsetSell := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
	}
	sellSideFeedPair, e := MakeFeedPair(
		config.DataTypeA,
		config.DataFeedAURL,
		config.DataTypeB,
		config.DataFeedBURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the buysell strategy because we could not make the sell side feed pair: %s", e)
	}
	orderConstraints := sdex.GetOrderConstraints(pair)
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,
		assetQuote,
		makeStaticSpreadLevelProvider(
			config.Levels,
			config.AmountOfABase,
			offsetSell,
			sellSideFeedPair,
			orderConstraints,
		),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)

	offsetBuy := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
		invert:       true,
	}
	buySideFeedPair, e := MakeFeedPair(
		config.DataTypeB,
		config.DataFeedBURL,
		config.DataTypeA,
		config.DataFeedAURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the buysell strategy because we could not make the buy side feed pair: %s", e)
	}
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetQuote,
		assetBase,
		makeStaticSpreadLevelProvider(
			config.Levels,
			config.AmountOfABase,
			offsetBuy,
			buySideFeedPair,
			orderConstraints,
		),
		config.PriceTolerance,
		config.AmountTolerance,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	), nil
}
