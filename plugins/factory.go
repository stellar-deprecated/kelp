package plugins

import (
	"fmt"
	"log"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/config"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

// strategyFactoryData is a data container that has all the information needed to make a strategy
type strategyFactoryData struct {
	sdex            *SDEX
	ieif            *IEIF
	tradingPair     *model.TradingPair
	assetBase       *horizon.Asset
	assetQuote      *horizon.Asset
	stratConfigPath string
	simMode         bool
}

// StrategyContainer contains the strategy factory method along with some metadata
type StrategyContainer struct {
	SortOrder   uint8
	Description string
	NeedsConfig bool
	Complexity  string
	makeFn      func(strategyFactoryData strategyFactoryData) (api.Strategy, error)
}

// strategies is a map of all the strategies available
var strategies = map[string]StrategyContainer{
	"buysell": {
		SortOrder:   1,
		Description: "Creates buy and sell offers based on a reference price with a pre-specified liquidity depth",
		NeedsConfig: true,
		Complexity:  "Beginner",
		makeFn: func(strategyFactoryData strategyFactoryData) (api.Strategy, error) {
			var cfg buySellConfig
			err := config.Read(strategyFactoryData.stratConfigPath, &cfg)
			utils.CheckConfigError(cfg, err, strategyFactoryData.stratConfigPath)
			utils.LogConfig(cfg)
			s, e := makeBuySellStrategy(strategyFactoryData.sdex, strategyFactoryData.tradingPair, strategyFactoryData.ieif, strategyFactoryData.assetBase, strategyFactoryData.assetQuote, &cfg)
			if e != nil {
				return nil, fmt.Errorf("makeFn failed: %s", e)
			}
			return s, nil
		},
	},
	"mirror": {
		SortOrder:   4,
		Description: "Mirrors an orderbook from another exchange by placing the same orders on Stellar",
		NeedsConfig: true,
		Complexity:  "Advanced",
		makeFn: func(strategyFactoryData strategyFactoryData) (api.Strategy, error) {
			var cfg mirrorConfig
			err := config.Read(strategyFactoryData.stratConfigPath, &cfg)
			utils.CheckConfigError(cfg, err, strategyFactoryData.stratConfigPath)
			utils.LogConfig(cfg)
			s, e := makeMirrorStrategy(strategyFactoryData.sdex, strategyFactoryData.ieif, strategyFactoryData.tradingPair, strategyFactoryData.assetBase, strategyFactoryData.assetQuote, &cfg, strategyFactoryData.simMode)
			if e != nil {
				return nil, fmt.Errorf("makeFn failed: %s", e)
			}
			return s, nil
		},
	},
	"sell": {
		SortOrder:   0,
		Description: "Creates sell offers based on a reference price with a pre-specified liquidity depth",
		NeedsConfig: true,
		Complexity:  "Beginner",
		makeFn: func(strategyFactoryData strategyFactoryData) (api.Strategy, error) {
			var cfg sellConfig
			err := config.Read(strategyFactoryData.stratConfigPath, &cfg)
			utils.CheckConfigError(cfg, err, strategyFactoryData.stratConfigPath)
			utils.LogConfig(cfg)
			s, e := makeSellStrategy(strategyFactoryData.sdex, strategyFactoryData.tradingPair, strategyFactoryData.ieif, strategyFactoryData.assetBase, strategyFactoryData.assetQuote, &cfg)
			if e != nil {
				return nil, fmt.Errorf("makeFn failed: %s", e)
			}
			return s, nil
		},
	},
	"balanced": {
		SortOrder:   3,
		Description: "Dynamically prices two tokens based on their relative demand",
		NeedsConfig: true,
		Complexity:  "Intermediate",
		makeFn: func(strategyFactoryData strategyFactoryData) (api.Strategy, error) {
			var cfg balancedConfig
			err := config.Read(strategyFactoryData.stratConfigPath, &cfg)
			utils.CheckConfigError(cfg, err, strategyFactoryData.stratConfigPath)
			utils.LogConfig(cfg)
			return makeBalancedStrategy(strategyFactoryData.sdex, strategyFactoryData.tradingPair, strategyFactoryData.ieif, strategyFactoryData.assetBase, strategyFactoryData.assetQuote, &cfg), nil
		},
	},
	"delete": {
		SortOrder:   2,
		Description: "Deletes all orders for the configured orderbook",
		NeedsConfig: false,
		Complexity:  "Beginner",
		makeFn: func(strategyFactoryData strategyFactoryData) (api.Strategy, error) {
			return makeDeleteStrategy(strategyFactoryData.sdex, strategyFactoryData.assetBase, strategyFactoryData.assetQuote), nil
		},
	},
}

// MakeStrategy makes a strategy
func MakeStrategy(
	sdex *SDEX,
	ieif *IEIF,
	tradingPair *model.TradingPair,
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	strategy string,
	stratConfigPath string,
	simMode bool,
) (api.Strategy, error) {
	log.Printf("Making strategy: %s\n", strategy)
	if s, ok := strategies[strategy]; ok {
		if s.NeedsConfig && stratConfigPath == "" {
			return nil, fmt.Errorf("the '%s' strategy needs a config file", strategy)
		}

		s, e := s.makeFn(strategyFactoryData{
			sdex:            sdex,
			ieif:            ieif,
			tradingPair:     tradingPair,
			assetBase:       assetBase,
			assetQuote:      assetQuote,
			stratConfigPath: stratConfigPath,
			simMode:         simMode,
		})
		if e != nil {
			return nil, fmt.Errorf("cannot make '%s' strategy: %s", strategy, e)
		}
		return s, nil
	}

	return nil, fmt.Errorf("invalid strategy type: %s", strategy)
}

// Strategies returns the list of strategies along with metadata
func Strategies() map[string]StrategyContainer {
	return strategies
}

// exchangeFactoryData is a data container that has all the information needed to make an exchange
type exchangeFactoryData struct {
	simMode bool
	apiKeys []api.ExchangeAPIKey
}

// ExchangeContainer contains the exchange factory method along with some metadata
type ExchangeContainer struct {
	SortOrder    uint16
	Description  string
	TradeEnabled bool
	Tested       bool
	makeFn       func(exchangeFactoryData exchangeFactoryData) (api.Exchange, error)
}

// exchanges is a map of all the exchange integrations available
var exchanges *map[string]ExchangeContainer

// getExchanges returns a map of all the exchange integrations available
func getExchanges() map[string]ExchangeContainer {
	if exchanges == nil {
		loadExchanges()
	}
	return *exchanges
}

func loadExchanges() {
	// marked as tested if key exists in this map (regardless of bool value)
	testedCcxtExchanges := map[string]bool{
		"binance": true,
	}

	exchanges = &map[string]ExchangeContainer{
		"kraken": {
			SortOrder:    0,
			Description:  "Kraken is a popular centralized cryptocurrency exchange",
			TradeEnabled: true,
			Tested:       true,
			makeFn: func(exchangeFactoryData exchangeFactoryData) (api.Exchange, error) {
				return makeKrakenExchange(exchangeFactoryData.apiKeys, exchangeFactoryData.simMode)
			},
		},
	}

	// add all CCXT exchanges (tested exchanges first)
	sortOrderIndex := len(*exchanges)
	for _, t := range []bool{true, false} {
		for _, exchangeName := range sdk.GetExchangeList() {
			key := fmt.Sprintf("ccxt-%s", exchangeName)
			_, tested := testedCcxtExchanges[exchangeName]
			if tested != t {
				continue
			}
			boundExchangeName := exchangeName

			(*exchanges)[key] = ExchangeContainer{
				SortOrder:    uint16(sortOrderIndex),
				Description:  exchangeName + " is automatically added via ccxt-rest",
				TradeEnabled: true,
				Tested:       tested,
				makeFn: func(exchangeFactoryData exchangeFactoryData) (api.Exchange, error) {
					return makeCcxtExchange(
						boundExchangeName,
						nil,
						exchangeFactoryData.apiKeys,
						exchangeFactoryData.simMode,
					)
				},
			}
			sortOrderIndex++
		}
	}
}

// MakeExchange is a factory method to make an exchange based on a given type
func MakeExchange(exchangeType string, simMode bool) (api.Exchange, error) {
	if exchange, ok := getExchanges()[exchangeType]; ok {
		exchangeAPIKey := api.ExchangeAPIKey{Key: "", Secret: ""}
		x, e := exchange.makeFn(exchangeFactoryData{
			simMode: simMode,
			apiKeys: []api.ExchangeAPIKey{exchangeAPIKey},
		})
		if e != nil {
			return nil, fmt.Errorf("error when making the '%s' exchange: %s", exchangeType, e)
		}
		return x, nil
	}

	return nil, fmt.Errorf("invalid exchange type: %s", exchangeType)
}

// MakeTradingExchange is a factory method to make an exchange based on a given type
func MakeTradingExchange(exchangeType string, apiKeys []api.ExchangeAPIKey, simMode bool) (api.Exchange, error) {
	if exchange, ok := getExchanges()[exchangeType]; ok {
		if !exchange.TradeEnabled {
			return nil, fmt.Errorf("trading is not enabled on this exchange: %s", exchangeType)
		}

		if len(apiKeys) == 0 {
			return nil, fmt.Errorf("cannot make trading exchange, apiKeys mising")
		}

		x, e := exchange.makeFn(exchangeFactoryData{
			simMode: simMode,
			apiKeys: apiKeys,
		})
		if e != nil {
			return nil, fmt.Errorf("error when making the '%s' exchange: %s", exchangeType, e)
		}
		return x, nil
	}

	return nil, fmt.Errorf("invalid exchange type: %s", exchangeType)
}

// Exchanges returns the list of exchanges
func Exchanges() map[string]ExchangeContainer {
	return getExchanges()
}
