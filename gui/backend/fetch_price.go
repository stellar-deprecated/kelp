package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/utils"
)

type fetchPriceInput struct {
	Type    string `json:"type"`
	FeedURL string `json:"feed_url"`
}

type fetchPriceOutput struct {
	Price float64 `json:"price"`
}

func (s *APIServer) fetchPrice(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s", e))
		return
	}
	log.Printf("requestJson: %s\n", string(bodyBytes))

	var input fetchPriceInput
	e = json.Unmarshal(bodyBytes, &input)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}

	pf, e := plugins.MakePriceFeed(input.Type, input.FeedURL)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to make price feed: %s", e))
		return
	}

	price, e := pf.GetPrice()
	if e != nil {
		if fiatAPIError, ok := errors.Cause(e).(plugins.ErrFiatAPI); ok && (fiatAPIError.Code == plugins.FiatErrorCodeInvalidAPIKey || fiatAPIError.Code == plugins.FiatErrorCodeExhaustedAPIKey || fiatAPIError.Code == plugins.FiatErrorCodeAccountInactive) {
			log.Printf("fiat API error when fetching price: %s\n", fiatAPIError)
			s.writeJson(w, fetchPriceOutput{Price: -1.0})
			return
		}

		s.writeErrorJson(w, fmt.Sprintf("unable to fetch price: %s", e))
		return
	}

	priceCappedPrecision := model.NumberFromFloat(price, utils.SdexPrecision).AsFloat()
	s.writeJson(w, fetchPriceOutput{
		Price: priceCappedPrecision,
	})
}
