package terminator

import (
	"fmt"

	"github.com/stellar/kelp/support/utils"
)

// Config represents the configuration params for the bot
type Config struct {
	SourceSecretSeed     string `valid:"-" toml:"SOURCE_SECRET_SEED"`
	TradingSecretSeed    string `valid:"-" toml:"TRADING_SECRET_SEED"`
	AllowInactiveMinutes int32  `valid:"-" toml:"ALLOW_INACTIVE_MINUTES"` // bots that are inactive for more than this time will have its offers deleted
	TickIntervalSeconds  int32  `valid:"-" toml:"TICK_INTERVAL_SECONDS"`
	HorizonURL           string `valid:"-" toml:"HORIZON_URL"`

	TradingAccount *string
	SourceAccount  *string // can be nil
}

// String impl.
func (c Config) String() string {
	return utils.StructString(c, 0, map[string]func(interface{}) interface{}{
		"SOURCE_SECRET_SEED":  utils.SecretKey2PublicKey,
		"TRADING_SECRET_SEED": utils.SecretKey2PublicKey,
	})
}

// Init initializes this config
func (c *Config) Init() error {
	var e error
	c.TradingAccount, e = utils.ParseSecret(c.TradingSecretSeed)
	if e != nil {
		return e
	}
	// trading account should never be nil
	if c.TradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	c.SourceAccount, e = utils.ParseSecret(c.SourceSecretSeed)
	return e
}
