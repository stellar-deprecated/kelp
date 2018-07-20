package priceFeed

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/exchange"
)

// priceFeed allows you to fetch the price of a feed
type priceFeed interface {
	getPrice() (float64, error)
}

func priceFeedFactory(feedType string, url string) priceFeed {
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
		xc := exchange.ExchangeFactory(urlParts[0])
		tradingPair := model.TradingPair{
			Base:  xc.GetAssetConverter().MustFromString(urlParts[1]),
			Quote: xc.GetAssetConverter().MustFromString(urlParts[2]),
		}
		return newExchangeFeed(&xc, &tradingPair)
	}
	return nil
}

func getJSON(client http.Client, url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
