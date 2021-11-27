package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/sdk"
)

const (
	STREAM_TICKER_FMT = "%s@ticker"
	STREAM_BOOK_FMT   = "%s@depth"

	//not from binance docs, just for convetion we use @streamName
	STREAM_USER = "@user"
	// key used to save last cursor in events, it must be something that couldn't be used in the map
	LAST_CURSOR_KEY        = STREAM_USER + "||lastCursor"
	TTLTIME                = time.Second * 3 // ttl time in seconds
	EVENT_EXECUTION_REPORT = "executionReport"
)

var (
	timeWaitForFirstEvent = time.Second * 2
)

var (
	ErrConversionWsMarketEvent       = errConversion{from: "interface", to: "*binance.WsMarketStatEvent"}
	ErrConversionWsPartialDepthEvent = errConversion{from: "interface", to: "*binance.WsPartialDepthEvent"}
	ErrConversionHistory             = errConversion{from: "interface", to: "History"}
	ErrConversionCursor              = errConversion{from: "interface", to: "int64"}
)

type History []*sdk.EventExecutionReport

type Subscriber func(symbol string, state *mapEvents) (*stream, error)
type errMissingSymbol struct {
	symbol string
}

func (err errMissingSymbol) Error() string {
	return fmt.Sprintf("Symbol %s is missing from exchange intizialization", err.symbol)
}

type errConversion struct {
	from string
	to   string
}

func (err errConversion) Error() string {
	return fmt.Sprintf("Error conversion from %s to %s", err.from, err.to)
}

type stream struct {
	doneC   chan struct{}
	stopC   chan struct{}
	cleanup func()
}

//Wait until the stream ends
func (s stream) Wait() {

	if s.doneC == nil {
		return
	}

	<-s.doneC
}

//Close the stream and cleanup any data
func (s stream) Close() {
	if s.stopC == nil {
		return
	}
	s.stopC <- struct{}{}
	s.stopC = nil

	if s.cleanup != nil {
		s.cleanup()
	}
}

//mapData... struct used to data from events and timestamp when they are cached
type mapData struct {
	data      interface{}
	err       error
	createdAt time.Time
}

//isStatle... check if data it's stale
func isStale(data mapData, ttl time.Duration) bool {

	return time.Now().Sub(data.createdAt).Seconds() > ttl.Seconds()
}

//struct used to cache events
type mapEvents struct {
	data map[string]mapData
	mtx  *sync.RWMutex
}

//Set ... set value
func (m *mapEvents) Set(key string, data interface{}, err error) {

	now := time.Now()

	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.data[key] = mapData{
		data:      data,
		createdAt: now,
		err:       err,
	}

}

//Get ... get value
func (m *mapEvents) Get(key string) (mapData, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	data, isData := m.data[key]

	return data, isData
}

//Del ... delete cached value
func (m *mapEvents) Del(key string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	delete(m.data, key)

}

// create new map for cache
func makeMapEvents() *mapEvents {
	return &mapEvents{
		data: make(map[string]mapData),
		mtx:  &sync.RWMutex{},
	}
}

//struct used to keep all cached data
type events struct {
	SymbolStats        *mapEvents
	BookStats          *mapEvents
	TradeHistoryEvents *mapEvents
}

func createStateEvents() *events {
	events := &events{
		SymbolStats:        makeMapEvents(),
		BookStats:          makeMapEvents(),
		TradeHistoryEvents: makeMapEvents(),
	}

	return events
}

// 24hr rolling window ticker statistics for a single symbol. These are NOT the statistics of the UTC day, but a 24hr rolling window for the previous 24hrs.
// Stream Name: <symbol>@ticker
// Update Speed: 1000ms
func subcribeTicker(symbol string, state *mapEvents) (*stream, error) {

	wsMarketStatHandler := func(ticker *binance.WsMarketStatEvent) {
		state.Set(symbol, ticker, nil)
	}

	errHandler := func(err error) {
		log.Printf("Error WsMarketsStat for symbol %s: %v\n", symbol, err)
	}

	doneC, stopC, err := binance.WsMarketStatServe(symbol, wsMarketStatHandler, errHandler)

	if err != nil {
		return nil, err
	}

	keepConnection(doneC, func() {
		subcribeTicker(symbol, state)
	})

	return &stream{doneC: doneC, stopC: stopC, cleanup: func() {
		state.Del(symbol)
	}}, err

}

