package strategy

import (
	"github.com/lightyeario/kelp/alfonso/strategy/sideStrategy"
	"github.com/lightyeario/kelp/support"
	"github.com/stellar/go/clients/horizon"
)

// SellConfig contains the configuration params for this SideStrategy
type SellConfig sideStrategy.SellSideConfig

// MakeSellStrategy is a factory method for SellStrategy
func MakeSellStrategy(
	txButler *kelp.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	config *SellConfig,
) Strategy {
	sellSideConfig := sideStrategy.SellSideConfig(*config)
	sellSideStrategy := sideStrategy.MakeSellSideStrategy(
		txButler,
		assetBase,
		assetQuote,
		&sellSideConfig,
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
