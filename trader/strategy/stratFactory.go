package strategy

import (
	"os"

	kelp "github.com/lightyeario/kelp/support"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

// StratFactory makes a strategy
func StratFactory(
	txButler *kelp.TxButler,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	stratType string,
	stratConfigPath string,
) Strategy {
	switch stratType {
	case "simple":
		var cfg SimpleConfig
		err := config.Read(stratConfigPath, &cfg)
		kelp.CheckConfigError(cfg, err)
		return MakeSimpleStrategy(txButler, assetBase, assetQuote, &cfg)
	case "mirror":
		var cfg MirrorConfig
		err := config.Read(stratConfigPath, &cfg)
		kelp.CheckConfigError(cfg, err)
		return MakeMirrorStrategy(txButler, assetBase, assetQuote, &cfg)
	case "sell":
		var cfg SellConfig
		err := config.Read(stratConfigPath, &cfg)
		kelp.CheckConfigError(cfg, err)
		return MakeSellStrategy(txButler, assetBase, assetQuote, &cfg)
	case "autonomous":
		var cfg AutonomousConfig
		err := config.Read(stratConfigPath, &cfg)
		kelp.CheckConfigError(cfg, err)
		return MakeAutonomousStrategy(txButler, assetBase, assetQuote, &cfg)
	}

	log.Errorf("invalid strategy type: %s", stratType)
	os.Exit(1)
	return nil
}
