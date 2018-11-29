package plugins

import (
	"net/http"
	"time"

	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/support/utils"
)

/*
{
	"success":true,
	"terms":"https:\/\/currencylayer.com\/terms",
	"privacy":"https:\/\/currencylayer.com\/privacy",
	"timestamp":1504027454,
	"source":"USD",
	"quotes":{"USDPHP":51.080002}
}
*/

type fiatAPIReturn struct {
	Quotes map[string]float64
}

type fiatFeed struct {
	url    string
	client http.Client
}

// ensure that it implements PriceFeed
var _ api.PriceFeed = &fiatFeed{}

func newFiatFeed(url string) *fiatFeed {
	m := new(fiatFeed)
	m.url = url
	m.client = http.Client{Timeout: 10 * time.Second}

	return m
}

// GetPrice impl
func (f *fiatFeed) GetPrice() (float64, error) {
	var ret fiatAPIReturn
	err := utils.GetJSON(f.client, f.url, &ret)
	if err != nil {
		return 0, err
	}
	var pA float64
	for _, value := range ret.Quotes {
		pA = value
	}

	return (1.0 / pA), nil
}
