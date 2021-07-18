package plugins

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

const (
	STREAM_TICKER_FMT = "%s@ticker"
	TTLTIME           = time.Second * 3 // ttl time in seconds
)

var (
	waitForFirstEvent = true
)

var (
	ErrConversionWsMarketEvent = errConversion{from: "interface", to: "*binance.WsMarketStatEvent"}
)

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
func (m *mapEvents) Set(key string, data interface{}) {

	now := time.Now()

	m.mtx.Lock()
	m.data[key] = mapData{
		data:      data,
		createdAt: now,
	}
	m.mtx.Unlock()
}

//Get ... get value
func (m *mapEvents) Get(key string) (mapData, bool) {
	m.mtx.RLock()
	data, isData := m.data[key]
	m.mtx.RUnlock()

	return data, isData
}

//Del ... delete cached value
func (m *mapEvents) Del(key string) {
	m.mtx.Lock()
	delete(m.data, key)
	m.mtx.Unlock()

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
	SymbolStats *mapEvents
}

func createStateEvents() *events {
	events := &events{
		SymbolStats: makeMapEvents(),
	}

	return events
}

// subscribe for symbol@ticker
func subcribeTicker(symbol string, state *mapEvents) (*stream, error) {

	wsMarketStatHandler := func(ticker *binance.WsMarketStatEvent) {
		state.Set(symbol, ticker)
	}

	errHandler := func(err error) {
		log.Printf("Error WsMarketsStat for symbol %s: %v\n", symbol, err)
	}

	doneC, stopC, err := binance.WsMarketStatServe(symbol, wsMarketStatHandler, errHandler)

	if err != nil {
		return nil, err
	}

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
}

// makeBinanceWs is a factory method to make an binance exchange over ws
func makeBinanceWs() (*binanceExchangeWs, error) {

	binance.WebsocketKeepalive = true

	events := createStateEvents()

	beWs := &binanceExchangeWs{
		events:         events,
		delimiter:      "",
		assetConverter: model.CcxtAssetConverter,
		streamLock:     &sync.Mutex{},
		streams:        make(map[string]*stream),
	}

	return beWs, nil
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
			stream, err := subcribeTicker(symbol, beWs.events.SymbolStats)

			if err != nil {
				return nil, fmt.Errorf("error when subscribing for %s: %s", symbol, err)
			}

			//Store stream
			beWs.streamLock.Lock()
			beWs.streams[fmt.Sprintf(STREAM_TICKER_FMT, symbol)] = stream
			beWs.streamLock.Unlock()

			if waitForFirstEvent {
				//Wait for binance to send events
				time.Sleep(time.Second)
			}

			tickerData, isTicker = beWs.events.SymbolStats.Get(symbol)

			//We couldn't subscribe for this pair
			if !isTicker {
				return nil, fmt.Errorf("error while fetching ticker price for trading pair %s", symbol)
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

//Unsubscribe ... unsubscribe from binance streams
func (beWs *binanceExchangeWs) Unsubscribe(stream string) {

	beWs.streamLock.Lock()

	if stream, isStream := beWs.streams[stream]; isStream {
		stream.Close()
	}

	beWs.streamLock.Unlock()
}
