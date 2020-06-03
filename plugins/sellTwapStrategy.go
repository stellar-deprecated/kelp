package plugins

import (
	"fmt"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// DayOfWeekFilterConfig is converted to a SubmitFilter and applied based on the current DOW
type DayOfWeekFilterConfig struct {
	Mo string `valid:"-" toml:"Mo"`
	Tu string `valid:"-" toml:"Tu"`
	We string `valid:"-" toml:"We"`
	Th string `valid:"-" toml:"Th"`
	Fr string `valid:"-" toml:"Fr"`
	Sa string `valid:"-" toml:"Sa"`
	Su string `valid:"-" toml:"Su"`
}

// sellTwapConfig contains the configuration params for this Strategy
type sellTwapConfig struct {
	StartAskFeedType       string  `valid:"-" toml:"START_ASK_FEED_TYPE"`
	StartAskFeedURL        string  `valid:"-" toml:"START_ASK_FEED_URL"`
	PriceTolerance         float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance        float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	RateOffsetPercent      float64 `valid:"-" toml:"RATE_OFFSET_PERCENT"`
	RateOffset             float64 `valid:"-" toml:"RATE_OFFSET"`
	RateOffsetPercentFirst bool    `valid:"-" toml:"RATE_OFFSET_PERCENT_FIRST"`
	// new params that are specific to the twap strategy
	DayOfWeekDailyCap                                     DayOfWeekFilterConfig `valid:"-" toml:"DAY_OF_WEEK_DAILY_CAP"`
	NumHoursToSell                                        int                   `valid:"-" toml:"NUM_HOURS_TO_SELL"`
	ParentBucketSizeSeconds                               int                   `valid:"-" toml:"PARENT_BUCKET_SIZE_SECONDS"`
	DistributeSurplusOverRemainingIntervalsPercentCeiling float64               `valid:"-" toml:"DISTRIBUTE_SURPLUS_OVER_REMAINING_INTERVALS_PERCENT_CEILING"`
	ExponentialSmoothingFactor                            float64               `valid:"-" toml:"EXPONENTIAL_SMOOTHING_FACTOR"` // a larger number results in a smoother distribution across the remaining intervals; 0 < x <= 1; set to 1.0 for a linear distribution and 0.0 to sell the entire surplus in the next interval
	MinChildOrderSizePercentOfParent                      float64               `valid:"-" toml:"MIN_CHILD_ORDER_SIZE_PERCENT_OF_PARENT"`
}

// String impl.
func (c sellTwapConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makeSellTwapStrategy is a factory method for SellTwapStrategy
func makeSellTwapStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	config *sellTwapConfig,
) (api.Strategy, error) {
	startPf, e := MakePriceFeed(config.StartAskFeedType, config.StartAskFeedURL)
	if e != nil {
		return nil, fmt.Errorf("error when making the start priceFeed: %s", e)
	}

	orderConstraints := sdex.GetOrderConstraints(pair)
	offset := rateOffset{
		percent:      config.RateOffsetPercent,
		absolute:     config.RateOffset,
		percentFirst: config.RateOffsetPercentFirst,
	}
	dowFilter := makeDowFilter(config.DayOfWeekDailyCap)
	levelProvider, e := makeSellTwapLevelProvider(
		startPf,
		offset,
		orderConstraints,
		dowFilter,
		config.NumHoursToSell,
		config.ParentBucketSizeSeconds,
		config.DistributeSurplusOverRemainingIntervalsPercentCeiling,
		config.ExponentialSmoothingFactor,
		config.MinChildOrderSizePercentOfParent,
	)
	if e != nil {
		return nil, fmt.Errorf("error when making a sellTwapLevelProvider: %s", e)
	}

	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,
		assetQuote,
		levelProvider,
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

func makeDowFilter(dowDailyCap DayOfWeekFilterConfig) map[string]SubmitFilter {
	// TODO
	return nil
}