func subcribeUserStream(listenKey string, state *mapEvents) (*stream, error) {

	userStreamLock := sync.Mutex{}

	wsUserStreamExecutinReportHandler := func(message []byte) {

		event := &sdk.EventExecutionReport{}
		err := json.Unmarshal(message, event)

		if err != nil {
			log.Printf("Error unmarshal %s to eventExecutionReport\n", string(message))
			return
		}

		userStreamLock.Lock()
		defer userStreamLock.Unlock()

		history, isHistory := state.Get(event.Symbol)

		if !isHistory {
			history.data = make(History, 0)
			state.Set(event.Symbol, history, nil)
		}

		now := time.Now()
		history.createdAt = now

		data, isOk := history.data.(History)

		if !isOk {
			log.Printf("Error conversion %v\n", ErrConversionHistory)
			state.Set(event.Symbol, history, ErrConversionHistory)
			return
		}

		history.data = append(data, event)

		lastCursor := event.TransactionTime

		lastCursorData, isCursor := state.Get(LAST_CURSOR_KEY)

		if isCursor {

			cursor, isOk := lastCursorData.data.(int64)

			if isOk {
				if cursor > lastCursor {
					lastCursor = cursor
				}
			} else {
				log.Printf("Error converting cursor %v\n", ErrConversionCursor)
				err = ErrConversionCursor
			}
		}

		state.Set(LAST_CURSOR_KEY, lastCursor, err)
	}

	wsUserStreamHandler := func(message []byte) {
		event := &sdk.EventBinance{}
		err := json.Unmarshal(message, event)

		if err != nil {
			log.Printf("Error unmarshal %s to eventBinance\n", string(message))
			return
		}

		switch event.Name {
		case EVENT_EXECUTION_REPORT:
			wsUserStreamExecutinReportHandler(message)
		}

	}

	errHandler := func(err error) {
		log.Printf("Error WsUserDataServe for listenKey %s: %v\n", listenKey, err)
	}

	doneC, stopC, err := binance.WsUserDataServe(listenKey, wsUserStreamHandler, errHandler)

	if err != nil {
		return nil, fmt.Errorf("error creating WsUserDataService:%s", err)
	}

	keepConnection(doneC, func() {
		subcribeUserStream(listenKey, state)
	})

	return &stream{doneC: doneC, stopC: stopC}, err

}

//restart Connection with ws
// Binance close each connection after 24 hours
func keepConnection(doneC chan struct{}, reconnect func()) {

	go func() {
		<-doneC
		reconnect()
	}()
}

// Top <levels> bids and asks, pushed every second. Valid <levels> are 5, 10, or 20.
// <symbol>@depth<levels>@100ms
// 100ms
func subcribeBook(symbol string, state *mapEvents) (*stream, error) {

	wsPartialDepthHandler := func(event *binance.WsPartialDepthEvent) {
		state.Set(symbol, event, nil)
	}

	errHandler := func(err error) {
		log.Printf("Error WsPartialDepthServe for symbol %s: %v\n", symbol, err)
	}

	//Subscribe to highest level
	doneC, stopC, err := binance.WsPartialDepthServe100Ms(symbol, "20", wsPartialDepthHandler, errHandler)

	if err != nil {
		return nil, err
	}

	keepConnection(doneC, func() {
		subcribeBook(symbol, state)
	})

	return &stream{doneC: doneC, stopC: stopC, cleanup: func() {
		state.Del(symbol)
	}}, err

}

type binanceExchangeWs struct {
	events *events

	streams    map[string]*stream
	streamLock *sync.Mutex

	assetConverter model.AssetConverterInterface
	delimiter      string

	client    *binance.Client
	listenKey string

	keys api.ExchangeAPIKey

	errUserStream error
}

// makeBinanceWs is a factory method to make an binance exchange over ws
func makeBinanceWs(keys api.ExchangeAPIKey) (*binanceExchangeWs, error) {

	binance.WebsocketKeepalive = true

	events := createStateEvents()

	streams := make(map[string]*stream)

	beWs := &binanceExchangeWs{
		events:         events,
		delimiter:      "",
		assetConverter: model.CcxtAssetConverter,
		streamLock:     &sync.Mutex{},
		streams:        streams,
		keys:           keys,
	}

	return beWs, nil
}

//ListenKey expires every 60 minutes
func (beWs *binanceExchangeWs) keepAliveStreamService(client *binance.Client, key string) {

	for {
		time.Sleep(time.Minute * 50)
		err := client.NewKeepaliveUserStreamService().ListenKey(key).Do(context.Background())

		if err != nil {
			log.Printf("Error keepAliveStreamService %v\n", err)
		}

		beWs.errUserStream = err
	}
}

