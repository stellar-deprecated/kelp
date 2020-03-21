package plugins

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// use a global variable for now so it is common across both instances (buy and sell side)
var price2LastPrice map[float64]float64 = map[float64]float64{}

// swingLevelProvider provides levels based on the concept of a swinging price
type swingLevelProvider struct {
	spread                        float64
	offsetSpread                  float64
	amountBase                    float64
	useMaxQuoteInTargetAmountCalc bool // else use maxBase
	maxLevels                     int16
	lastTradePrice                float64
	priceLimit                    float64 // last price for which to place order
	minBase                       float64
	tradeFetcher                  api.TradeFetcher
	tradingPair                   *model.TradingPair
	lastTradeCursor               interface{}
	isFirstTradeHistoryRun        bool
	incrementTimestampCursor      bool
	orderConstraints              *model.OrderConstraints
}

// ensure it implements LevelProvider
var _ api.LevelProvider = &swingLevelProvider{}

// makeSwingLevelProvider is the factory method
func makeSwingLevelProvider(
	spread float64,
	offsetSpread float64,
	useMaxQuoteInTargetAmountCalc bool,
	amountBase float64,
	maxLevels int16,
	lastTradePrice float64,
	priceLimit float64,
	minBase float64,
	tradeFetcher api.TradeFetcher,
	tradingPair *model.TradingPair,
	lastTradeCursor interface{},
	incrementTimestampCursor bool,
	orderConstraints *model.OrderConstraints,
) *swingLevelProvider {
	return &swingLevelProvider{
		spread:                        spread,
		offsetSpread:                  offsetSpread,
		useMaxQuoteInTargetAmountCalc: useMaxQuoteInTargetAmountCalc,
		amountBase:                    amountBase,
		maxLevels:                     maxLevels,
		lastTradePrice:                lastTradePrice,
		priceLimit:                    priceLimit,
		minBase:                       minBase,
		tradeFetcher:                  tradeFetcher,
		tradingPair:                   tradingPair,
		lastTradeCursor:               lastTradeCursor,
		isFirstTradeHistoryRun:        true,
		incrementTimestampCursor:      incrementTimestampCursor,
		orderConstraints:              orderConstraints,
	}
}

