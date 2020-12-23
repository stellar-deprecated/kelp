package plugins

import (
	"fmt"
	"time"

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
	ExponentialSmoothingFactor                            float64               `valid:"-" toml:"EXPONENTIAL_SMOOTHING_FACTOR"`
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
	filterFactory *FilterFactory,
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
	dowFilter, e := makeDowFilter(filterFactory, config.DayOfWeekDailyCap)
	if e != nil {
		return nil, fmt.Errorf("error when making dowFilter: %s", e)
	}
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
		time.Now().UnixNano(),
		false,
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

func makeDowFilter(filterFactory *FilterFactory, dowDailyCap DayOfWeekFilterConfig) ([7]volumeFilter, error) {
	var dowVolumeFilters [7]volumeFilter
	var dowFilter [7]SubmitFilter
	var e error

	// time.Weekday begins with Sunday so we set the first value in the array to be Sunday
	dowFilter[0], e = filterFactory.MakeFilter(dowDailyCap.Su)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Sunday: %s", e)
	}

	dowFilter[1], e = filterFactory.MakeFilter(dowDailyCap.Mo)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Monday: %s", e)
	}

	dowFilter[2], e = filterFactory.MakeFilter(dowDailyCap.Tu)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Tuesday: %s", e)
	}

	dowFilter[3], e = filterFactory.MakeFilter(dowDailyCap.We)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Wednesday: %s", e)
	}

	dowFilter[4], e = filterFactory.MakeFilter(dowDailyCap.Th)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Thursday: %s", e)
	}

	dowFilter[5], e = filterFactory.MakeFilter(dowDailyCap.Fr)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Friday: %s", e)
	}

	dowFilter[6], e = filterFactory.MakeFilter(dowDailyCap.Sa)
	if e != nil {
		return dowVolumeFilters, fmt.Errorf("unable to make filter for entry Saturday: %s", e)
	}

	// enforce the filters to be of type volumeFilter
	for i, f := range dowFilter {
		vf, ok := f.(*volumeFilter)
		if !ok {
			return dowVolumeFilters, fmt.Errorf("could not cast %d-th filter to a volumeFilter", i)
		}
		dowVolumeFilters[i] = *vf
	}

	return dowVolumeFilters, nil
}