func (beWs *binanceExchangeWs) isSubscribedUserStream() bool {

	_, isStream := beWs.streams[STREAM_USER]

	return isStream
}

func (beWs *binanceExchangeWs) subscribeUserStream() error {

	beWs.streamLock.Lock()
	defer beWs.streamLock.Unlock()

	if beWs.isSubscribedUserStream() {
		return nil
	}

	binanceClient := binance.NewClient(beWs.keys.Key, beWs.keys.Secret)

	listenKey, err := binanceClient.NewStartUserStreamService().Do(context.Background())

	if err != nil {
		return fmt.Errorf("error when creating listenKey: %s", err)
	}

	go beWs.keepAliveStreamService(binanceClient, listenKey)

	streamUser, err := subcribeUserStream(listenKey, beWs.events.TradeHistoryEvents)

	if err != nil {
		return fmt.Errorf("error when subscribing to user stream: %s", err)
	}

	beWs.streams[STREAM_USER] = streamUser
	beWs.client = binanceClient
	beWs.listenKey = listenKey

	//Wait for first
	time.Sleep(timeWaitForFirstEvent)

	return nil
}

//getPrceision... get precision for float string
func getPrecision(floatStr string) int8 {

	strs := strings.Split(floatStr, ".")

	if len(strs) != 2 {
		log.Printf("could not get precision for float %s\n", floatStr)
		return 0
	}

	return int8(len(strs[1]))
}

//subscribeStream and wait for the first event
func (beWs *binanceExchangeWs) subscribeStream(symbol, format string, subscribe Subscriber, state *mapEvents) (mapData, error) {

	stream, err := subscribe(symbol, state)

	streamName := fmt.Sprintf(format, symbol)

	if err != nil {
		return mapData{}, fmt.Errorf("error when subscribing for %s: %s", streamName, err)
	}

	//Store stream
	beWs.streamLock.Lock()
	beWs.streams[streamName] = stream
	beWs.streamLock.Unlock()

	//Wait for binance to send events
	time.Sleep(timeWaitForFirstEvent)

	data, isStream := state.Get(symbol)

	//We couldn't subscribe for this pair
	if !isStream {
		return mapData{}, fmt.Errorf("error data doesn't exist for %s", streamName)
	}

	return data, nil
}

// GetTickerPrice impl.
func (beWs *binanceExchangeWs) GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]api.Ticker, error) {

	priceResult := map[model.TradingPair]api.Ticker{}
	for _, p := range pairs {

		symbol, err := p.ToString(beWs.assetConverter, beWs.delimiter)

		if err != nil {
			return nil, err
		}

		tickerData, isTicker := beWs.events.SymbolStats.Get(symbol)

		if !isTicker {
			tickerData, err = beWs.subscribeStream(symbol, STREAM_TICKER_FMT, subcribeTicker, beWs.events.SymbolStats)

			if err != nil {
				return nil, fmt.Errorf("error when subscribing to stream %s:%s", fmt.Sprintf(STREAM_TICKER_FMT, symbol), err)
			}

		}

		//Show how old is the ticker
		log.Printf("Ticker for %s is %d milliseconds old!\n", symbol, time.Now().Sub(tickerData.createdAt).Milliseconds())

		if isStale(tickerData, TTLTIME) {
			return nil, fmt.Errorf("ticker for %s symbols is older than %v", symbol, TTLTIME)
		}

		tickerI := tickerData.data

		//Convert to WsMarketStatEvent
		ticker, isOk := tickerI.(*binance.WsMarketStatEvent)

		if !isOk {
			return nil, ErrConversionWsMarketEvent
		}

		askPrice, e := strconv.ParseFloat(ticker.AskPrice, 64)
		if e != nil {
			return nil, fmt.Errorf("unable to correctly parse 'ask': %s", e)
		}
		bidPrice, e := strconv.ParseFloat(ticker.BidPrice, 64)
		if e != nil {
			return nil, fmt.Errorf("unable to correctly parse 'bid': %s", e)
		}
		lastPrice, e := strconv.ParseFloat(ticker.LastPrice, 64)
		if e != nil {
			return nil, fmt.Errorf("unable to correctly parse 'last': %s", e)
		}

		priceResult[p] = api.Ticker{
			AskPrice:  model.NumberFromFloat(askPrice, getPrecision(ticker.AskPrice)),
			BidPrice:  model.NumberFromFloat(bidPrice, getPrecision(ticker.BidPrice)),
			LastPrice: model.NumberFromFloat(lastPrice, getPrecision(ticker.LastPrice)),
		}
	}

	return priceResult, nil
}

