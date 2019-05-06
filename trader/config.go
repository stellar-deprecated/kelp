package trader

import (
	"fmt"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/support/utils"
)

// XLM is a constant for XLM
const XLM = "XLM"

// FeeConfig represents input data for how to deal with network fees
type FeeConfig struct {
	CapacityTrigger float64 `valid:"-" toml:"CAPACITY_TRIGGER"`   // trigger when "ledger_capacity_usage" in /fee_stats is >= this value
	Percentile      uint8   `valid:"-" toml:"PERCENTILE"`         // percentile computation to use from /fee_stats (10, 20, ..., 90, 95, 99)
	MaxOpFeeStroops uint64  `valid:"-" toml:"MAX_OP_FEE_STROOPS"` // max fee in stroops per operation to use
}

// BotConfig represents the configuration params for the bot
type BotConfig struct {
	SourceSecretSeed                   string     `valid:"-" toml:"SOURCE_SECRET_SEED"`
	TradingSecretSeed                  string     `valid:"-" toml:"TRADING_SECRET_SEED"`
	AssetCodeA                         string     `valid:"-" toml:"ASSET_CODE_A"`
	IssuerA                            string     `valid:"-" toml:"ISSUER_A"`
	AssetCodeB                         string     `valid:"-" toml:"ASSET_CODE_B"`
	IssuerB                            string     `valid:"-" toml:"ISSUER_B"`
	TickIntervalSeconds                int32      `valid:"-" toml:"TICK_INTERVAL_SECONDS"`
	MaxTickDelayMillis                 int64      `valid:"-" toml:"MAX_TICK_DELAY_MILLIS"`
	DeleteCyclesThreshold              int64      `valid:"-" toml:"DELETE_CYCLES_THRESHOLD"`
	SubmitMode                         string     `valid:"-" toml:"SUBMIT_MODE"`
	FillTrackerSleepMillis             uint32     `valid:"-" toml:"FILL_TRACKER_SLEEP_MILLIS"`
	FillTrackerDeleteCyclesThreshold   int64      `valid:"-" toml:"FILL_TRACKER_DELETE_CYCLES_THRESHOLD"`
	HorizonURL                         string     `valid:"-" toml:"HORIZON_URL"`
	CcxtRestURL                        *string    `valid:"-" toml:"CCXT_REST_URL"`
	Fee                                *FeeConfig `valid:"-" toml:"FEE"`
	CentralizedPricePrecisionOverride  *int8      `valid:"-" toml:"CENTRALIZED_PRICE_PRECISION_OVERRIDE"`
	CentralizedVolumePrecisionOverride *int8      `valid:"-" toml:"CENTRALIZED_VOLUME_PRECISION_OVERRIDE"`
	// Deprecated: use CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE instead
	MinCentralizedBaseVolumeDeprecated *float64 `valid:"-" toml:"MIN_CENTRALIZED_BASE_VOLUME" deprecated:"true"`
	CentralizedMinBaseVolumeOverride   *float64 `valid:"-" toml:"CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE"`
	CentralizedMinQuoteVolumeOverride  *float64 `valid:"-" toml:"CENTRALIZED_MIN_QUOTE_VOLUME_OVERRIDE"`
	AlertType                          string   `valid:"-" toml:"ALERT_TYPE"`
	AlertAPIKey                        string   `valid:"-" toml:"ALERT_API_KEY"`
	MonitoringPort                     uint16   `valid:"-" toml:"MONITORING_PORT"`
	MonitoringTLSCert                  string   `valid:"-" toml:"MONITORING_TLS_CERT"`
	MonitoringTLSKey                   string   `valid:"-" toml:"MONITORING_TLS_KEY"`
	GoogleClientID                     string   `valid:"-" toml:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret                 string   `valid:"-" toml:"GOOGLE_CLIENT_SECRET"`
	AcceptableEmails                   string   `valid:"-" toml:"ACCEPTABLE_GOOGLE_EMAILS"`
	TradingExchange                    string   `valid:"-" toml:"TRADING_EXCHANGE"`
	ExchangeAPIKeys                    []struct {
		Key    string `valid:"-" toml:"KEY"`
		Secret string `valid:"-" toml:"SECRET"`
	} `valid:"-" toml:"EXCHANGE_API_KEYS"`
	ExchangeParams []struct {
		Param string `valid:"-" toml:"PARAM"`
		Value string `valid:"-" toml:"VALUE"`
	} `valid:"-" toml:"EXCHANGE_PARAMS"`
	ExchangeHeaders []struct {
		Header string `valid:"-" toml:"HEADER"`
		Value  string `valid:"-" toml:"VALUE"`
	} `valid:"-" toml:"EXCHANGE_HEADERS"`

	// initialized later
	tradingAccount *string
	sourceAccount  *string // can be nil
	assetBase      horizon.Asset
	assetQuote     horizon.Asset
	isTradingSdex  bool
}

// String impl.
func (b BotConfig) String() string {
	return utils.StructString(b, map[string]func(interface{}) interface{}{
		"EXCHANGE_API_KEYS":                     utils.Hide,
		"EXCHANGE_PARAMS":                       utils.Hide,
		"EXCHANGE_HEADERS":                      utils.Hide,
		"SOURCE_SECRET_SEED":                    utils.SecretKey2PublicKey,
		"TRADING_SECRET_SEED":                   utils.SecretKey2PublicKey,
		"ALERT_API_KEY":                         utils.Hide,
		"GOOGLE_CLIENT_ID":                      utils.Hide,
		"GOOGLE_CLIENT_SECRET":                  utils.Hide,
		"ACCEPTABLE_GOOGLE_EMAILS":              utils.Hide,
		"CENTRALIZED_PRICE_PRECISION_OVERRIDE":  utils.UnwrapInt8Pointer,
		"CENTRALIZED_VOLUME_PRECISION_OVERRIDE": utils.UnwrapInt8Pointer,
		"MIN_CENTRALIZED_BASE_VOLUME":           utils.UnwrapFloat64Pointer,
		"CENTRALIZED_MIN_BASE_VOLUME_OVERRIDE":  utils.UnwrapFloat64Pointer,
		"CENTRALIZED_MIN_QUOTE_VOLUME_OVERRIDE": utils.UnwrapFloat64Pointer,
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

// IsTradingSdex returns whether the config is set to trade on SDEX
func (b *BotConfig) IsTradingSdex() bool {
	return b.isTradingSdex
}

// Init initializes this config
func (b *BotConfig) Init() error {
	b.isTradingSdex = b.TradingExchange == "" || b.TradingExchange == "sdex"

	if b.AssetCodeA == b.AssetCodeB && b.IssuerA == b.IssuerB {
		return fmt.Errorf("error: both assets cannot be the same '%s:%s'", b.AssetCodeA, b.IssuerA)
	}

	asset, e := utils.ParseAsset(b.AssetCodeA, b.IssuerA)
	if e != nil {
		return fmt.Errorf("Error while parsing Asset A: %s", e)
	}
	b.assetBase = *asset

	asset, e = utils.ParseAsset(b.AssetCodeB, b.IssuerB)
	if e != nil {
		return fmt.Errorf("Error while parsing Asset B: %s", e)
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
