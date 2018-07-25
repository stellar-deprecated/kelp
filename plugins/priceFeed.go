package plugins

import (
	"strings"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
)

func makePriceFeed(feedType string, url string) api.PriceFeed {
	switch feedType {
	case "crypto":
		return NewCMCFeed(url)
	case "fiat":
		return newFiatFeed(url)
	case "fixed":
		return newFixedFeed(url)
	case "exchange":
		// [0] = exchangeType, [1] = base, [2] = quote
		urlParts := strings.Split(url, "/")
		xc := MakeExchange(urlParts[0])
		tradingPair := model.TradingPair{
			Base:  xc.GetAssetConverter().MustFromString(urlParts[1]),
			Quote: xc.GetAssetConverter().MustFromString(urlParts[2]),
		}
		return newExchangeFeed(&xc, &tradingPair)
	}
	return nil
}

// MakeFeedPair is the factory method that we expose
func MakeFeedPair(dataTypeA, dataFeedAUrl, dataTypeB, dataFeedBUrl string) *api.FeedPair {
	return &api.FeedPair{
		FeedA: makePriceFeed(dataTypeA, dataFeedAUrl),
		FeedB: makePriceFeed(dataTypeB, dataFeedBUrl),
	}
}
