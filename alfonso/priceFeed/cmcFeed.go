package priceFeed

import (
	"net/http"
	"strconv"
	"time"
	//"github.com/stellar/go/support/log"
)

/*
example JSON returned by coinmarketcap
[
    {
        "id": "stellar",
        "name": "Stellar Lumens",
        "symbol": "XLM",
        "rank": "27",
        "price_usd": "0.0220156",
        "price_btc": "0.00000527",
        "24h_volume_usd": "27604200.0",
        "market_cap_usd": "245811502.0",
        "available_supply": "11165332853.0",
        "total_supply": "103195955318",
        "percent_change_1h": "0.67",
        "percent_change_24h": "10.76",
        "percent_change_7d": "23.69",
        "last_updated": "1503513850"
    }
]
*/

// for getting data out of coinmarketcap
type cmcAPIReturn struct {
	Price string `json:"price_usd"`
}

type cmcFeed struct {
	url    string
	client http.Client
}

// ensure that it implements priceFeed
var _ priceFeed = &cmcFeed{}

func newCMCFeed(url string) *cmcFeed {
	m := new(cmcFeed)
	m.url = url
	m.client = http.Client{Timeout: 10 * time.Second}
	return m
}

func (c *cmcFeed) getPrice() (float64, error) {
	var retA []cmcAPIReturn
	err := getJSON(c.client, c.url, &retA)
	if err != nil {
		return 0, err
	}

	pA, err := strconv.ParseFloat(retA[0].Price, 64)
	if err != nil {
		return 0, err
	}

	return pA, nil
}
