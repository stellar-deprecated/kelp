package plugins

import (
	"fmt"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// buySellConfig contains the configuration params for this strategy
type buySellConfig struct {
	PRICE_TOLERANCE           float64       `valid:"-"`
	AMOUNT_TOLERANCE          float64       `valid:"-"`
	RATE_OFFSET_PERCENT       float64       `valid:"-"`
	RATE_OFFSET               float64       `valid:"-"`
	RATE_OFFSET_PERCENT_FIRST bool          `valid:"-"`
	AMOUNT_OF_A_BASE          float64       `valid:"-"` // the size of order to keep on either side
	DATA_TYPE_A               string        `valid:"-"`
	DATA_FEED_A_URL           string        `valid:"-"`
	DATA_TYPE_B               string        `valid:"-"`
	DATA_FEED_B_URL           string        `valid:"-"`
	LEVELS                    []staticLevel `valid:"-"`
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
		percent:      config.RATE_OFFSET_PERCENT,
		absolute:     config.RATE_OFFSET,
		percentFirst: config.RATE_OFFSET_PERCENT_FIRST,
	}
	sellSideFeedPair, e := MakeFeedPair(
		config.DATA_TYPE_A,
		config.DATA_FEED_A_URL,
		config.DATA_TYPE_B,
		config.DATA_FEED_B_URL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the buysell strategy because we could not make the sell side feed pair: %s", e)
	}
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		assetBase,
		assetQuote,
		makeStaticSpreadLevelProvider(
			config.LEVELS,
			config.AMOUNT_OF_A_BASE,
			offsetSell,
			sellSideFeedPair,
		),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		false,
	)

	offsetBuy := rateOffset{
		percent:      config.RATE_OFFSET_PERCENT,
		absolute:     config.RATE_OFFSET,
		percentFirst: config.RATE_OFFSET_PERCENT_FIRST,
		invert:       true,
	}
	buySideFeedPair, e := MakeFeedPair(
		config.DATA_TYPE_B,
		config.DATA_FEED_B_URL,
		config.DATA_TYPE_A,
		config.DATA_FEED_A_URL,
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
			config.LEVELS,
			config.AMOUNT_OF_A_BASE,
			offsetBuy,
			buySideFeedPair,
		),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	), nil
}
