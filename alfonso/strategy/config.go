package strategy

import (
	"fmt"
	"os"

	kelp "github.com/lightyeario/kelp/support"
	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

// BotConfig represents the configuration params for the bot
type BotConfig struct {
	SOURCE_SECRET_SEED    string `valid:"-"`
	TRADING_SECRET_SEED   string `valid:"-"`
	ASSET_CODE_A          string `valid:"-"`
	ISSUER_A              string `valid:"-"`
	ASSET_CODE_B          string `valid:"-"`
	ISSUER_B              string `valid:"-"`
	TICK_INTERVAL_SECONDS int32  `valid:"-"`
	HORIZON_URL           string `valid:"-"`

	tradingAccount *string
	sourceAccount  *string // can be nil
	assetA         horizon.Asset
	assetB         horizon.Asset
}

// TradingAccount returns the config's trading account
func (b *BotConfig) TradingAccount() string {
	return *b.tradingAccount
}

// SourceAccount returns the config's source account
func (b *BotConfig) SourceAccount() string {
	if b.sourceAccount == nil {
		return ""
	}
	return *b.sourceAccount
}

// AssetA returns the config's assetA
func (b *BotConfig) AssetA() horizon.Asset {
	return b.assetA
}

// AssetB returns the config's assetB
func (b *BotConfig) AssetB() horizon.Asset {
	return b.assetB
}

// Init initializes this config
func (b *BotConfig) Init() error {
	if b.ASSET_CODE_A == "XLM" {
		b.assetA = kelp.Asset2Asset2(build.NativeAsset())
	} else {
		b.assetA = kelp.Asset2Asset2(build.CreditAsset(b.ASSET_CODE_A, b.ISSUER_A))
	}

	if b.ASSET_CODE_B == "XLM" {
		b.assetB = kelp.Asset2Asset2(build.NativeAsset())
	} else {
		b.assetB = kelp.Asset2Asset2(build.CreditAsset(b.ASSET_CODE_B, b.ISSUER_B))
	}

	var e error
	b.tradingAccount, e = parseSecret(b.TRADING_SECRET_SEED)
	if e != nil {
		return e
	}
	// trading account should never be nil
	if b.tradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	b.sourceAccount, e = parseSecret(b.SOURCE_SECRET_SEED)
	if e != nil {
		return e
	}

	return nil
}

// CheckConfigError checks configs for errors
func CheckConfigError(cfg interface{}, e error) {
	fmt.Printf("Result: %+v\n", cfg)

	if e != nil {
		switch cause := errors.Cause(e).(type) {
		case *config.InvalidConfigError:
			log.Error("config file: ", cause)
		default:
			log.Error(e)
		}
		os.Exit(1)
	}
}

func parseSecret(secret string) (*string, error) {
	if secret == "" {
		return nil, nil
	}

	sourceKP, err := keypair.Parse(secret)
	if err != nil {
		return nil, err
	}

	address := sourceKP.Address()
	return &address, nil
}
