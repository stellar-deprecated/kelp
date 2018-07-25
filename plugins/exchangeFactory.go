package plugins

import (
	"github.com/lightyeario/kelp/api"
)

// MakeExchange is a factory method to make an exchange based on a given type
func MakeExchange(exchangeType string) api.Exchange {
	switch exchangeType {
	case "kraken":
		return MakeKrakenExchange()
	}
	return nil
}
