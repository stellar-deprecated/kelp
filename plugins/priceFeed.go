package plugins

import (
	"strings"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
)

// MakePriceFeed makes a PriceFeed
func MakePriceFeed(feedType string, url string) api.PriceFeed {
	switch feedType {
	case "crypto":
		return newCMCFeed(url)
	case "fiat":
		return newFiatFeed(url)
	case "fixed":
		return newFixedFeed(url)
	case "exchange":
		// [0] = exchangeType, [1] = base, [2] = quote
		urlParts := strings.Split(url, "/")
		exchange := MakeExchange(urlParts[0])
		tradingPair := model.TradingPair{
			Base:  exchange.GetAssetConverter().MustFromString(urlParts[1]),
			Quote: exchange.GetAssetConverter().MustFromString(urlParts[2]),
		}
		tickerAPI := api.TickerAPI(exchange)
		return newExchangeFeed(&tickerAPI, &tradingPair)
	}
	return nil
}

// MakeFeedPair is the factory method that we expose
func MakeFeedPair(dataTypeA, dataFeedAUrl, dataTypeB, dataFeedBUrl string) *api.FeedPair {
	return &api.FeedPair{
		FeedA: MakePriceFeed(dataTypeA, dataFeedAUrl),
		FeedB: MakePriceFeed(dataTypeB, dataFeedBUrl),
	}
}
