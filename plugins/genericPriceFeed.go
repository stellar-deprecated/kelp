package plugins

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type HttpClient interface {
	Get(url string) ([]byte, error)
}

type JsonParser interface {
	GetRawJsonValue(json []byte, path string) (string, error)
}

type GenericPriceFeed struct {
	url        string
	jsonPath   string
	httpClient HttpClient
	jsonParser JsonParser
}

func NewGenericPriceFeed(url string, jsonPath string, httpClient HttpClient, jsonParser JsonParser) *GenericPriceFeed {
	return &GenericPriceFeed{
		url:        url,
		jsonPath:   jsonPath,
		httpClient: httpClient,
		jsonParser: jsonParser,
	}
}

func (gpf GenericPriceFeed) GetPrice() (float64, error) {
	res, err := gpf.httpClient.Get(gpf.url)
	if err != nil {
		return 0, fmt.Errorf("generic price feed error: %w", err)
	}

	rawValue, err := gpf.jsonParser.GetRawJsonValue(res, gpf.jsonPath)
	if err != nil {
		return 0, fmt.Errorf("generic price feed error: %w", err)
	}

	rawPrice := strings.Trim(rawValue, "\" ")

	price, err := strconv.ParseFloat(rawPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("generic price feed error: %w", err)
	}

	log.Println(fmt.Sprintf("price retrieved for generic %f", price))

	return price, nil
}
