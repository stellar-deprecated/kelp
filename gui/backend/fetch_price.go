package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s\n", e))
		return
	}

	var input fetchPriceInput
	e = json.Unmarshal(bodyBytes, &input)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s\n", e, string(bodyBytes)))
		return
	}

	s.writeJson(w, fetchPriceOutput{
		Price: 0.12,
	})
}
