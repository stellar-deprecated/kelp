package strategy

import (
	"fmt"

	kelp "github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
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
	assetBase      horizon.Asset
	assetQuote     horizon.Asset
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

// AssetBase returns the config's assetBase
func (b *BotConfig) AssetBase() horizon.Asset {
	return b.assetBase
}

// AssetQuote returns the config's assetQuote
func (b *BotConfig) AssetQuote() horizon.Asset {
	return b.assetQuote
}

// Init initializes this config
func (b *BotConfig) Init() error {
	if b.ASSET_CODE_A == "XLM" {
		b.assetBase = kelp.Asset2Asset2(build.NativeAsset())
	} else {
		b.assetBase = kelp.Asset2Asset2(build.CreditAsset(b.ASSET_CODE_A, b.ISSUER_A))
	}

	if b.ASSET_CODE_B == "XLM" {
		b.assetQuote = kelp.Asset2Asset2(build.NativeAsset())
	} else {
		b.assetQuote = kelp.Asset2Asset2(build.CreditAsset(b.ASSET_CODE_B, b.ISSUER_B))
	}

	var e error
	b.tradingAccount, e = kelp.ParseSecret(b.TRADING_SECRET_SEED)
	if e != nil {
		return e
	}
	// trading account should never be nil
	if b.tradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	b.sourceAccount, e = kelp.ParseSecret(b.SOURCE_SECRET_SEED)
	return e
}
