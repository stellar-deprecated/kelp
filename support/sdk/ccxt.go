package sdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/utils"
)

// ccxtBaseURL should not have suffix of '/'
var ccxtBaseURL = "http://localhost:3000"

// SetBaseURL allows setting the base URL for ccxt
func SetBaseURL(baseURL string) error {
	ccxtBaseURL = strings.TrimSuffix(baseURL, "/")
	log.Printf("updated ccxtBaseURL to '%s'\n", ccxtBaseURL)
	return nil
}

// GetBaseURL returns the base URL for ccxt
func GetBaseURL() string {
	return ccxtBaseURL
}

// Ccxt Rest SDK (https://github.com/franz-see/ccxt-rest, https://github.com/ccxt/ccxt/)
type Ccxt struct {
	httpClient   *http.Client
	exchangeName string
	instanceName string
	markets      map[string]CcxtMarket
	headersMap   map[string]networking.HeaderFn
}

// CcxtMarket represents the result of a LoadMarkets call
type CcxtMarket struct {
	// only contains currently needed data
	Symbol string `json:"symbol"`
	Base   string `json:"base"`
	Quote  string `json:"quote"`
	Limits struct {
		Amount struct {
			Min float64 `json:"min"`
		} `json:"amount"`
		Price struct {
			Min float64 `json:"min"`
		} `json:"price"`
		Cost struct {
			Min float64 `json:"min"`
		} `json:"cost"`
	} `json:"limits"`
	Precision struct {
		Amount int8 `json:"amount"`
		Price  int8 `json:"price"`
	} `json:"precision"`
}

const pathExchanges = "/exchanges"

// MakeInitializedCcxtExchange constructs an instance of Ccxt that is bound to a specific exchange instance on the CCXT REST server
func MakeInitializedCcxtExchange(exchangeName string, apiKey api.ExchangeAPIKey, params []api.ExchangeParam, headers []api.ExchangeHeader) (*Ccxt, error) {
	if strings.HasSuffix(ccxtBaseURL, "/") {
		return nil, fmt.Errorf("invalid format for ccxtBaseURL: %s", ccxtBaseURL)
	}

	instanceName, e := makeInstanceName(exchangeName, apiKey, params, headers)
	if e != nil {
		return nil, fmt.Errorf("cannot make instance name: %s", e)
	}
	c := &Ccxt{
		httpClient:   http.DefaultClient,
		exchangeName: exchangeName,
		instanceName: instanceName,
	}

	e = c.initialize(apiKey, params, headers)
	if e != nil {
		return nil, fmt.Errorf("error when initializing Ccxt exchange: %s", e)
	}

	return c, nil
}

// exchangeList contains a list of supported exchanges
var exchangeList *[]string

// GetExchangeList gets a list of all supported exchanges
func GetExchangeList() []string {
	if exchangeList == nil {
		loadExchangeList()
	}
	return *exchangeList
}

func loadExchangeList() {
	var output []string
	e := networking.JSONRequest(http.DefaultClient, "GET", ccxtBaseURL+pathExchanges, "", map[string]string{}, &output, "error")
	if e != nil {
		eMsg1 := strings.Contains(e.Error(), "could not execute http request")
		eMsg2 := strings.Contains(e.Error(), ccxtBaseURL+"/exchanges: dial tcp")
		eMsg3 := strings.Contains(e.Error(), "connection refused")
		if eMsg1 && eMsg2 && eMsg3 {
			log.Printf("ccxt-rest is not running at %s so we cannot include those exchanges: %s", ccxtBaseURL, e.Error())
		} else {
			log.Printf("error getting list of supported exchanges at URL %s by CCXT so we cannot include those exchanges: %s", ccxtBaseURL, e)
		}
	}
	exchangeList = &output
}

