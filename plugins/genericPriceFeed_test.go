package plugins_test

import (
	"fmt"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPrice_HttpClient_Error(t *testing.T) {
	httpClient := mockHttpClient{
		bytes: []byte{},
		err:   fmt.Errorf(tests.RandomString()),
	}

	genericPriceFeed := plugins.NewGenericPriceFeed(tests.RandomString(), tests.RandomString(),
		httpClient, mockJsonParser{})

	price, err := genericPriceFeed.GetPrice()

	expected := fmt.Sprintf("generic price feed error: %s", httpClient.err.Error())

	assert.EqualError(t, err, expected)
	assert.Equal(t, float64(0), price)
}

func TestGetPrice_JsonParser_Error(t *testing.T) {
	httpClient := mockHttpClient{
		bytes: []byte{},
		err:   nil,
	}

	jsonParser := mockJsonParser{
		rawValue: "",
		err:      fmt.Errorf(tests.RandomString()),
	}

	genericPriceFeed := plugins.NewGenericPriceFeed(tests.RandomString(), tests.RandomString(),
		httpClient, jsonParser)

	price, err := genericPriceFeed.GetPrice()

	expected := fmt.Sprintf("generic price feed error: %s", jsonParser.err.Error())

	assert.EqualError(t, err, expected)
	assert.Equal(t, float64(0), price)
}

func TestGetPrice_ParseFloat_Error(t *testing.T) {
	httpClient := mockHttpClient{
		bytes: []byte{},
		err:   nil,
	}

	jsonParser := mockJsonParser{
		rawValue: tests.RandomString(),
		err:      nil,
	}

	genericPriceFeed := plugins.NewGenericPriceFeed(tests.RandomString(), tests.RandomString(),
		httpClient, jsonParser)

	price, err := genericPriceFeed.GetPrice()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), jsonParser.rawValue)
	assert.Equal(t, float64(0), price)
}

func TestGetPrice_Float(t *testing.T) {
	httpClient := mockHttpClient{
		bytes: []byte{},
		err:   nil,
	}

	expected := tests.RandomFloat64()
	jsonParser := mockJsonParser{
		rawValue: fmt.Sprintf("%f", expected),
		err:      nil,
	}

	genericPriceFeed := plugins.NewGenericPriceFeed(tests.RandomString(), tests.RandomString(),
		httpClient, jsonParser)

	price, err := genericPriceFeed.GetPrice()

	assert.Equal(t, expected, price)
	assert.NoError(t, err)
}

func TestGetPrice_Trim_DoubleQuotes(t *testing.T) {
	httpClient := mockHttpClient{
		bytes: []byte{},
		err:   nil,
	}

	expected := tests.RandomFloat64()
	jsonParser := mockJsonParser{
		rawValue: fmt.Sprintf("\"%f\"", expected),
		err:      nil,
	}

	genericPriceFeed := plugins.NewGenericPriceFeed(tests.RandomString(), tests.RandomString(),
		httpClient, jsonParser)

	price, err := genericPriceFeed.GetPrice()

	assert.Equal(t, expected, price)
	assert.NoError(t, err)
}

func TestGetPrice_Trim_WhiteSpace(t *testing.T) {
	httpClient := mockHttpClient{
		bytes: []byte{},
		err:   nil,
	}

	expected := tests.RandomFloat64()
	jsonParser := mockJsonParser{
		rawValue: fmt.Sprintf(" %f ", expected),
		err:      nil,
	}

	genericPriceFeed := plugins.NewGenericPriceFeed(tests.RandomString(), tests.RandomString(),
		httpClient, jsonParser)

	price, err := genericPriceFeed.GetPrice()

	assert.Equal(t, expected, price)
	assert.NoError(t, err)
}

type mockHttpClient struct {
	bytes []byte
	err   error
}

func (m mockHttpClient) Get(url string) ([]byte, error) {
	return m.bytes, m.err
}

type mockJsonParser struct {
	rawValue string
	err      error
}

func (m mockJsonParser) GetRawJsonValue(json []byte, path string) (string, error) {
	return m.rawValue, m.err
}
