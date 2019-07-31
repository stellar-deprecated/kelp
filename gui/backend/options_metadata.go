package backend

import (
	"fmt"
	"net/http"
)

type metadata interface{}

type dropdown struct {
	Type    string                    `json:"type"`
	Options map[string]dropdownOption `json:"options"`
}

var _ metadata = dropdown{}

type text struct {
	Type         string      `json:"type"`
	DefaultValue string      `json:"defaultValue"`
	Subtype      interface{} `json:"subtype"`
}

var _ metadata = text{}

type dropdownOption struct {
	Value   string   `json:"value"`
	Text    string   `json:"text"`
	Subtype metadata `json:"subtype"`
}

const dropdownString = "dropdown"
const textString = "text"

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

func (s *APIServer) optionsMetadata(w http.ResponseWriter, r *http.Request) {
	mdata := dropdown{
		Type: dropdownString,
		Options: map[string]dropdownOption{
			"crypto": dropdownOption{
				Value: "crypto",
				Text:  "Crypto from CMC",
				Subtype: dropdown{
					Type: dropdownString,
					Options: map[string]dropdownOption{
						"https://api.coinmarketcap.com/v1/ticker/stellar/": dropdownOption{
							Value:   "https://api.coinmarketcap.com/v1/ticker/stellar/",
							Text:    "Stellar",
							Subtype: nil,
						},
						"https://api.coinmarketcap.com/v1/ticker/bitcoin/": dropdownOption{
							Value:   "https://api.coinmarketcap.com/v1/ticker/bitcoin/",
							Text:    "Bitcoin",
							Subtype: nil,
						},
					},
				},
			},
			"exchange": dropdownOption{
				Value: "exchange",
				Text:  "Centralized Exchange",
				Subtype: dropdown{
					Type: dropdownString,
					Options: map[string]dropdownOption{
						"kraken": dropdownOption{
							Value: "kraken",
							Text:  "Kraken",
							Subtype: dropdown{
								Type: dropdownString,
								Options: map[string]dropdownOption{
									"XXLM/ZUSD": dropdownOption{
										Value:   "XXLM/ZUSD",
										Text:    "XLM/USD",
										Subtype: nil,
									},
								},
							},
						},
					},
				},
			},
			"fixed": dropdownOption{
				Value: "fixed",
				Text:  "Fixed Value",
				Subtype: text{
					Type:         textString,
					DefaultValue: "1.0",
					Subtype:      nil,
				},
			},
		},
	}

	var e error
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("cannot get optionsMetadata: %s", e))
		return
	}

	s.writeJson(w, mdata)
}