func (c *Ccxt) initialize(apiKey api.ExchangeAPIKey, params []api.ExchangeParam, headers []api.ExchangeHeader) error {
	// validate that exchange name is in the exchange list
	exchangeListed := false
	el := GetExchangeList()
	for _, name := range el {
		if name == c.exchangeName {
			exchangeListed = true
			break
		}
	}
	if !exchangeListed {
		return fmt.Errorf("exchange name '%s' is not in the list of %d exchanges available: %v", c.exchangeName, len(el), el)
	}

	// list all the instances of the exchange
	var instanceList []string
	e := networking.JSONRequest(c.httpClient, "GET", ccxtBaseURL+pathExchanges+"/"+c.exchangeName, "", map[string]string{}, &instanceList, "error")
	if e != nil {
		return fmt.Errorf("error getting list of exchange instances for exchange '%s': %s", c.exchangeName, e)
	}

	// make a new instance if needed
	if !c.hasInstance(instanceList) {
		e = c.newInstance(apiKey, params)
		if e != nil {
			return fmt.Errorf("error creating new instance '%s' for exchange '%s': %s", c.instanceName, c.exchangeName, e)
		}
		log.Printf("created new instance '%s' for exchange '%s'\n", c.instanceName, c.exchangeName)
	} else {
		log.Printf("instance '%s' for exchange '%s' already exists\n", c.instanceName, c.exchangeName)
	}

	// load markets to populate fields related to markets
	var marketsResponse interface{}
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/loadMarkets"
	e = networking.JSONRequest(c.httpClient, "POST", url, "", map[string]string{}, &marketsResponse, "error")
	if e != nil {
		return fmt.Errorf("error loading markets for exchange instance (exchange=%s, instanceName=%s): %s", c.exchangeName, c.instanceName, e)
	}
	// decode markets and sets it on the ccxt instance
	var markets map[string]CcxtMarket
	e = mapstructure.Decode(marketsResponse, &markets)
	if e != nil {
		return fmt.Errorf("error converting loadMarkets output to a map of Market for exchange instance (exchange=%s, instanceName=%s): %s", c.exchangeName, c.instanceName, e)
	}
	c.markets = markets

	headersMap := map[string]networking.HeaderFn{}
	ccxtHeaderMappings := makeHeaderMappingsFromNewTimestamp()
	for _, header := range headers {
		headerFn, e := networking.MakeHeaderFn(header.Value, ccxtHeaderMappings)
		if e != nil {
			return fmt.Errorf("unable to make header function with key (%s) and value (%s): %s", header.Header, header.Value, e)
		}
		headersMap[header.Header] = headerFn
	}
	c.headersMap = headersMap

	return nil
}

// makeInstanceName takes all those inputs that create a distinctly initialized instance
func makeInstanceName(exchangeName string, apiKey api.ExchangeAPIKey, params []api.ExchangeParam, headers []api.ExchangeHeader) (string, error) {
	keyHash := ""
	if apiKey.Key != "" {
		keyHashNum, e := utils.HashString(apiKey.Key)
		if e != nil {
			return "", fmt.Errorf("could not hash apiKey.Key: %s", e)
		}
		keyHash = fmt.Sprintf("%d", keyHashNum)
	}

	paramsHash := ""
	if len(params) > 0 {
		paramsHashNum, e := utils.ToJSONHash(params)
		if e != nil {
			s := fmt.Sprintf("%v", params)
			return "", fmt.Errorf("could not hash params (%s): %s", s, e)
		}
		paramsHash = fmt.Sprintf("%d", paramsHashNum)
	}

	headersHash := ""
	if len(headers) > 0 {
		headersHashNum, e := utils.ToJSONHash(headers)
		if e != nil {
			s := fmt.Sprintf("%v", headers)
			return "", fmt.Errorf("could not hash headers (%s): %s", s, e)
		}
		headersHash = fmt.Sprintf("%d", headersHashNum)
	}

	return fmt.Sprintf("%s_%s_%s_%s", exchangeName, keyHash, paramsHash, headersHash), nil
}

func (c *Ccxt) hasInstance(instanceList []string) bool {
	for _, i := range instanceList {
		if i == c.instanceName {
			return true
		}
	}
	return false
}

