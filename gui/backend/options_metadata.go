package backend

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

type metadata interface{}

type dropdownType struct {
	Type    string                    `json:"type"`
	Options map[string]dropdownOption `json:"options"`
}

var _ metadata = dropdownType{}

type textType struct {
	Type         string      `json:"type"`
	DefaultValue string      `json:"defaultValue"`
	Subtype      interface{} `json:"subtype"`
}

var _ metadata = textType{}

type dropdownOption struct {
	Value   string   `json:"value"`
	Text    string   `json:"text"`
	Subtype metadata `json:"subtype"`
}

const dropdownString = "dropdown"
const textString = "text"

var ccxtExchangeNames = map[string]string{
	"bitfinex2":     "Bitfinex [alternate]",
	"hitbtc2":       "Hitbtc [alternate]",
	"_1broker":      "1broker",
	"_1btcxe":       "1btcxe",
	"coinmarketcap": "CoinMarketCap",
}

var ccxtBlacklist = utils.StringSet([]string{
	"_1broker",
	"allcoin",
	"anybits",
	"bibox",
	"bitsane",
	"bitz",
	"btctradeim",
	"ccex",
	"coinegg",
	"coingi",
	"cointiger",
	"coolcoin",
	"cryptopia",
	"flowbtc",
	"gatecoin",
	"huobicny",
	"kucoin",
	"liqui",
	"rightbtc",
	"theocean",
	"tidex",
	"wex",
	"xbtce",
	"yunbi",
})

//   const optionsMetadata = {
// 	type: "dropdown",
// 	options: {
// 	  "crypto": {
// 		value: "crypto",
// 		text: "Crypto from CMC",
// 		subtype: {
// 		  type: "dropdown",
// 		  options: {
// 			"https://api.coinmarketcap.com/v1/ticker/stellar/": {
// 			  value: "https://api.coinmarketcap.com/v1/ticker/stellar/",
// 			  text: "Stellar",
// 			  subtype: null,
// 			},
// 			"https://api.coinmarketcap.com/v1/ticker/bitcoin/": {
// 			  value: "https://api.coinmarketcap.com/v1/ticker/bitcoin/",
// 			  text: "Bitcoin",
// 			  subtype: null,
// 			},
// 			"https://api.coinmarketcap.com/v1/ticker/ethereum/": {
// 			  value: "https://api.coinmarketcap.com/v1/ticker/ethereum/",
// 			  text: "Ethereum",
// 			  subtype: null,
// 			},
// 			"https://api.coinmarketcap.com/v1/ticker/litecoin/": {
// 			  value: "https://api.coinmarketcap.com/v1/ticker/litecoin/",
// 			  text: "Litecoin",
// 			  subtype: null,
// 			},
// 			"https://api.coinmarketcap.com/v1/ticker/tether/": {
// 			  value: "https://api.coinmarketcap.com/v1/ticker/tether/",
// 			  text: "Tether",
// 			  subtype: null,
// 			}
// 		  }
// 		}
// 	  },
// 	  "fiat": {
// 		value: "fiat",
// 		text: "Fiat from CurrencyLayer",
// 		subtype: {
// 		  type: "dropdown",
// 		  options: {
// 			"http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=USD": {
// 			  value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=USD",
// 			  text: "USD",
// 			  subtype: null,
// 			},
// 			"http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=EUR": {
// 			  value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=EUR",
// 			  text: "EUR",
// 			  subtype: null,
// 			},
// 			"http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=GBP": {
// 			  value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=GBP",
// 			  text: "GBP",
// 			  subtype: null,
// 			},
// 			"http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=INR": {
// 			  value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=INR",
// 			  text: "INR",
// 			  subtype: null,
// 			},
// 			"http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=PHP": {
// 			  value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=PHP",
// 			  text: "PHP",
// 			  subtype: null,
// 			},
// 			"http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=NGN": {
// 			  value: "http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=NGN",
// 			  text: "NGN",
// 			  subtype: null,
// 			}
// 		  }
// 		}
// 	  },
// 	  "fixed": {
// 		value: "fixed",
// 		text: "Fixed value",
// 		subtype: {
// 		  type: "text",
// 		  defaultValue: "1.0",
// 		  subtype: null,
// 		}
// 	  },
// 	  "exchange": {
// 		value: "exchange",
// 		text: "Centralized Exchange",
// 		subtype: {
// 		  type: "dropdown",
// 		  options: {
// 			"kraken": {
// 			  value: "kraken",
// 			  text: "Kraken",
// 			  subtype: {
// 				type: "dropdown",
// 				options: {
// 				  "XXLM/ZUSD": {
// 					value: "XXLM/ZUSD",
// 					text: "XLM/USD",
// 					subtype: null,
// 				  },
// 				  "XXLM/XXBT": {
// 					value: "XXLM/XXBT",
// 					text: "XLM/BTC",
// 					subtype: null,
// 				  },
// 				  "XXBT/ZUSD": {
// 					value: "XXBT/ZUSD",
// 					text: "BTC/USD",
// 					subtype: null,
// 				  },
// 				  "XETH/ZUSD": {
// 					value: "XETH/ZUSD",
// 					text: "ETH/USD",
// 					subtype: null,
// 				  },
// 				  "XETH/XXBT": {
// 					value: "XETH/XXBT",
// 					text: "ETH/BTC",
// 					subtype: null,
// 				  }
// 				}
// 			  }
// 			},
// 			"ccxt-binance": {
// 			  value: "ccxt-binance",
// 			  text: "Binance (via CCXT)",
// 			  subtype: {
// 				type: "dropdown",
// 				options: {
// 				  "BTC/USDT": {
// 					value: "BTC/USDT",
// 					text: "BTC/USDT",
// 					subtype: null,
// 				  },
// 				  "ETH/USDT": {
// 					value: "ETH/USDT",
// 					text: "ETH/USDT",
// 					subtype: null,
// 				  },
// 				  "BNB/USDT": {
// 					value: "BNB/USDT",
// 					text: "BNB/USDT",
// 					subtype: null,
// 				  },
// 				  "BNB/BTC": {
// 					value: "BNB/USDT",
// 					text: "BNB/USDT",
// 					subtype: null,
// 				  },
// 				  "XLM/USDT": {
// 					value: "XLM/USDT",
// 					text: "XLM/USDT",
// 					subtype: null,
// 				  },
// 				}
// 			  }
// 			}
// 		  }
// 		}
// 	  }
// 	}
//   };

