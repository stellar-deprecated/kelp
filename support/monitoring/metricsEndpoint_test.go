package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stellar/kelp/support/networking"
	"github.com/stretchr/testify/assert"
)

func TestMetricsEndpoint_NoAuthEndpoint(t *testing.T) {
	testMetrics, e := MakeMetricsRecorder(map[string]interface{}{"this is a test message": true})
	if !assert.Nil(t, e) {
		return
	}
	testEndpoint, e := MakeMetricsEndpoint("/test", testMetrics, networking.NoAuth)
	if !assert.Nil(t, e) {
		return
	}

	req, e := http.NewRequest("GET", "/test", nil)
	if !assert.Nil(t, e) {
		return
	}
	w := httptest.NewRecorder()
	testEndpoint.GetHandlerFunc().ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"this is a test message\":true}", w.Body.String())

	// Mutate the metrics and test if the server response changes
	testMetrics.UpdateMetrics(map[string]interface{}{"this is a test message": false})

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/test", nil)
	testEndpoint.GetHandlerFunc().ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"this is a test message\":false}", w.Body.String())
}
