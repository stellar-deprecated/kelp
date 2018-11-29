package trader

import (
	"fmt"

	"github.com/interstellar/kelp/support/utils"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// XLM is a constant for XLM
const XLM = "XLM"

// BotConfig represents the configuration params for the bot
type BotConfig struct {
	SourceSecretSeed      string `valid:"-" toml:"SOURCE_SECRET_SEED"`
	TradingSecretSeed     string `valid:"-" toml:"TRADING_SECRET_SEED"`
	AssetCodeA            string `valid:"-" toml:"ASSET_CODE_A"`
	IssuerA               string `valid:"-" toml:"ISSUER_A"`
	AssetCodeB            string `valid:"-" toml:"ASSET_CODE_B"`
	IssuerB               string `valid:"-" toml:"ISSUER_B"`
	TickIntervalSeconds   int32  `valid:"-" toml:"TICK_INTERVAL_SECONDS"`
	DeleteCyclesThreshold int64  `valid:"-" toml:"DELETE_CYCLES_THRESHOLD"`
	HorizonURL            string `valid:"-" toml:"HORIZON_URL"`
	AlertType             string `valid:"-" toml:"ALERT_TYPE"`
	AlertAPIKey           string `valid:"-" toml:"ALERT_API_KEY"`
	MonitoringPort        uint16 `valid:"-" toml:"MONITORING_PORT"`
	MonitoringTLSCert     string `valid:"-" toml:"MONITORING_TLS_CERT"`
	MonitoringTLSKey      string `valid:"-" toml:"MONITORING_TLS_KEY"`
	GoogleClientID        string `valid:"-" toml:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret    string `valid:"-" toml:"GOOGLE_CLIENT_SECRET"`
	AcceptableEmails      string `valid:"-" toml:"ACCEPTABLE_GOOGLE_EMAILS"`

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
	if b.AssetCodeA == b.AssetCodeB && b.IssuerA == b.IssuerB {
		return fmt.Errorf("error: both assets cannot be the same '%s:%s'", b.AssetCodeA, b.IssuerA)
	}

	asset, e := parseAsset(b.AssetCodeA, b.IssuerA, "A")
	if e != nil {
		return e
	}
	b.assetBase = *asset

	asset, e = parseAsset(b.AssetCodeB, b.IssuerB, "B")
	if e != nil {
		return e
	}
	b.assetQuote = *asset

	b.tradingAccount, e = utils.ParseSecret(b.TradingSecretSeed)
	if e != nil {
		return e
	}
	if b.tradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	b.sourceAccount, e = utils.ParseSecret(b.SourceSecretSeed)
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
