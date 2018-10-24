package trader

import (
	"fmt"

	"github.com/lightyeario/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

const XLM = "XLM"

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
	ALERT_TYPE            string `valid:"-"`
	ALERT_API_KEY         string `valid:"-"`

	tradingAccount *string
	sourceAccount  *string // can be nil
	assetBase      horizon.Asset
	assetQuote     horizon.Asset
}

// String impl.
func (b BotConfig) String() string {
	return utils.StructString(b, map[string]func(interface{}) interface{}{
		"SOURCE_SECRET_SEED":  utils.SecretKey2PublicKey,
		"TRADING_SECRET_SEED": utils.SecretKey2PublicKey,
	})
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
	if b.ASSET_CODE_A == b.ASSET_CODE_B && b.ISSUER_A == b.ISSUER_B {
		return fmt.Errorf("error: both assets cannot be the same '%s:%s'", b.ASSET_CODE_A, b.ISSUER_A)
	}

	asset, e := parseAsset(b.ASSET_CODE_A, b.ISSUER_A, "A")
	if e != nil {
		return e
	}
	b.assetBase = *asset

	asset, e = parseAsset(b.ASSET_CODE_B, b.ISSUER_B, "B")
	if e != nil {
		return e
	}
	b.assetQuote = *asset

	b.tradingAccount, e = utils.ParseSecret(b.TRADING_SECRET_SEED)
	if e != nil {
		return e
	}
	if b.tradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	b.sourceAccount, e = utils.ParseSecret(b.SOURCE_SECRET_SEED)
	return e
}

func parseAsset(code string, issuer string, letter string) (*horizon.Asset, error) {
	if code != XLM && issuer == "" {
		return nil, fmt.Errorf("error: ISSUER_%s can only be empty if ASSET_CODE_%s is '%s'", letter, letter, XLM)
	}

	if code == XLM && issuer != "" {
		return nil, fmt.Errorf("error: ISSUER_%s needs to be empty if ASSET_CODE_%s is '%s'", letter, letter, XLM)
	}

	if code == XLM {
		asset := utils.Asset2Asset2(build.NativeAsset())
		return &asset, nil
	}

	asset := utils.Asset2Asset2(build.CreditAsset(code, issuer))
	return &asset, nil
}
