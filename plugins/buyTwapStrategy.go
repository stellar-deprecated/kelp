package plugins

import (
	"fmt"
	"time"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// makeBuyTwapStrategy is a factory method for BuyTwapStrategy
func makeBuyTwapStrategy(
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
		true,
	)
	if e != nil {
		return nil, fmt.Errorf("error when making a sellTwapLevelProvider: %s", e)
	}

	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetQuote,
		assetBase,
		levelProvider,
		config.PriceTolerance,
		config.AmountTolerance,
		true,
	)

	// use assetBase as param to assetBase argument, since the delete strategy is
	// on the sell side. so the params should line up correctly with the arguments
	deleteSideStrategy := makeDeleteSideStrategy(sdex, assetBase, assetQuote)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		deleteSideStrategy,
	), nil

}
