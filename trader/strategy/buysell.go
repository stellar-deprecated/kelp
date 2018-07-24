package strategy

import (
	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader/strategy/level"
	"github.com/lightyeario/kelp/trader/strategy/sideStrategy"
	"github.com/stellar/go/clients/horizon"
)

// BuySellConfig contains the configuration params for this strategy
type BuySellConfig struct {
	PRICE_TOLERANCE  float64             `valid:"-"`
	AMOUNT_TOLERANCE float64             `valid:"-"`
	AMOUNT_OF_A_BASE float64             `valid:"-"` // the size of order to keep on either side
	DATA_TYPE_A      string              `valid:"-"`
	DATA_FEED_A_URL  string              `valid:"-"`
	DATA_TYPE_B      string              `valid:"-"`
	DATA_FEED_B_URL  string              `valid:"-"`
	LEVELS           []level.StaticLevel `valid:"-"`
}

// MakeBuySellStrategy is a factory method
func MakeBuySellStrategy(
	txButler *utils.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *BuySellConfig,
) api.Strategy {
	sellSideStrategy := sideStrategy.MakeSellSideStrategy(
		txButler,
		assetBase,
		assetQuote,
		level.MakeStaticSpreadLevelProvider(
			config.LEVELS,
			config.AMOUNT_OF_A_BASE,
			plugins.MakeFeedPair(
				config.DATA_TYPE_A,
				config.DATA_FEED_A_URL,
				config.DATA_TYPE_B,
				config.DATA_FEED_B_URL,
			),
		),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		false,
	)
	// switch sides of base/quote here for buy side
	buySideStrategy := sideStrategy.MakeSellSideStrategy(
		txButler,
		assetQuote,
		assetBase,
		level.MakeStaticSpreadLevelProvider(
			config.LEVELS,
			config.AMOUNT_OF_A_BASE,
			plugins.MakeFeedPair(
				config.DATA_TYPE_B,
				config.DATA_FEED_B_URL,
				config.DATA_TYPE_A,
				config.DATA_FEED_A_URL,
			),
		),
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
