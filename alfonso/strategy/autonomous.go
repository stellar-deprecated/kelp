package strategy

import (
	"github.com/lightyeario/kelp/alfonso/strategy/level"
	"github.com/lightyeario/kelp/alfonso/strategy/sideStrategy"
	"github.com/lightyeario/kelp/support"
	"github.com/stellar/go/clients/horizon"
)

// AutonomousConfig contains the configuration params for this Strategy
type AutonomousConfig struct {
	DATA_TYPE_A                  string  `valid:"-"`
	DATA_FEED_A_URL              string  `valid:"-"`
	DATA_TYPE_B                  string  `valid:"-"`
	DATA_FEED_B_URL              string  `valid:"-"`
	PRICE_TOLERANCE              float64 `valid:"-"`
	AMOUNT_TOLERANCE             float64 `valid:"-"`
	SPREAD                       float64 `valid:"-"` // this is the bid-ask spread (i.e. it is not the spread from the center price)
	PLATEAU_THRESHOLD_PERCENTAGE float64 `valid:"-"`
	AMOUNT_SPREAD                float64 `valid:"-"` // reduces the order size by this percentage resulting in a gain anytime 1 unit more than the first layer is consumed
	LEVELS                       int8    `valid:"-"` // number of levels to have on either side
}

// MakeAutonomousStrategy is a factory method for AutonomousStrategy
func MakeAutonomousStrategy(
	txButler *kelp.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *AutonomousConfig,
) Strategy {
	sellSideStrategy := sideStrategy.MakeSellSideStrategy(
		txButler,
		assetBase,
		assetQuote,
		level.MakeAutonomousLevelProvider(config.SPREAD, config.PLATEAU_THRESHOLD_PERCENTAGE, false, config.AMOUNT_SPREAD, config.LEVELS),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		false,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := sideStrategy.MakeSellSideStrategy(
		txButler,
		assetQuote,
		assetBase,
		level.MakeAutonomousLevelProvider(config.SPREAD, config.PLATEAU_THRESHOLD_PERCENTAGE, true, config.AMOUNT_SPREAD, config.LEVELS), // real base is passed in as quote so pass in true
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		true,
	)

	return MakeComposeStrategy(
		assetBase,
		assetQuote,
		buySideStrategy,
		sellSideStrategy,
	)
}
