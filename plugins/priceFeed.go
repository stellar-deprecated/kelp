package plugins

import (
	"fmt"
	"strings"

	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// privateSdexHack is a temporary hack struct for SDEX price feeds pending refactor
type privateSdexHack struct {
	API     *horizon.Client
	Ieif    *IEIF
	Network build.Network
}

// privateSdexHackVar is a temporary hack variable for SDEX price feeds pending refactor
var privateSdexHackVar *privateSdexHack

// SetPrivateSdexHack sets the privateSdexHack variable which is temporary until the pending SDEX price feed refactor
func SetPrivateSdexHack(api *horizon.Client, ieif *IEIF, network build.Network) error {
	if privateSdexHackVar != nil {
		return fmt.Errorf("privateSdexHack is already set: %+v", privateSdexHackVar)
	}

	privateSdexHackVar = &privateSdexHack{
		API:     api,
		Ieif:    ieif,
		Network: network,
	}
	return nil
}

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
		sdex, e := makeSDEXFeed(url)
		if e != nil {
			return nil, fmt.Errorf("error occurred while making the SDEX price feed: %s", e)
		}
		return sdex, nil
	}
	return nil, fmt.Errorf("unable to make price feed for feedType=%s and url=%s", feedType, url)
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
