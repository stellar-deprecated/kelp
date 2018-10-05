package plugins

import (
	"fmt"
	"strings"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
)

// MakePriceFeed makes a PriceFeed
func MakePriceFeed(feedType string, url string) (api.PriceFeed, error) {
	switch feedType {
	case "crypto":
		return newCMCFeed(url), nil
	case "fiat":
		return newFiatFeed(url), nil
	case "fixed":
		return newFixedFeed(url), nil
	case "exchange":
		// [0] = exchangeType, [1] = base, [2] = quote
		urlParts := strings.Split(url, "/")
		exchange, e := MakeExchange(urlParts[0])
		if e != nil {
			return nil, fmt.Errorf("cannot make priceFeed because of an error when making the '%s' exchange: %s", urlParts[0], e)
		}
		tradingPair := model.TradingPair{
			Base:  exchange.GetAssetConverter().MustFromString(urlParts[1]),
			Quote: exchange.GetAssetConverter().MustFromString(urlParts[2]),
		}
		tickerAPI := api.TickerAPI(exchange)
		return newExchangeFeed(&tickerAPI, &tradingPair), nil
	}
	return nil, nil
}

// MakeFeedPair is the factory method that we expose
func MakeFeedPair(dataTypeA, dataFeedAUrl, dataTypeB, dataFeedBUrl string) (*api.FeedPair, error) {
	feedA, e := MakePriceFeed(dataTypeA, dataFeedAUrl)
	if e != nil {
		return nil, fmt.Errorf("cannot make a feed pair because of an error when making priceFeed A: %s", e)
	}

	feedB, e := MakePriceFeed(dataTypeB, dataFeedBUrl)
	if e != nil {
		return nil, fmt.Errorf("cannot make a feed pair because of an error when making priceFeed B: %s", e)
	}

	return &api.FeedPair{
		FeedA: feedA,
		FeedB: feedB,
	}, nil
}
