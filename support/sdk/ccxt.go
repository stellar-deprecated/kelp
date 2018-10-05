package sdk

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lightyeario/kelp/support/utils"
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
	e := utils.JSONRequest(c.httpClient, "GET", c.ccxtBaseURL+pathExchanges, "", map[string]string{}, &exchangeList)
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
	e = utils.JSONRequest(c.httpClient, "GET", c.ccxtBaseURL+pathExchanges+"/"+c.exchangeName, "", map[string]string{}, &instanceList)
	if e != nil {
		return fmt.Errorf("error getting list of exchange instances for exchange '%s': %s", c.exchangeName, e)
	}

	// make a new instance if needed
	if len(instanceList) == 0 {
		instanceName := c.exchangeName + "1"
		var newInstance map[string]interface{}
		// TODO better JSON structure
		e = utils.JSONRequest(c.httpClient, "POST", c.ccxtBaseURL+pathExchanges+"/"+c.exchangeName, "{\"id\": \""+instanceName+"\"}", map[string]string{}, &newInstance)
		if e != nil {
			return fmt.Errorf("error creating new exchange instance for exchange '%s': %s", c.exchangeName, e)
		}
		if _, ok := newInstance["urls"]; !ok {
			return fmt.Errorf("unable to create a new instance of exchange '%s' with instanceName: %s", c.exchangeName, instanceName)
		}
		c.instanceName = instanceName
	} else {
		c.instanceName = instanceList[0]
	}

	// load markets to populate fields related to markets
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/loadMarkets"
	e = utils.JSONRequest(c.httpClient, "POST", url, "", map[string]string{}, nil)
	if e != nil {
		return fmt.Errorf("error loading markets for exchange instance (exchange=%s, instanceName=%s): %s", c.exchangeName, c.instanceName, e)
	}

	return nil
}

// FetchTicker calls the /fetchTicker endpoint on CCXT, trading pair is the CCXT version of the trading pair
func (c *Ccxt) FetchTicker(tradingPair string) (map[string]interface{}, error) {
	// get list of symbols available on exchange
	url := c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var exchangeOutput interface{}
	e := utils.JSONRequest(c.httpClient, "GET", url, "", map[string]string{}, &exchangeOutput)
	if e != nil {
		return nil, fmt.Errorf("error fetching details of exchange instance (exchange=%s, instanceName=%s): %s", c.exchangeName, c.instanceName, e)
	}

	exchangeMap := exchangeOutput.(map[string]interface{})
	if _, ok := exchangeMap["symbols"]; !ok {
		return nil, fmt.Errorf("'symbols' field not in result of exchange details (exchange=%s, instanceName=%s)", c.exchangeName, c.instanceName)
	}
	symbolsList := exchangeMap["symbols"].([]interface{})
	symbolExists := false
	for _, p := range symbolsList {
		symbol := p.(string)
		if tradingPair == symbol {
			symbolExists = true
			break
		}
	}
	if !symbolExists {
		return nil, fmt.Errorf("trading pair '%s' does not exist in the list of %d symbols on exchange '%s'", tradingPair, len(symbolsList), c.exchangeName)
	}

	// fetch ticker for symbol
	url = c.ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchTicker"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var tickerOutput interface{}
	// TODO better JSON structure
	e = utils.JSONRequest(c.httpClient, "POST", url, "[\""+tradingPair+"\"]", map[string]string{}, &tickerOutput)
	if e != nil {
		return nil, fmt.Errorf("error fetching tickers for trading pair '%s': %s", tradingPair, e)
	}

	tickerMap := tickerOutput.(map[string]interface{})
	return tickerMap, nil
}
