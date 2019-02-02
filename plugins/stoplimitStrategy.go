package plugins

import (
	"fmt"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// stopLimitConfig contains the configuration params for this Strategy
type stopLimitConfig struct {
	AmountOfBase    float64 `valid:"-" toml:"AMOUNT_OF_BASE"` // the size of the sell order
	StopPrice       float64 `valid:"-" toml:"STOP_PRICE"`     //the price the bid must fall below to trigger the order
	LimitPrice      float64 `valid:"-" toml:"LIMIT_PRICE"`    //the price at which the order is placed
	PriceTolerance  float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	DataTypeA       string  `valid:"-" toml:"DATA_TYPE_A"`
	DataFeedAURL    string  `valid:"-" toml:"DATA_FEED_A_URL"`
	DataTypeB       string  `valid:"-" toml:"DATA_TYPE_B"`
	DataFeedBURL    string  `valid:"-" toml:"DATA_FEED_B_URL"`
}

// String impl.
func (c stopLimitConfig) String() string {
	return utils.StructString(c, nil)
}

// makeStopLimitStrategy is a factory method for StopLimitStrategy
func makeStopLimitStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *stopLimitConfig,
) (api.Strategy, error) {
	orderConstraints := sdex.GetOrderConstraints(pair)
	pf, e := MakeFeedPair(
		config.DataTypeA,
		config.DataFeedAURL,
		config.DataTypeB,
		config.DataFeedBURL,
	)
	if e != nil {
		return nil, fmt.Errorf("cannot make the stoplimit strategy because we could not make the feed pair: %s", e)
	}
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		assetBase,
		assetQuote,

		makeStopLimitLevelProvider(
			pf,
			assetBase,
			assetQuote,
			config.AmountOfBase,
			config.StopPrice,
			config.LimitPrice,
			orderConstraints),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)
	// we're just waiting to sell so nothing on the other side
	deleteSideStrategy := makeDeleteSideStrategy(sdex, assetQuote, assetBase)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		sellSideStrategy,
		deleteSideStrategy,
	), nil
}
