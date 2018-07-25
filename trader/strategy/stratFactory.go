package strategy

import (
	"os"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/plugins"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

// StratFactory makes a strategy
func StratFactory(
	sdex *plugins.SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	stratType string,
	stratConfigPath string,
) api.Strategy {
	switch stratType {
	case "buysell":
		var cfg BuySellConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return MakeBuySellStrategy(sdex, assetBase, assetQuote, &cfg)
	case "mirror":
		var cfg MirrorConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return MakeMirrorStrategy(sdex, assetBase, assetQuote, &cfg)
	case "sell":
		var cfg SellConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return MakeSellStrategy(sdex, assetBase, assetQuote, &cfg)
	case "balanced":
		var cfg BalancedConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return MakeBalancedStrategy(sdex, assetBase, assetQuote, &cfg)
	case "delete":
		return MakeDeleteStrategy(sdex, assetBase, assetQuote)
	}

	log.Errorf("invalid strategy type: %s", stratType)
	os.Exit(1)
	return nil
}