func printPrice2LastPriceMap() {
	keys := []float64{}
	for k, _ := range price2LastPrice {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	log.Printf("price2LastPrice map (%d elements):\n", len(price2LastPrice))
	for _, k := range keys {
		log.Printf("    %.8f -> %.8f\n", k, price2LastPrice[k])
	}
}

func getLastPriceFromMap(mapKey *model.Number, lastTradeIsBuy bool) float64 {
	tradePrice := mapKey.AsFloat()
	if lp, ok := price2LastPrice[tradePrice]; ok {
		log.Printf("getLastPriceFromMap, found in map for tradePrice = %.8f: last price (%.8f)\n", tradePrice, lp)
		return lp
	}

	closestOfferPrice := -1.0
	lp := -1.0
	for offerPrice, offerLastPrice := range price2LastPrice {
		firstIter := closestOfferPrice == -1
		buyTrigger := lastTradeIsBuy && offerPrice > tradePrice && offerPrice < closestOfferPrice
		sellTrigger := !lastTradeIsBuy && offerPrice < tradePrice && offerPrice > closestOfferPrice
		if firstIter || buyTrigger || sellTrigger {
			closestOfferPrice = offerPrice
			lp = offerLastPrice
		}
	}
	log.Printf("getLastPriceFromMap, calculated for tradePrice = %.8f: closest offerPrice (%.8f) and last price (%.8f) when it was not in map\n", tradePrice, closestOfferPrice, lp)
	return lp
}

// GetFillHandlers impl
func (s *swingLevelProvider) GetFillHandlers() ([]api.FillHandler, error) {
	return nil, nil
}

// GetLevels impl.
func (p *swingLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]api.Level, error) {
	if maxAssetBase <= p.minBase {
		return []api.Level{}, nil
	}

	lastPrice, lastCursor, lastIsBuy, e := p.fetchLatestTradePrice()
	if e != nil {
		return nil, fmt.Errorf("error in fetchLatestTradePrice: %s", e)
	}

	// update it only if there's no error
	if p.isFirstTradeHistoryRun {
		p.isFirstTradeHistoryRun = false
		p.lastTradeCursor = lastCursor
		log.Printf("isFirstTradeHistoryRun so updated lastTradeCursor=%v, leaving unchanged lastTradePrice=%.10f", p.lastTradeCursor, p.lastTradePrice)
	} else if lastCursor == p.lastTradeCursor {
		log.Printf("lastCursor == p.lastTradeCursor leaving lastTradeCursor=%v and lastTradePrice=%.10f", p.lastTradeCursor, p.lastTradePrice)
	} else {
		p.lastTradeCursor = lastCursor
		mapKey := model.NumberFromFloat(lastPrice, p.orderConstraints.PricePrecision)
		printPrice2LastPriceMap()
		p.lastTradePrice = getLastPriceFromMap(mapKey, lastIsBuy)
		log.Printf("updated lastTradeCursor=%v and lastTradePrice=%.10f (converted=%.10f)", p.lastTradeCursor, lastPrice, p.lastTradePrice)
	}

	levels := []api.Level{}
	newPrice := p.lastTradePrice
	if p.useMaxQuoteInTargetAmountCalc {
		// invert lastTradePrice here -- it's always kept in the actual quote price at all other times
		newPrice = 1 / newPrice
	}
	baseExposed := 0.0
	for i := 0; i < int(p.maxLevels); i++ {
		newPrice = newPrice * (1 + p.spread/2)
		priceToUse := newPrice * (1 + p.offsetSpread/2)

		// check what the balance would be if we were to place this level, ensuring it will still be within the limits
		expectedBaseUsage := p.amountBase
		if p.useMaxQuoteInTargetAmountCalc {
			expectedBaseUsage = expectedBaseUsage / priceToUse
		}
		expectedEndingBase := maxAssetBase - baseExposed - expectedBaseUsage
		if expectedEndingBase <= p.minBase {
			log.Printf("early exiting level creation loop (sideIsBuy=%v), expectedEndingBase=%.10f, minBase=%.10f\n", p.useMaxQuoteInTargetAmountCalc, expectedEndingBase, p.minBase)
			break
		}

		if p.useMaxQuoteInTargetAmountCalc && 1/priceToUse < p.priceLimit {
			log.Printf("early exiting level creation loop (buy side) because we crossed minPrice, priceLimit=%.10f, current price=%.10f\n", p.priceLimit, 1/priceToUse)
			break
		}

		if !p.useMaxQuoteInTargetAmountCalc && priceToUse > p.priceLimit {
			log.Printf("early exiting level creation loop (sell side) because we crossed maxPrice, priceLimit=%.10f, current price=%.10f\n", p.priceLimit, priceToUse)
			break
		}

		levels = append(levels, api.Level{
			Price:  *model.NumberFromFloat(priceToUse, p.orderConstraints.PricePrecision),
			Amount: *model.NumberFromFloat(p.amountBase, p.orderConstraints.VolumePrecision),
		})

		// update last price map here
		mapKey := model.NumberFromFloat(priceToUse, p.orderConstraints.PricePrecision)
		mapValue := newPrice
		if p.useMaxQuoteInTargetAmountCalc {
			mapKey = model.NumberFromFloat(1/priceToUse, p.orderConstraints.PricePrecision)
			mapValue = 1 / newPrice
		}
		price2LastPrice[mapKey.AsFloat()] = mapValue

		baseExposed += expectedBaseUsage
	}

	log.Printf("levels created (sideIsBuy=%v): %v\n", p.useMaxQuoteInTargetAmountCalc, levels)
	printPrice2LastPriceMap()

	return levels, nil
}

func (p *swingLevelProvider) fetchLatestTradePrice() (float64, interface{}, bool, error) {
	lastPrice := p.lastTradePrice
	lastCursor := p.lastTradeCursor
	lastIsBuy := false
	for {
		tradeHistoryResult, e := p.tradeFetcher.GetTradeHistory(*p.tradingPair, lastCursor, nil)
		if e != nil {
			return 0, "", false, fmt.Errorf("error in tradeFetcher.GetTradeHistory: %s", e)
		}

		// TODO need to check for volume here too at some point (if full lot is not taken then we don't want to update last price)

		if len(tradeHistoryResult.Trades) == 0 {
			return lastPrice, tradeHistoryResult.Cursor, lastIsBuy, nil
		}

		log.Printf("listing %d trades since last cycle", len(tradeHistoryResult.Trades))
		for _, t := range tradeHistoryResult.Trades {
			log.Printf("    Trade: %v\n", t)
		}

		lastTrade := tradeHistoryResult.Trades[len(tradeHistoryResult.Trades)-1]
		if p.incrementTimestampCursor {
			i64Cursor, e := strconv.Atoi(lastTrade.Order.Timestamp.String())
			if e != nil {
				return 0, "", false, fmt.Errorf("unable to convert order timestamp to integer for binance cursor: %s", e)
			}
			// increment last timestamp cursor for binance because it's inclusive
			lastCursor = strconv.FormatInt(int64(i64Cursor)+1, 10)
		} else {
			lastCursor = lastTrade.TransactionID.String()
		}
		lastIsBuy = lastTrade.Order.OrderAction == model.OrderActionBuy
		price := lastTrade.Order.Price.AsFloat()
		lastPrice = price
	}
}
