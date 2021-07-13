package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type fiatFeedOxr struct {
	url    string
	client http.Client
}

type oxrRates struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

type oxrError struct {
	Err         bool   `json:"error"`
	Status      int    `json:"status"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

func (e oxrError) Error() string {
	return fmt.Sprintf("%v: %v", e.Message, e.Description)
}

var oxrErrorCodeMsg = map[int]string{
	404: "not_found",
	401: "invalid_or_missing_app_id",
	429: "not_allowed",
	403: "access_restricted",
	400: "invalid_base",
}

func newFiatFeedOxr(url string) *fiatFeedOxr {
	return &fiatFeedOxr{
		url:    url,
		client: http.Client{Timeout: 10 * time.Second},
	}
}

func (f *fiatFeedOxr) GetPrice() (float64, error) {
	res, err := f.client.Get(f.url)
	if err != nil {
		return 0, fmt.Errorf("oxr: error %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		var e oxrError
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return 0, err
		}
		return 0, e
	}

	var rates oxrRates
	if err := json.NewDecoder(res.Body).Decode(&rates); err != nil {
		return 0, err
	}

	if len(rates.Rates) != 1 {
		return 0, fmt.Errorf("oxr: error rates must contain single value found len %d", len(rates.Rates))
	}

	for _, v := range rates.Rates {
		return v, nil
	}

	return 0, fmt.Errorf("oxr: error retrieving price")
}
