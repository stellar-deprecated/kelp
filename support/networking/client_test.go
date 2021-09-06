package networking

import (
	"fmt"
	"github.com/stellar/kelp/tests"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	expected := fmt.Sprintf("http client error: status code %d", http.StatusInternalServerError)

	c := NewHttpClient()
	res, err := c.Get(ts.URL)

	assert.Nil(t, res)
	assert.Contains(t, err.Error(), expected)
}

func TestClientGet_BodyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}))
	defer ts.Close()

	expected := fmt.Sprint("http client error: could not read body unexpected EOF")

	c := NewHttpClient()
	res, err := c.Get(ts.URL)

	assert.Nil(t, res)
	assert.Contains(t, err.Error(), expected)
}

func TestClientGet_Ok(t *testing.T) {
	response := tests.RandomString()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer ts.Close()

	c := NewHttpClient()
	res, err := c.Get(ts.URL)

	assert.Nil(t, err)
	assert.Equal(t, response, string(res))
}
