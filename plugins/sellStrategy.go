package plugins

import (
	"fmt"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// sellConfig contains the configuration params for this Strategy
type sellConfig struct {
	DataTypeA              string        `valid:"-" toml:"DATA_TYPE_A"`
	DataFeedAURL           string        `valid:"-" toml:"DATA_FEED_A_URL"`
	DataTypeB              string        `valid:"-" toml:"DATA_TYPE_B"`
	DataFeedBURL           string        `valid:"-" toml:"DATA_FEED_B_URL"`
	PriceTolerance         float64       `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance        float64       `valid:"-" toml:"AMOUNT_TOLERANCE"`
	AmountOfABase          float64       `valid:"-" toml:"AMOUNT_OF_A_BASE"` // the size of order
	RateOffsetPercent      float64       `valid:"-" toml:"RATE_OFFSET_PERCENT"`
	RateOffset             float64       `valid:"-" toml:"RATE_OFFSET"`
	RateOffsetPercentFirst bool          `valid:"-" toml:"RATE_OFFSET_PERCENT_FIRST"`
	Levels                 []staticLevel `valid:"-" toml:"LEVELS"`
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
) (api.Strategy, error) {
	pf, e := MakeFeedPair(
		config.DataTypeA,
		config.DataFeedAURL,
		config.DataTypeB,
		config.DataFeedBURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the sell strategy because we could not make the feed pair: %s", e)
	}

	offset := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
	}
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		assetBase,
		assetQuote,
		makeStaticSpreadLevelProvider(config.Levels, config.AmountOfABase, offset, pf),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)
	// switch sides of base/quote here for the delete side
	deleteSideStrategy := makeDeleteSideStrategy(sdex, assetQuote, assetBase)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		deleteSideStrategy,
		sellSideStrategy,
	), nil
}