func dropdown(options *dropdownOptionsBuilder) dropdownType {
	return dropdownType{
		Type:    dropdownString,
		Options: options._build(),
	}
}

func text(defaultValue string) textType {
	return textType{
		Type:         textString,
		DefaultValue: defaultValue,
		Subtype:      nil,
	}
}

type dropdownOptionsBuilder struct {
	m map[string]dropdownOption
}

func optionsBuilder() *dropdownOptionsBuilder {
	return &dropdownOptionsBuilder{
		m: map[string]dropdownOption{},
	}
}

func (dob *dropdownOptionsBuilder) ccxtMarket(marketCode string) *dropdownOptionsBuilder {
	return dob._leaf(marketCode, marketCode)
}

func (dob *dropdownOptionsBuilder) krakenMarket(marketCode string, marketName string) *dropdownOptionsBuilder {
	return dob._leaf(marketCode, marketName)
}

func (dob *dropdownOptionsBuilder) coinmarketcap(tickerCode string, currencyName string) *dropdownOptionsBuilder {
	return dob._leaf(fmt.Sprintf("https://api.coinmarketcap.com/v1/ticker/%s/", tickerCode), currencyName)
}

func (dob *dropdownOptionsBuilder) currencylayer(tickerCode string) *dropdownOptionsBuilder {
	return dob._leaf(fmt.Sprintf("http://apilayer.net/api/live?access_key=8db4ba3aa504c601dd513777193f4f2b&currencies=%s", tickerCode), tickerCode)
}

func (dob *dropdownOptionsBuilder) _leaf(value string, text string) *dropdownOptionsBuilder {
	return dob.option(value, text, nil)
}

func (dob *dropdownOptionsBuilder) option(value string, text string, subtype metadata) *dropdownOptionsBuilder {
	dob.m[value] = dropdownOption{
		Value:   value,
		Text:    text,
		Subtype: subtype,
	}
	return dob
}

func (dob *dropdownOptionsBuilder) includeOptions(other *dropdownOptionsBuilder) *dropdownOptionsBuilder {
	for key, value := range other._build() {
		dob.option(key, value.Text, value.Subtype)
	}
	return dob
}

func (dob *dropdownOptionsBuilder) _build() map[string]dropdownOption {
	return dob.m
}

func loadOptionsMetadata() (metadata, error) {
	ccxtOptions := optionsBuilder()
	for _, ccxtExchangeName := range sdk.GetExchangeList() {
		if _, ok := ccxtBlacklist[ccxtExchangeName]; ok {
			continue
		}

		displayName := strings.Title(ccxtExchangeName)
		if name, ok := ccxtExchangeNames[ccxtExchangeName]; ok {
			displayName = name
		}
		displayName = displayName + " (via CCXT)"

		c, e := sdk.MakeInitializedCcxtExchange(ccxtExchangeName, api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
		if e != nil {
			// don't block if we are unable to load an exchange
			log.Printf("unable to make ccxt exchange '%s' when trying to load options metadata, continuing: %s\n", ccxtExchangeName, e)
			continue
		}

		marketsBuilder := optionsBuilder()
		for tradingPair := range c.GetMarkets() {
			marketsBuilder.ccxtMarket(tradingPair)
		}
		ccxtOptions.option("ccxt-"+ccxtExchangeName, displayName, dropdown(marketsBuilder))
	}

	builder := optionsBuilder().
		option("crypto", "Crypto (CMC)", dropdown(optionsBuilder().
			coinmarketcap("stellar", "Stellar").
			coinmarketcap("bitcoin", "Bitcoin").
			coinmarketcap("ethereum", "Ethereum").
			coinmarketcap("litecoin", "Litecoin"))).
		option("exchange", "Centralized Exchange", dropdown(optionsBuilder().
			option("kraken", "Kraken", dropdown(optionsBuilder().
				krakenMarket("XXLM/ZUSD", "XLM/USD"))).
			includeOptions(ccxtOptions))).
		option("fiat", "Fiat (CurrencyLayer)", dropdown(optionsBuilder().
			currencylayer("USD").
			currencylayer("EUR").
			currencylayer("GBP").
			currencylayer("INR"))).
		option("fixed", "Fixed Value", text("1.0"))
	mdata := dropdown(builder)
	return mdata, nil
}

func (s *APIServer) optionsMetadata(w http.ResponseWriter, r *http.Request) {
	s.writeJsonWithLog(w, s.cachedOptionsMetadata, false)
}
