package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stellar/kelp/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetPrice_ShouldReturnZero_ClientError(t *testing.T) {
	oxrFeed := newFiatFeedOxr(tests.RandomString())
	price, err := oxrFeed.GetPrice()
	assert.Equal(t, price, float64(0))
	assert.Contains(t, err.Error(), "oxr: error ")
}

func Test_GetPrice_ShouldReturnZero_OxrError(t *testing.T) {
	keys := make([]int, 0, len(oxrErrorCodeMsg))
	for k, _ := range oxrErrorCodeMsg {
		keys = append(keys, k)
	}

	expected := oxrError{
		Err:         true,
		Status:      keys[tests.RandomIntWithMax(len(keys))],
		Message:     tests.RandomString(),
		Description: tests.RandomString(),
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expected.Status)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(expected)
		require.NoError(t, err)
	}))
	defer ts.Close()

	oxrFeed := newFiatFeedOxr(ts.URL)
	price, err := oxrFeed.GetPrice()

	assert.Equal(t, float64(0), price)
	assert.Equal(t, expected, err)
}

func Test_GetPrice_ShouldReturnInvalidRateLength(t *testing.T) {
	response := createOxrResponse("")
	response.Rates = nil

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer ts.Close()

	oxrFeed := newFiatFeedOxr(ts.URL)
	expected, err := oxrFeed.GetPrice()

	assert.Equal(t, float64(0), expected)
	assert.Equal(t, fmt.Errorf("oxr: error rates must contain single value found len %d", len(response.Rates)), err)
}

func Test_GetPrice_ShouldReturnRates(t *testing.T) {
	symbol := tests.RandomString()
	response := createOxrResponse(symbol)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer ts.Close()

	oxrFeed := newFiatFeedOxr(ts.URL)
	price, err := oxrFeed.GetPrice()

	assert.Equal(t, response.Rates[symbol], price)
	assert.NoError(t, err)
}

func createOxrResponse(symbol string) oxrRates {
	return oxrRates{
		Disclaimer: tests.RandomString(),
		License:    tests.RandomString(),
		Timestamp:  int64(tests.RandomInt()),
		Base:       tests.RandomString(),
		Rates: map[string]float64{
			symbol: float64(tests.RandomInt()),
		},
	}
}
