package networking

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpClient struct {
	client *http.Client
}

func NewHttpClient() *httpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (hc httpClient) Get(url string) ([]byte, error) {
	res, err := hc.client.Get(url)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http client error: status code %d %w", res.StatusCode, err)
	}
	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("http client error: could not read body %w", err)
	}

	return bytes, nil
}