//GetOrderBook impl
func (beWs *binanceExchangeWs) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {

	var (
		fetchSize = int(maxCount)
	)

	if fetchSize > 20 {
		return nil, fmt.Errorf("Max supported depth level is 20")
	}

	symbol, err := pair.ToString(beWs.assetConverter, beWs.delimiter)
	if err != nil {
		return nil, fmt.Errorf("error converting pair to string: %s", err)
	}

	bookData, isBook := beWs.events.BookStats.Get(symbol)

	if !isBook {

		bookData, err = beWs.subscribeStream(symbol, STREAM_BOOK_FMT, subcribeBook, beWs.events.BookStats)

		if err != nil {
			return nil, err
		}

	}

	if bookData.err != nil {
		return nil, fmt.Errorf("error from stream:%v", bookData.err)
	}

	//Show how old is the orderbook
	log.Printf("OrderBook for %s is %d milliseconds old!\n", symbol, time.Now().Sub(bookData.createdAt).Milliseconds())

	if isStale(bookData, TTLTIME) {
		return nil, fmt.Errorf("ticker for %s symbols is older than %v", symbol, TTLTIME)
	}

	bookI := bookData.data

	//Convert to WsMarketStatEvent
	book, isOk := bookI.(*binance.WsPartialDepthEvent)

	if !isOk {
		return nil, ErrConversionWsPartialDepthEvent
	}

	askCcxtOrders := book.Asks
	bidCcxtOrders := book.Bids

	if len(askCcxtOrders) > fetchSize {
		askCcxtOrders = askCcxtOrders[:fetchSize]

	}

	if len(bidCcxtOrders) > fetchSize {
		bidCcxtOrders = bidCcxtOrders[:fetchSize]
	}

	asks, err := beWs.readOrders(askCcxtOrders, pair, model.OrderActionSell)

	if err != nil {
		return nil, fmt.Errorf("error when reading ask orders:%s", err)
	}

	bids, err := beWs.readOrders(bidCcxtOrders, pair, model.OrderActionBuy)

	if err != nil {
		return nil, fmt.Errorf("error when reading bid orders:%s", err)
	}

	return model.MakeOrderBook(pair, asks, bids), nil
}

//readOrders... transform orders from binance to model.Order
func (beWs *binanceExchangeWs) readOrders(orders []common.PriceLevel, pair *model.TradingPair, orderAction model.OrderAction) ([]model.Order, error) {

	pricePrecision := getPrecision(orders[0].Price)
	volumePrecision := getPrecision(orders[0].Quantity)

	result := []model.Order{}
	for _, o := range orders {

		price, quantity, err := o.Parse()

		if err != nil {
			return nil, err
		}

		result = append(result, model.Order{
			Pair:        pair,
			OrderAction: orderAction,
			OrderType:   model.OrderTypeLimit,
			Price:       model.NumberFromFloat(price, pricePrecision),
			Volume:      model.NumberFromFloat(quantity, volumePrecision),
			Timestamp:   nil,
		})
	}
	return result, nil
}

// GetTradeHistory impl
func (beWs *binanceExchangeWs) GetTradeHistory(pair model.TradingPair, maybeCursorStart interface{}, maybeCursorEnd interface{}) (*api.TradeHistoryResult, error) {

	if beWs.errUserStream != nil {
		return nil, fmt.Errorf("error from update listen key:%v", beWs.errUserStream)
	}

	if !beWs.isSubscribedUserStream() {
		if err := beWs.subscribeUserStream(); err != nil {
			return nil, fmt.Errorf("error subscribing to user stream: %s", err)
		}
	}

	symbol, err := pair.ToString(beWs.assetConverter, beWs.delimiter)
	if err != nil {
		return nil, fmt.Errorf("error converting symbol to string: %s", err)
	}

	data, isOrders := beWs.events.TradeHistoryEvents.Get(symbol)

	if !isOrders {
		return nil, fmt.Errorf("no trade history for trading pair '%s'", symbol)
	}

	if data.err != nil {
		return nil, fmt.Errorf("error from stream:%v", data.err)
	}

	history, isOk := data.data.(History)

	if !isOk {
		return nil, ErrConversionHistory
	}

	trades := []model.Trade{}
	for _, raw := range history {

		t, err := beWs.readTrade(&pair, symbol, raw)
		if err != nil {
			return nil, fmt.Errorf("error while reading trade: %s", err)
		}

		t.OrderID = fmt.Sprintf("%d", raw.OrderID)

		trades = append(trades, *t)
	}

	sort.Sort(model.TradesByTsID(trades))
	cursor := maybeCursorStart

	if len(trades) > 0 {
		cursor, err = beWs.getCursor(trades)
		if err != nil {
			return nil, fmt.Errorf("error getting cursor when fetching trades: %s", err)
		}
	}

	return &api.TradeHistoryResult{
		Cursor: cursor,
		Trades: trades,
	}, nil
}