func (c *Ccxt) newInstance(apiKey api.ExchangeAPIKey, params []api.ExchangeParam) error {
	// this is a map of string to interface{} becuase the param can be of type string, number, or bool
	data := map[string]interface{}{
		"id":     c.instanceName,
		"apiKey": apiKey.Key,
		"secret": apiKey.Secret,
	}
	// values that occur later in the list will override previous values (this is by design, so default values can be overriden by config values)
	for _, param := range params {
		data[param.Param] = param.Value
	}
	jsonData, e := json.Marshal(data)
	if e != nil {
		return fmt.Errorf("error marshaling instanceName '%s' as ID for exchange '%s': %s", c.instanceName, c.exchangeName, e)
	}

	var newInstance map[string]interface{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", ccxtBaseURL+pathExchanges+"/"+c.exchangeName, string(jsonData), c.headersMap, &newInstance, "error")
	if e != nil {
		return fmt.Errorf("error in web request when creating new exchange instance for exchange '%s': %s", c.exchangeName, e)
	}

	if _, ok := newInstance["urls"]; !ok {
		return fmt.Errorf("check for new instance of exchange '%s' failed for instanceName: %s", c.exchangeName, c.instanceName)
	}
	return nil
}

// symbolExists returns an error if the symbol does not exist
func (c *Ccxt) symbolExists(tradingPair string) error {
	if _, ok := c.markets[tradingPair]; ok {
		log.Printf("found trading pair symbol '%s' in markets map", tradingPair)
		return nil
	}

	// get list of symbols available on exchange
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var exchangeOutput interface{}
	e := networking.JSONRequestDynamicHeaders(c.httpClient, "GET", url, "", c.headersMap, &exchangeOutput, "error")
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

// GetMarket returns the CcxtMarket instance
func (c *Ccxt) GetMarket(tradingPair string) *CcxtMarket {
	if v, ok := c.markets[tradingPair]; ok {
		return &v
	}
	return nil
}

// GetMarkets returns all the markets
func (c *Ccxt) GetMarkets() map[string]CcxtMarket {
	return c.markets
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
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchTicker"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
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
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchOrderBook"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
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
	Amount    float64     `json:"amount"`
	Cost      float64     `json:"cost"`
	Datetime  string      `json:"datetime"`
	ID        string      `json:"id"`   // tradeID
	Info      interface{} `json:"info"` // raw trade response gotten from the exchange site's API
	Price     float64     `json:"price"`
	Side      string      `json:"side"`
	Symbol    string      `json:"symbol"`
	Timestamp int64       `json:"timestamp"`
	Fee       struct {
		Cost     float64 `json:"cost"`
		Currency string  `json:"currency"`
	} `json:"fee"`
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
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchTrades"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	output := []CcxtTrade{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
	if e != nil {
		return nil, fmt.Errorf("error fetching trades for trading pair '%s': %s", tradingPair, e)
	}
	return output, nil
}

// FetchMyTrades calls the /fetchMyTrades endpoint on CCXT, trading pair is the CCXT version of the trading pair
func (c *Ccxt) FetchMyTrades(tradingPair string, limit int, maybeCursorStart interface{}) ([]CcxtTrade, error) {
	e := c.symbolExists(tradingPair)
	if e != nil {
		return nil, fmt.Errorf("symbol does not exist: %s", e)
	}

	// marshal input data
	var data []byte
	if maybeCursorStart == nil {
		data, e = json.Marshal(&[]string{tradingPair, strconv.Itoa(limit)})
		if e != nil {
			return nil, fmt.Errorf("error marshaling input (tradingPair=%s) as an array for exchange '%s': %s", tradingPair, c.exchangeName, e)
		}
	} else {
		cursorString := fmt.Sprintf("%v", maybeCursorStart)
		data, e = json.Marshal(&[]string{tradingPair, cursorString, strconv.Itoa(limit)})
		if e != nil {
			return nil, fmt.Errorf("error marshaling input (tradingPair=%s, maybeCursorStart=%v) as an array for exchange '%s': %s", tradingPair, maybeCursorStart, c.exchangeName, e)
		}
	}

	// fetch trades for symbol
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchMyTrades"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	output := []CcxtTrade{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
	if e != nil {
		return nil, fmt.Errorf("error fetching trades for trading pair '%s': %s", tradingPair, e)
	}
	return output, nil
}

// CcxtBalance represents the balance for an asset
type CcxtBalance struct {
	Total float64
	Used  float64
	Free  float64
}

// FetchBalance calls the /fetchBalance endpoint on CCXT
func (c *Ccxt) FetchBalance() (map[string]CcxtBalance, error) {
	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchBalance"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e := networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, "", c.headersMap, &output, "error")
	if e != nil {
		return nil, fmt.Errorf("error fetching balance: %s", e)
	}

	outputMap := output.(map[string]interface{})
	if _, ok := outputMap["total"]; !ok {
		return nil, fmt.Errorf("result from call to fetchBalance did not contain 'total' field: %v", output)
	}
	totals := outputMap["total"].(map[string]interface{})

	result := map[string]CcxtBalance{}
	for asset, v := range totals {
		var totalBalance float64
		if b, ok := v.(float64); ok {
			totalBalance = b
		} else {
			return nil, fmt.Errorf("could not convert total balance for asset '%s' from interface{} to float64", asset)
		}
		if totalBalance == 0 {
			continue
		}

		assetData := outputMap[asset].(map[string]interface{})
		var assetBalance CcxtBalance
		e = mapstructure.Decode(assetData, &assetBalance)
		if e != nil {
			return nil, fmt.Errorf("error converting balance map to CcxtBalance for asset '%s': %s", asset, e)
		}
		result[asset] = assetBalance
	}
	return result, nil
}

// CcxtOpenOrder represents an open order
type CcxtOpenOrder struct {
	Amount    float64
	Cost      float64
	Filled    float64
	ID        string
	Price     float64
	Side      string
	Status    string
	Symbol    string
	Type      string
	Timestamp int64
}

// FetchOpenOrders calls the /fetchOpenOrders endpoint on CCXT
func (c *Ccxt) FetchOpenOrders(tradingPairs []string) (map[string][]CcxtOpenOrder, error) {
	for _, p := range tradingPairs {
		e := c.symbolExists(p)
		if e != nil {
			return nil, fmt.Errorf("symbol does not exist: %s", e)
		}
	}

	// marshal input data
	data, e := json.Marshal(&tradingPairs)
	if e != nil {
		return nil, fmt.Errorf("error marshaling input (tradingPairs=%v) for exchange '%s': %s", tradingPairs, c.exchangeName, e)
	}

	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/fetchOpenOrders"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
	if e != nil {
		return nil, fmt.Errorf("error fetching open orders: %s", e)
	}

	result := map[string][]CcxtOpenOrder{}
	outputList := output.([]interface{})
	for _, elem := range outputList {
		elemMap, ok := elem.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not convert the element in the result to a map[string]interface{}, type = %s", reflect.TypeOf(elem))
		}

		var openOrder CcxtOpenOrder
		e = mapstructure.Decode(elemMap, &openOrder)
		if e != nil {
			return nil, fmt.Errorf("could not decode open order element (%v): %s", elemMap, e)
		}

		var orderList []CcxtOpenOrder
		if l, ok := result[openOrder.Symbol]; ok {
			orderList = l
		} else {
			orderList = []CcxtOpenOrder{}
		}

		orderList = append(orderList, openOrder)
		result[openOrder.Symbol] = orderList
	}
	return result, nil
}

// CreateLimitOrder calls the /createOrder endpoint on CCXT with a limit price and the order type set to "limit"
func (c *Ccxt) CreateLimitOrder(tradingPair string, side string, amount float64, price float64, maybeExchangeSpecificParams interface{}) (*CcxtOpenOrder, error) {
	orderType := "limit"
	e := c.symbolExists(tradingPair)
	if e != nil {
		return nil, fmt.Errorf("symbol does not exist: %s", e)
	}

	// marshal input data
	inputData := []interface{}{
		tradingPair,
		orderType,
		side,
		amount,
		price,
	}
	if maybeExchangeSpecificParams != nil {
		inputData = append(inputData, maybeExchangeSpecificParams)
	}
	data, e := json.Marshal(&inputData)
	if e != nil {
		return nil, fmt.Errorf("error marshaling input (%v) for exchange '%s': %s", inputData, c.exchangeName, e)
	}

	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/createOrder"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
	if e != nil {
		return nil, fmt.Errorf("error creating order: %s", e)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not convert the output to a map[string]interface{}, type = %s", reflect.TypeOf(output))
	}

	var openOrder CcxtOpenOrder
	e = mapstructure.Decode(outputMap, &openOrder)
	if e != nil {
		return nil, fmt.Errorf("could not decode outputMap to openOrder (%v): %s", outputMap, e)
	}

	return &openOrder, nil
}

// CancelOrder calls the /cancelOrder endpoint on CCXT with the orderID and tradingPair
func (c *Ccxt) CancelOrder(orderID string, tradingPair string) (*CcxtOpenOrder, error) {
	e := c.symbolExists(tradingPair)
	if e != nil {
		return nil, fmt.Errorf("symbol does not exist: %s", e)
	}

	// marshal input data
	inputData := []interface{}{
		orderID,
		tradingPair,
	}
	data, e := json.Marshal(&inputData)
	if e != nil {
		return nil, fmt.Errorf("error marshaling input (%v) for exchange '%s': %s", inputData, c.exchangeName, e)
	}

	url := ccxtBaseURL + pathExchanges + "/" + c.exchangeName + "/" + c.instanceName + "/cancelOrder"
	// decode generic data (see "https://blog.golang.org/json-and-go#TOC_4.")
	var output interface{}
	e = networking.JSONRequestDynamicHeaders(c.httpClient, "POST", url, string(data), c.headersMap, &output, "error")
	if e != nil {
		return nil, fmt.Errorf("error canceling order: %s", e)
	}

	outputMap, ok := output.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not convert the output to a map[string]interface{}, type = %s", reflect.TypeOf(output))
	}

	var openOrder CcxtOpenOrder
	e = mapstructure.Decode(outputMap, &openOrder)
	if e != nil {
		return nil, fmt.Errorf("could not decode outputMap to openOrder (%v): %s", outputMap, e)
	}

	return &openOrder, nil
}
