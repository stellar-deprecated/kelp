package plugins

import (
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// balancedConfig contains the configuration params for this Strategy
type balancedConfig struct {
	PriceTolerance                float64 `valid:"-" toml:"PRICE_TOLERANCE"`
	AmountTolerance               float64 `valid:"-" toml:"AMOUNT_TOLERANCE"`
	Spread                        float64 `valid:"-" toml:"SPREAD"`                          // this is the bid-ask spread (i.e. it is not the spread from the center price)
	MinAmountSpread               float64 `valid:"-" toml:"MIN_AMOUNT_SPREAD"`               // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	MaxAmountSpread               float64 `valid:"-" toml:"MAX_AMOUNT_SPREAD"`               // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	MaxLevels                     int16   `valid:"-" toml:"MAX_LEVELS"`                      // max number of levels to have on either side
	LevelDensity                  float64 `valid:"-" toml:"LEVEL_DENSITY"`                   // value between 0.0 to 1.0 used as a probability
	EnsureFirstNLevels            int16   `valid:"-" toml:"ENSURE_FIRST_N_LEVELS"`           // always adds the first N levels, meaningless if LevelDensity = 1.0
	MinAmountCarryoverSpread      float64 `valid:"-" toml:"MIN_AMOUNT_CARRYOVER_SPREAD"`     // the minimum spread % we take off the amountCarryover before placing the orders
	MaxAmountCarryoverSpread      float64 `valid:"-" toml:"MAX_AMOUNT_CARRYOVER_SPREAD"`     // the maximum spread % we take off the amountCarryover before placing the orders
	CarryoverInclusionProbability float64 `valid:"-" toml:"CARRYOVER_INCLUSION_PROBABILITY"` // probability of including the carryover at a level that will be added
	VirtualBalanceBase            float64 `valid:"-" toml:"VIRTUAL_BALANCE_BASE"`            // virtual balance to use so we can smoothen out the curve
	VirtualBalanceQuote           float64 `valid:"-" toml:"VIRTUAL_BALANCE_QUOTE"`           // virtual balance to use so we can smoothen out the curve
}

// String impl.
func (c balancedConfig) String() string {
	return utils.StructString(c, 0, nil)
}

// makeBalancedStrategy is a factory method for balancedStrategy
func makeBalancedStrategy(
	sdex *SDEX,
	pair *model.TradingPair,
	ieif *IEIF,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	config *balancedConfig,
) api.Strategy {
	orderConstraints := sdex.GetOrderConstraints(pair)
	sellSideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetBase,
		assetQuote,
		makeBalancedLevelProvider(
			config.Spread,
			false,
			config.MinAmountSpread,
			config.MaxAmountSpread,
			config.MaxLevels,
			config.LevelDensity,
			config.EnsureFirstNLevels,
			config.MinAmountCarryoverSpread,
			config.MaxAmountCarryoverSpread,
			config.CarryoverInclusionProbability,
			config.VirtualBalanceBase,
			config.VirtualBalanceQuote,
			orderConstraints),
		config.PriceTolerance,
		config.AmountTolerance,
		false,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := makeSellSideStrategy(
		sdex,
		orderConstraints,
		ieif,
		assetQuote,
		assetBase,
		makeBalancedLevelProvider(
			config.Spread,
			true, // real base is passed in as quote so pass in true
			config.MinAmountSpread,
			config.MaxAmountSpread,
			config.MaxLevels,
			config.LevelDensity,
			config.EnsureFirstNLevels,
			config.MinAmountCarryoverSpread,
			config.MaxAmountCarryoverSpread,
			config.CarryoverInclusionProbability,
			config.VirtualBalanceQuote,
			config.VirtualBalanceBase,
			orderConstraints),
		config.PriceTolerance,
		config.AmountTolerance,
		true,
	)

	return makeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	)
}