func (beWs *binanceExchangeWs) getCursor(trades []model.Trade) (interface{}, error) {
	lastTrade := trades[len(trades)-1]

	lastCursor := lastTrade.Order.Timestamp.AsInt64()
	// add 1 to lastCursor so we don't repeat the same cursor on the next run
	fetchedCursor := strconv.FormatInt(lastCursor+1, 10)

	// update cursor accordingly
	return fetchedCursor, nil
}

// GetLatestTradeCursor impl.
func (beWs *binanceExchangeWs) GetLatestTradeCursor() (interface{}, error) {

	if beWs.errUserStream != nil {
		return nil, fmt.Errorf("error from update listen key:%v", beWs.errUserStream)
	}

	if !beWs.isSubscribedUserStream() {
		if err := beWs.subscribeUserStream(); err != nil {
			return nil, fmt.Errorf("error subscribing to user stream: %s", err)
		}
	}

	lastTradeCursor, isCursor := beWs.events.TradeHistoryEvents.Get(LAST_CURSOR_KEY)

	if !isCursor {
		return nil, errors.New("Missing cursor")
	}

	if lastTradeCursor.err != nil {
		return nil, fmt.Errorf("error from stream:%v", lastTradeCursor.err)
	}

	cursor, isOk := lastTradeCursor.data.(int64)

	if !isOk {
		return nil, ErrConversionCursor
	}

	return fmt.Sprintf("%d", cursor), nil
}

func (beWs *binanceExchangeWs) readTrade(pair *model.TradingPair, symbol string, rawTrade *sdk.EventExecutionReport) (*model.Trade, error) {
	if rawTrade.Symbol != symbol {
		return nil, fmt.Errorf("expected '%s' for 'symbol' field, got: %s", symbol, rawTrade.Symbol)
	}

	pricePrecision := getPrecision(rawTrade.OrderPrice)
	volumePrecision := getPrecision(rawTrade.OrderQuantity)
	// use bigger precision for fee and cost since they are logically derived from amount and price
	feecCostPrecision := pricePrecision
	if volumePrecision > pricePrecision {
		feecCostPrecision = volumePrecision
	}

	orderPrice, err := strconv.ParseFloat(rawTrade.OrderPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting OrderPrice:%v", err)
	}

	orderQuantity, err := strconv.ParseFloat(rawTrade.OrderQuantity, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting OrderQuantity:%v", err)
	}

	comissionAmount, err := strconv.ParseFloat(rawTrade.ComissionAmount, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting ComissionAmount:%v", err)
	}

	trade := model.Trade{
		Order: model.Order{
			Pair:      pair,
			Price:     model.NumberFromFloat(orderPrice, pricePrecision),
			Volume:    model.NumberFromFloat(orderQuantity, volumePrecision),
			OrderType: model.OrderTypeLimit,
			Timestamp: model.MakeTimestamp(rawTrade.TransactionTime),
		},
		TransactionID: model.MakeTransactionID(strconv.FormatInt(rawTrade.OrderID, 10)),
		Cost:          model.NumberFromFloat(comissionAmount, feecCostPrecision),
		// OrderID read by calling function depending on override set for exchange params in "orderId" field of Info object
	}

	if rawTrade.Side == "sell" {
		trade.OrderAction = model.OrderActionSell
	} else if rawTrade.Side == "buy" {
		trade.OrderAction = model.OrderActionBuy
	} else {
		return nil, fmt.Errorf("unrecognized value for 'side' field: %s (rawTrade = %+v)", rawTrade.Side, rawTrade)
	}

	if trade.Cost.AsFloat() < 0 {
		return nil, fmt.Errorf("trade.Cost was < 0 (%f)", trade.Cost.AsFloat())
	}
	if trade.Order.Volume.AsFloat() < 0 {
		return nil, fmt.Errorf("trade.Order.Volume was < 0 (%f)", trade.Order.Volume.AsFloat())
	}

	return &trade, nil
}

//Unsubscribe ... unsubscribe from binance streams
func (beWs *binanceExchangeWs) Unsubscribe(stream string) {

	beWs.streamLock.Lock()

	if stream, isStream := beWs.streams[stream]; isStream {
		stream.Close()
	}

	beWs.streamLock.Unlock()
}
