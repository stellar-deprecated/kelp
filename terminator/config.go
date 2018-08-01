package terminator

import (
	"fmt"

	"github.com/lightyeario/kelp/support/utils"
)

// Config represents the configuration params for the bot
type Config struct {
	SOURCE_SECRET_SEED     string `valid:"-"`
	TRADING_SECRET_SEED    string `valid:"-"`
	ALLOW_INACTIVE_MINUTES int32  `valid:"-"` // bots that are inactive for more than this time will have its offers deleted
	TICK_INTERVAL_SECONDS  int32  `valid:"-"`
	HORIZON_URL            string `valid:"-"`

	TradingAccount *string
	SourceAccount  *string // can be nil
}

// String impl.
func (c Config) String() string {
	return utils.StructString(c, map[string]func(interface{}) interface{}{
		"SOURCE_SECRET_SEED":  utils.SecretKey2PublicKey,
		"TRADING_SECRET_SEED": utils.SecretKey2PublicKey,
	})
}

// Init initializes this config
func (c *Config) Init() error {
	var e error
	c.TradingAccount, e = utils.ParseSecret(c.TRADING_SECRET_SEED)
	if e != nil {
		return e
	}
	// trading account should never be nil
	if c.TradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	c.SourceAccount, e = utils.ParseSecret(c.SOURCE_SECRET_SEED)
	return e
}
