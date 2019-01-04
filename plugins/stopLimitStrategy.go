package plugins

import (
	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// stopLimitConfig contains the configuration params for this Strategy
type stopLimitConfig struct {
	AmountOfA       float64 `valid:"-" toml:"AMOUNT_OF_A"` // the size of order
	StopPrice       float64 `valid:"-" toml:"STOP_PRICE"`  //the price the bid must fall below to trigger the order
	LimitPrice      float64 `valid:"-" toml:"LIMIT_PRICE"` //the price at which the order is placed
	PriceTolerance  float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
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
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		assetBase,
		assetQuote,
		makeStopLimitLevelProvider(
			config.AmountOfA,
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
		deleteSideStrategy,
		sellSideStrategy,
	), nil
}
