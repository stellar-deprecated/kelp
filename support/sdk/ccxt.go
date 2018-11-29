package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/interstellar/kelp/support/networking"
)

// Ccxt Rest SDK (https://github.com/franz-see/ccxt-rest, https://github.com/ccxt/ccxt/)
type Ccxt struct {
	httpClient   *http.Client
	ccxtBaseURL  string
	exchangeName string
	instanceName string
}

const pathExchanges = "/exchanges"

// MakeInitializedCcxtExchange constructs an instance of Ccxt that is bound to a specific exchange instance on the CCXT REST server
func MakeInitializedCcxtExchange(ccxtBaseURL string, exchangeName string) (*Ccxt, error) {
	if strings.HasSuffix(ccxtBaseURL, "/") {
		return nil, fmt.Errorf("invalid format for ccxtBaseURL: %s", ccxtBaseURL)
	}

	c := &Ccxt{
		httpClient:   http.DefaultClient,
		ccxtBaseURL:  ccxtBaseURL,
		exchangeName: exchangeName,
		// don't initialize instanceName since it's initialized in the call to init() below
	}
	e := c.init()
	if e != nil {
		return nil, fmt.Errorf("error when initializing Ccxt exchange: %s", e)
	}

	return c, nil
}

func (c *Ccxt) init() error {
	// get exchange list
	var exchangeList []string
	e := networking.JSONRequest(c.httpClient, "GET", c.ccxtBaseURL+pathExchanges, "", map[string]string{}, &exchangeList)
	if e != nil {
		return fmt.Errorf("error getting list of supported exchanges by CCXT: %s", e)
	}

	// validate that exchange name is in the exchange list
	exchangeListed := false
	for _, name := range exchangeList {
		if name == c.exchangeName {
			exchangeListed = true
			break
		}
	}
	if !exchangeListed {
		return fmt.Errorf("exchange name '%s' is not in the list of %d exchanges available", c.exchangeName, len(exchangeList))
	}

	// list all the instances of the exchange
	var instanceList []string
	e = networking.JSONRequest(c.httpClient, "GET", c.ccxtBaseURL+pathExchanges+"/"+c.exchangeName, "", map[string]string{}, &instanceList)
	if e != nil {
		return fmt.Errorf("error getting list of exchange instances for exchange '%s': %s", c.exchangeName, e)
	}

	// make a new instance if needed
	if len(instanceList) == 0 {
		instanceName := c.exchangeName + "1"
		e = c.newInstance(instanceName)
		if e != nil {
			return fmt.Errorf("error creating new instance '%s' for exchange '%s': %s", instanceName, c.exchangeName, e)
		}
		log.Printf("created new instance '%s' for exchange '%s'\n", instanceName, c.exchangeName)
		c.instanceName = instanceName
	} else {
		c.instanceName = instanceList[0]
	}

	// load markets to populate fields related to markets
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/loadMarkets"
	e = networking.JSONRequest(c.httpClient, "POST", url, "", map[string]string{}, nil)
	if e != nil {
		return fmt.Errorf("error loading markets for exchange instance (exchange=%s, instanceName=%s): %s", c.exchangeName, c.instanceName, e)
	}

	return nil
}

func (c *Ccxt) newInstance(instanceName string) error {
	data, e := json.Marshal(&struct {
		ID string `json:"id"`
	}{
		ID: instanceName,
	})
	if e != nil {
		return fmt.Errorf("error marshaling instanceName '%s' as ID for exchange '%s': %s", instanceName, c.exchangeName, e)
	}

	var newInstance map[string]interface{}
	e = networking.JSONRequest(c.httpClient, "POST", c.ccxtBaseURL+pathExchanges+"/"+c.exchangeName, string(data), map[string]string{}, &newInstance)
	if e != nil {
		return fmt.Errorf("error in web request when creating new exchange instance for exchange '%s': %s", c.exchangeName, e)
	}

	if _, ok := newInstance["urls"]; !ok {
		return fmt.Errorf("check for new instance of exchange '%s' failed for instanceName: %s", c.exchangeName, instanceName)
	}
	return nil
}

// symbolExists returns an error if the symbol does not exist
func (c *Ccxt) symbolExists(tradingPair string) error {
	// get list of symbols available on exchange
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var exchangeOutput interface{}
	e := networking.JSONRequest(c.httpClient, "GET", url, "", map[string]string{}, &exchangeOutput)
	if e != nil {
		return fmt.Errorf("error fetching details of exchange instance (exchange=%s, instanceName=%s): %s", c.exchangeName, c.instanceName, e)
	}

	exchangeMap := exchangeOutput.(map[string]interface{})
	if _, ok := exchangeMap["symbols"]; !ok {
		return fmt.Errorf("'symbols' field not in result of exchange details (exchange=%s, instanceName=%s)", c.exchangeName, c.instanceName)
	}

	symbolsList := exchangeMap["symbols"].([]interface{})
	for _, p := range symbolsList {
		symbol := p.(string)
		if tradingPair == symbol {
			// exists
			return nil
		}
	}
	return fmt.Errorf("trading pair '%s' does not exist in the list of %d symbols on exchange '%s'", tradingPair, len(symbolsList), c.exchangeName)
}

