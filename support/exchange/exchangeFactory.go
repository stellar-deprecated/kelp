package exchange

import (
	"github.com/lightyeario/kelp/support/exchange/api"
	"github.com/lightyeario/kelp/support/exchange/kraken"
)

// ExchangeFactory is a factory method to make an exchange based on a given type
func ExchangeFactory(exchangeType string) api.Exchange {
	switch exchangeType {
	case "kraken":
		return kraken.MakeKrakenExchange()
	}
	return nil
}
