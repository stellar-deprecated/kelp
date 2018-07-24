package strategy

import (
	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader/strategy/level"
	"github.com/lightyeario/kelp/trader/strategy/sideStrategy"
	"github.com/stellar/go/clients/horizon"
)

// SellConfig contains the configuration params for this Strategy
type SellConfig struct {
	DATA_TYPE_A            string              `valid:"-"`
	DATA_FEED_A_URL        string              `valid:"-"`
	DATA_TYPE_B            string              `valid:"-"`
	DATA_FEED_B_URL        string              `valid:"-"`
	PRICE_TOLERANCE        float64             `valid:"-"`
	AMOUNT_TOLERANCE       float64             `valid:"-"`
	AMOUNT_OF_A_BASE       float64             `valid:"-"` // the size of order
	DIVIDE_AMOUNT_BY_PRICE bool                `valid:"-"` // whether we want to divide the amount by the price, usually true if this is on the buy side
	LEVELS                 []level.StaticLevel `valid:"-"`
}

// MakeSellStrategy is a factory method for SellStrategy
func MakeSellStrategy(
	txButler *utils.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *SellConfig,
) api.Strategy {
	pf := plugins.MakeFeedPair(
		config.DATA_TYPE_A,
		config.DATA_FEED_A_URL,
		config.DATA_TYPE_B,
		config.DATA_FEED_B_URL,
	)
	sellSideStrategy := sideStrategy.MakeSellSideStrategy(
		txButler,
		assetBase,
		assetQuote,
		level.MakeStaticSpreadLevelProvider(config.LEVELS, config.AMOUNT_OF_A_BASE, pf),
		config.PRICE_TOLERANCE,
		config.AMOUNT_TOLERANCE,
		config.DIVIDE_AMOUNT_BY_PRICE,
	)
	// switch sides of base/quote here for the delete side
	deleteSideStrategy := sideStrategy.MakeDeleteSideStrategy(txButler, assetQuote, assetBase)

	return MakeComposeStrategy(
		assetBase,
		assetQuote,
		deleteSideStrategy,
		sellSideStrategy,
	)
}
