package plugins

import (
	"fmt"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
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
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		assetBase,
		assetQuote,
		makeStaticSpreadLevelProvider(
			config.Levels,
			config.AmountOfABase,
			offsetSell,
			sellSideFeedPair,
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
		assetQuote,
		assetBase,
		makeStaticSpreadLevelProvider(
			config.Levels,
			config.AmountOfABase,
			offsetBuy,
			buySideFeedPair,
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
