package plugins

import (
	"os"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

// MakeExchange is a factory method to make an exchange based on a given type
func MakeExchange(exchangeType string) api.Exchange {
	switch exchangeType {
	case "kraken":
		return makeKrakenExchange()
	}
	return nil
}

// MakeStrategy makes a strategy
func MakeStrategy(
	sdex *SDEX,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	stratType string,
	stratConfigPath string,
) api.Strategy {
	switch stratType {
	case "buysell":
		var cfg buySellConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return makeBuySellStrategy(sdex, assetBase, assetQuote, &cfg)
	case "mirror":
		var cfg mirrorConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return makeMirrorStrategy(sdex, assetBase, assetQuote, &cfg)
	case "sell":
		var cfg sellConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return makeSellStrategy(sdex, assetBase, assetQuote, &cfg)
	case "balanced":
		var cfg balancedConfig
		err := config.Read(stratConfigPath, &cfg)
		utils.CheckConfigError(cfg, err)
		return makeBalancedStrategy(sdex, assetBase, assetQuote, &cfg)
	case "delete":
		return makeDeleteStrategy(sdex, assetBase, assetQuote)
	}

	log.Errorf("invalid strategy type: %s", stratType)
	os.Exit(1)
	return nil
}
