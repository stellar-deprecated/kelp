package main

import (
	"fmt"

	"github.com/lightyeario/kelp/support"
)

// Config represents the configuration params for the bot
type Config struct {
	SOURCE_SECRET_SEED     string `valid:"-"`
	TRADING_SECRET_SEED    string `valid:"-"`
	ALLOW_INACTIVE_MINUTES int32  `valid:"-"` // bots that are inactive for more than this time will have its offers deleted
	TICK_INTERVAL_SECONDS  int32  `valid:"-"`
	HORIZON_URL            string `valid:"-"`

	tradingAccount *string
	sourceAccount  *string // can be nil
}

// Init initializes this config
func (c *Config) Init() error {
	var e error
	c.tradingAccount, e = kelp.ParseSecret(c.TRADING_SECRET_SEED)
	if e != nil {
		return e
	}
	// trading account should never be nil
	if c.tradingAccount == nil {
		return fmt.Errorf("no trading account specified")
	}

	c.sourceAccount, e = kelp.ParseSecret(c.SOURCE_SECRET_SEED)
	if e != nil {
		return e
	}

	return nil
}
