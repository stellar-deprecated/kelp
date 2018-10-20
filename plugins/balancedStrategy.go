package plugins

import (
	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
)

// balancedConfig contains the configuration params for this Strategy
type balancedConfig struct {
	priceTolerance                float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	amountTolerance               float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	spread                        float64 `valid:"-" toml:"SPREAD"`                          // this is the bid-ask spread (i.e. it is not the spread from the center price)
	minAmountSpread               float64 `valid:"-" toml:"MIN_AMOUNT_SPREAD"`               // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	maxAmountSpread               float64 `valid:"-" toml:"MAX_AMOUNT_SPREAD"`               // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	maxLevels                     int16   `valid:"-" toml:"MAX_LEVELS"`                      // max number of levels to have on either side
	levelDensity                  float64 `valid:"-" toml:"LEVEL_DENSITY"`                   // value between 0.0 to 1.0 used as a probability
	ensureFirstNLevels            int16   `valid:"-" toml:"ENSURE_FIRST_N_LEVELS"`           // always adds the first N levels, meaningless if levelDensity = 1.0
	minAmountCarryoverSpread      float64 `valid:"-" toml:"MIN_AMOUNT_CARRYOVER_SPREAD"`     // the minimum spread % we take off the amountCarryover before placing the orders
	maxAmountCarryoverSpread      float64 `valid:"-" toml:"MAX_AMOUNT_CARRYOVER_SPREAD"`     // the maximum spread % we take off the amountCarryover before placing the orders
	carryoverInclusionProbability float64 `valid:"-" toml:"CARRYOVER_INCLUSION_PROBABILITY"` // probability of including the carryover at a level that will be added
	virtualBalanceBase            float64 `valid:"-" toml:"VIRTUAL_BALANCE_BASE"`            // virtual balance to use so we can smoothen out the curve
	virtualBalanceQuote           float64 `valid:"-" toml:"VIRTUAL_BALANCE_QUOTE"`           // virtual balance to use so we can smoothen out the curve
}

// String impl.
func (c balancedConfig) String() string {
	return utils.StructString(c, nil)
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
			config.spread,
			false,
			config.minAmountSpread,
			config.maxAmountSpread,
			config.maxLevels,
			config.levelDensity,
			config.ensureFirstNLevels,
			config.minAmountCarryoverSpread,
			config.maxAmountCarryoverSpread,
			config.carryoverInclusionProbability,
			config.virtualBalanceBase,
			config.virtualBalanceQuote),
		config.priceTolerance,
		config.amountTolerance,
		false,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		assetQuote,
		assetBase,
		makeBalancedLevelProvider(
			config.spread,
			true, // real base is passed in as quote so pass in true
			config.minAmountSpread,
			config.maxAmountSpread,
			config.maxLevels,
			config.levelDensity,
			config.ensureFirstNLevels,
			config.minAmountCarryoverSpread,
			config.maxAmountCarryoverSpread,
			config.carryoverInclusionProbability,
			config.virtualBalanceQuote,
			config.virtualBalanceBase),
		config.priceTolerance,
		config.amountTolerance,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	)
}