// FetchTicker calls the /fetchTicker endpoint on CCXT, trading pair is the CCXT version of the trading pair
func (c *Ccxt) FetchTicker(tradingPair string) (map[string]interface{}, error) {
	e := c.symbolExists(tradingPair)
	if e != nil {
		return nil, fmt.Errorf("symbol does not exist: %s", e)
	}

	// marshal input data
	data, e := json.Marshal(&[]string{tradingPair})
	if e != nil {
		return nil, fmt.Errorf("error marshaling tradingPair '%s' as an array for exchange '%s': %s", tradingPair, c.exchangeName, e)
	}

	// fetch ticker for symbol
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchTicker"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequest(c.httpClient, "POST", url, string(data), map[string]string{}, &output)
	if e != nil {
		return nil, fmt.Errorf("error fetching tickers for trading pair '%s': %s", tradingPair, e)
	}

	tickerMap := output.(map[string]interface{})
	return tickerMap, nil
}

// CcxtOrder represents an order in the orderbook
type CcxtOrder struct {
	Price  float64
	Amount float64
}

// FetchOrderBook calls the /fetchOrderBook endpoint on CCXT, trading pair is the CCXT version of the trading pair
func (c *Ccxt) FetchOrderBook(tradingPair string, limit *int) (map[string][]CcxtOrder, error) {
	e := c.symbolExists(tradingPair)
	if e != nil {
		return nil, fmt.Errorf("symbol does not exist: %s", e)
	}

	// marshal input data
	var data []byte
	if limit != nil {
		data, e = json.Marshal(&[]string{tradingPair, fmt.Sprintf("%d", *limit)})
		if e != nil {
			return nil, fmt.Errorf("error marshaling tradingPair '%s' as an array for exchange '%s' with limit=%d: %s", tradingPair, c.exchangeName, *limit, e)
		}
	} else {
		data, e = json.Marshal(&[]string{tradingPair})
		if e != nil {
			return nil, fmt.Errorf("error marshaling tradingPair '%s' as an array for exchange '%s' with no limit: %s", tradingPair, c.exchangeName, e)
		}
	}

	// fetch orderbook for symbol
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchOrderBook"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequest(c.httpClient, "POST", url, string(data), map[string]string{}, &output)
	if e != nil {
		return nil, fmt.Errorf("error fetching orderbook for trading pair '%s': %s", tradingPair, e)
	}

	result := map[string][]CcxtOrder{}
	tickerMap := output.(map[string]interface{})
	for k, v := range tickerMap {
		if k != "asks" && k != "bids" {
			continue
		}

		parsedList := []CcxtOrder{}
		// parse the list into the struct
		ordersList := v.([]interface{})
		for _, o := range ordersList {
			order := o.([]interface{})
			parsedList = append(parsedList, CcxtOrder{
				Price:  order[0].(float64),
				Amount: order[1].(float64),
			})
		}
		result[k] = parsedList
	}
	return result, nil
}

// CcxtTrade represents a trade
type CcxtTrade struct {
	Amount    float64 `json:"amount"`
	Cost      float64 `json:"cost"`
	Datetime  string  `json:"datetime"`
	ID        string  `json:"id"`
	Price     float64 `json:"price"`
	Side      string  `json:"side"`
	Symbol    string  `json:"symbol"`
	Timestamp int64   `json:"timestamp"`
}

// FetchTrades calls the /fetchTrades endpoint on CCXT, trading pair is the CCXT version of the trading pair
// TODO take in since and limit values to match CCXT's API
func (c *Ccxt) FetchTrades(tradingPair string) ([]CcxtTrade, error) {
	e := c.symbolExists(tradingPair)
	if e != nil {
		return nil, fmt.Errorf("symbol does not exist: %s", e)
	}

	// marshal input data
	data, e := json.Marshal(&[]string{tradingPair})
	if e != nil {
		return nil, fmt.Errorf("error marshaling input (tradingPair=%s) as an array for exchange '%s': %s", tradingPair, c.exchangeName, e)
	}

	// fetch trades for symbol
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchTrades"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	output := []CcxtTrade{}
	e = networking.JSONRequest(c.httpClient, "POST", url, string(data), map[string]string{}, &output)
	if e != nil {
		return nil, fmt.Errorf("error fetching trades for trading pair '%s': %s", tradingPair, e)
	}
	return output, nil
}
