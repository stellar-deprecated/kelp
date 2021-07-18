package plugins

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

const (
	STREAMTICKER = "@ticker"
	TTLTIME      = time.Second * 3 // ttl time in seconds
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

type mapData struct {
	data      interface{}
	createdAt time.Time
}

func isStale(data mapData, ttl time.Duration) bool {

	return time.Now().Sub(data.createdAt).Seconds() > ttl.Seconds()
}

type mapEvents struct {
	data map[string]mapData
	mtx  *sync.Mutex
}

func (m *mapEvents) Set(key string, data interface{}) {

	now := time.Now()

	m.mtx.Lock()
	m.data[key] = mapData{
		data:      data,
		createdAt: now,
	}
	m.mtx.Unlock()
}

func (m *mapEvents) Get(key string) (mapData, bool) {
	m.mtx.Lock()
	data, isData := m.data[key]
	m.mtx.Unlock()

	return data, isData
}

func (m *mapEvents) Del(key string) {
	m.mtx.Lock()
	delete(m.data, key)
	m.mtx.Unlock()

}

func makeMapEvents() *mapEvents {
	return &mapEvents{
		data: make(map[string]mapData),
		mtx:  &sync.Mutex{},
	}
}

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
	events    *events
	delimiter string

	streams map[string]*stream

	assetConverter model.AssetConverterInterface

	mtx *sync.Mutex
}

// makeBinanceWs is a factory method to make an binance exchange over ws
func makeBinanceWs() (*binanceExchangeWs, error) {

	binance.WebsocketKeepalive = true

	events := createStateEvents()

	beWs := &binanceExchangeWs{
		events:         events,
		delimiter:      "",
		assetConverter: model.CcxtAssetConverter,
		mtx:            &sync.Mutex{},
		streams:        make(map[string]*stream),
	}

	return beWs, nil
}

func getPrecision(floatStr string) int8 {

	strs := strings.Split(floatStr, ".")

	if len(strs) != 2 {
		log.Printf("error get precision for float %s\n", floatStr)
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
			beWs.mtx.Lock()
			beWs.streams[symbol+STREAMTICKER] = stream
			beWs.mtx.Unlock()

			//Wait for binance to send events
			time.Sleep(time.Second)

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

	beWs.mtx.Lock()

	if stream, isStream := beWs.streams[stream]; isStream {
		stream.Close()
	}

	beWs.mtx.Unlock()
}
