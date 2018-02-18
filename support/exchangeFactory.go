package kelp

import (
	"github.com/lightyeario/kelp/support/exchange"
	"github.com/lightyeario/kelp/support/kraken"
)

// ExchangeFactory is a factory method to make an exchange based on a given type
func ExchangeFactory(exchangeType string) exchange.Exchange {
	switch exchangeType {
	case "kraken":
		return kraken.MakeKrakenExchange()
	}
	return nil
}
