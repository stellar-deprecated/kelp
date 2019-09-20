package plugins

import (
	"fmt"
	"net/http"
	"time"

	"github.com/stellar/go/support/errors"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/utils"
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

const FiatErrorCodeInvalidAPIKey = 101
const FiatErrorCodeAccountInactive = 102
const FiatErrorCodeExhaustedAPIKey = 104

// ErrFiatAPI is a custom error currently used to identify when we have an invalid APIKey
type ErrFiatAPI struct {
	Code int
	Type string
	Info string
}

var _ error = ErrFiatAPI{}

func (e ErrFiatAPI) Error() string {
	return fmt.Sprintf("ErrFiatAPI[code=%d, type=%s, info='%s']", e.Code, e.Type, e.Info)
}

type fiatAPIReturn struct {
	Success bool
	Quotes  map[string]float64
	Error   ErrFiatAPI
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
	e := utils.GetJSON(f.client, f.url, &ret)
	if e != nil {
		return 0, fmt.Errorf("unable to get price from fiat feed: %s", e)
	}

	if !ret.Success {
		return -1, errors.Wrap(ret.Error, "call to get price from fiat feed failed")
	}

	if len(ret.Quotes) != 1 {
		return 0, fmt.Errorf("incorrect number of quotes returned (%d), was expecting only 1", len(ret.Quotes))
	}

	for _, price := range ret.Quotes {
		return (1.0 / price), nil
	}
	return -1, fmt.Errorf("unexpected error, should not have reached here")
}
