package strategy

import (
	"github.com/lightyeario/kelp/support/utils"
	"github.com/lightyeario/kelp/trader/strategy/sideStrategy"
	"github.com/stellar/go/clients/horizon"
)

// MakeDeleteStrategy is a factory method
func MakeDeleteStrategy(
	txButler *utils.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
) Strategy {
	return MakeComposeStrategy(
		assetBase,
		assetQuote,
		sideStrategy.MakeDeleteSideStrategy(txButler, assetQuote, assetBase), // switch sides of base/quote here for the buy side
		sideStrategy.MakeDeleteSideStrategy(txButler, assetBase, assetQuote),
	)
}
