package priceFeed

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lightyeario/kelp/support"

	"github.com/lightyeario/kelp/support/exchange/assets"
)

// priceFeed allows you to fetch the price of a feed
type priceFeed interface {
	getPrice() (float64, error)
}

func priceFeedFactory(feedType string, url string) priceFeed {
	switch feedType {
	case "crypto":
		return newCMCFeed(url)
	case "fiat":
		return newFiatFeed(url)
	case "fixed":
		return newFixedFeed(url)
	case "exchange":
		// TODO 2 this should not be needed here because this gives you two prices (1 ratio),
		// whereas priceFeed typically only gives you 1 price. Maybe this should not be a priceFeed
		// or it should be a different impl. of priceFeed or something? something is different here that I need to think about.

		// [0] = exchangeType, [1] = base, [2] = quote, [1] = bidPrice/askPrice
		urlParts := strings.Split(url, "/")
		exchange := kelp.ExchangeFactory(urlParts[0])
		tradingPair := assets.TradingPair{
			Base:  exchange.GetAssetConverter().MustFromString(urlParts[1]),
			Quote: exchange.GetAssetConverter().MustFromString(urlParts[2]),
		}
		useBidPrice := urlParts[3] == "bidPrice"
		return newExchangeFeed(&exchange, &tradingPair, useBidPrice)
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
