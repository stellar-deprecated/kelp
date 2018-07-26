package plugins

import (
	"github.com/lightyeario/kelp/api"
	"github.com/stellar/go/clients/horizon"
)

// balancedConfig contains the configuration params for this Strategy
type balancedConfig struct {
	PRICE_TOLERANCE                 float64 `valid:"-"`
	AMOUNT_TOLERANCE                float64 `valid:"-"`
	SPREAD                          float64 `valid:"-"` // this is the bid-ask spread (i.e. it is not the spread from the center price)
	MIN_AMOUNT_SPREAD               float64 `valid:"-"` // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	MAX_AMOUNT_SPREAD               float64 `valid:"-"` // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	MAX_LEVELS                      int16   `valid:"-"` // max number of levels to have on either side
	LEVEL_DENSITY                   float64 `valid:"-"` // value between 0.0 to 1.0 used as a probability
	ENSURE_FIRST_N_LEVELS           int16   `valid:"-"` // always adds the first N levels, meaningless if levelDensity = 1.0
	MIN_AMOUNT_CARRYOVER_SPREAD     float64 `valid:"-"` // the minimum spread % we take off the amountCarryover before placing the orders
	MAX_AMOUNT_CARRYOVER_SPREAD     float64 `valid:"-"` // the maximum spread % we take off the amountCarryover before placing the orders
	CARRYOVER_INCLUSION_PROBABILITY float64 `valid:"-"` // probability of including the carryover at a level that will be added
	VIRTUAL_BALANCE_BASE            float64 `valid:"-"` // virtual balance to use so we can smoothen out the curve
	VIRTUAL_BALANCE_QUOTE           float64 `valid:"-"` // virtual balance to use so we can smoothen out the curve
}

// makeBalancedStrategy is a factory method for balancedStrategy
func makeBalancedStrategy(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *balancedConfig,
) api.Strategy {
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		assetBase,
		assetQuote,
		makeBalancedLevelProvider(
			config.SPREAD,
			false,
			config.MIN_AMOUNT_SPREAD,
			config.MAX_AMOUNT_SPREAD,
			config.MAX_LEVELS,
			config.LEVEL_DENSITY,
			config.ENSURE_FIRST_N_LEVELS,
			config.MIN_AMOUNT_CARRYOVER_SPREAD,
			config.MAX_AMOUNT_CARRYOVER_SPREAD,
			config.CARRYOVER_INCLUSION_PROBABILITY,
			config.VIRTUAL_BALANCE_BASE,
			config.VIRTUAL_BALANCE_QUOTE),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		false,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		assetQuote,
		assetBase,
		makeBalancedLevelProvider(
			config.SPREAD,
			true, // real base is passed in as quote so pass in true
			config.MIN_AMOUNT_SPREAD,
			config.MAX_AMOUNT_SPREAD,
			config.MAX_LEVELS,
			config.LEVEL_DENSITY,
			config.ENSURE_FIRST_N_LEVELS,
			config.MIN_AMOUNT_CARRYOVER_SPREAD,
			config.MAX_AMOUNT_CARRYOVER_SPREAD,
			config.CARRYOVER_INCLUSION_PROBABILITY,
			config.VIRTUAL_BALANCE_QUOTE,
			config.VIRTUAL_BALANCE_BASE),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	)
}
