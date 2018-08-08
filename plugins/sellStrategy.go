package plugins

import (
	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// sellConfig contains the configuration params for this Strategy
type sellConfig struct {
	DATA_TYPE_A               string        `valid:"-"`
	DATA_FEED_A_URL           string        `valid:"-"`
	DATA_TYPE_B               string        `valid:"-"`
	DATA_FEED_B_URL           string        `valid:"-"`
	PRICE_TOLERANCE           float64       `valid:"-"`
	AMOUNT_TOLERANCE          float64       `valid:"-"`
	AMOUNT_OF_A_BASE          float64       `valid:"-"` // the size of order
	RATE_OFFSET_PERCENT       float64       `valid:"-"`
	RATE_OFFSET               float64       `valid:"-"`
	RATE_OFFSET_PERCENT_FIRST bool          `valid:"-"`
	LEVELS                    []staticLevel `valid:"-"`
}

// String impl.
func (c sellConfig) String() string {
	return utils.StructString(c, nil)
}

// makeSellStrategy is a factory method for SellStrategy
func makeSellStrategy(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *sellConfig,
) api.Strategy {
	pf := MakeFeedPair(
		config.DATA_TYPE_A,
		config.DATA_FEED_A_URL,
		config.DATA_TYPE_B,
		config.DATA_FEED_B_URL,
	)
	offset := rateOffset{
		percent:      config.RATE_OFFSET_PERCENT,
		absolute:     config.RATE_OFFSET,
		percentFirst: config.RATE_OFFSET_PERCENT_FIRST,
	}
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		assetBase,
		assetQuote,
		makeStaticSpreadLevelProvider(config.LEVELS, config.AMOUNT_OF_A_BASE, offset, pf),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		false,
	)
	// switch sides of base/quote here for the delete side
	deleteSideStrategy := makeDeleteSideStrategy(sdex, assetQuote, assetBase)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		deleteSideStrategy,
		sellSideStrategy,
	)
}
