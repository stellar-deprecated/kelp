package plugins

import (
	"fmt"
	"strings"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/model"
)

// MakePriceFeed makes a PriceFeed
func MakePriceFeed(sdex *SDEX, feedType, url string) (api.PriceFeed, error) {
	switch feedType {
	case "crypto":
		return NewCMCFeed(url), nil
	case "fiat":
		return newFiatFeed(url), nil
	case "fixed":
		return newFixedFeed(url), nil
	case "exchange":
		// [0] = exchangeType, [1] = base, [2] = quote
		urlParts := strings.Split(url, "/")
		exchange, e := MakeExchange(urlParts[0], true)
		if e != nil {
			return nil, fmt.Errorf("cannot make priceFeed because of an error when making the '%s' exchange: %s", urlParts[0], e)
		}
		baseAsset, e := exchange.GetAssetConverter().FromString(urlParts[1])
		if e != nil {
			return nil, fmt.Errorf("cannot make priceFeed because of an error when converting the base asset: %s", e)
		}
		quoteAsset, e := exchange.GetAssetConverter().FromString(urlParts[2])
		if e != nil {
			return nil, fmt.Errorf("cannot make priceFeed because of an error when converting the quote asset: %s", e)
		}
		tradingPair := model.TradingPair{
			Base:  baseAsset,
			Quote: quoteAsset,
		}
		tickerAPI := api.TickerAPI(exchange)
		return newExchangeFeed(url, &tickerAPI, &tradingPair), nil
	case "sdex":
		SDEXfeed, e := newSDEXFeed(sdex, url)
		if e != nil {
			return nil, fmt.Errorf("unable to create SDEX priceFeed ")
		}
		return SDEXfeed, nil
	}
	return nil, nil
}

// MakeFeedPair is the factory method that we expose
func MakeFeedPair(sdex *SDEX, dataTypeA, dataFeedAUrl, dataTypeB, dataFeedBUrl string) (*api.FeedPair, error) {
	feedA, e := MakePriceFeed(sdex, dataTypeA, dataFeedAUrl)
	if e != nil {
		return nil, fmt.Errorf("cannot make a feed pair because of an error when making priceFeed A: %s", e)
	}

	feedB, e := MakePriceFeed(sdex, dataTypeB, dataFeedBUrl)
	if e != nil {
		return nil, fmt.Errorf("cannot make a feed pair because of an error when making priceFeed B: %s", e)
	}

	return &api.FeedPair{
		FeedA: feedA,
		FeedB: feedB,
	}, nil
}
