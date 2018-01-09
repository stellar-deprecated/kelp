package priceFeed

import (
	"encoding/json"
	"net/http"
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
