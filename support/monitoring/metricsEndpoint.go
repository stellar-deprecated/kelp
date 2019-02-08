package monitoring

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/kelp/support/networking"
)

// metricsEndpoint represents a monitoring API endpoint that always responds with a JSON
// encoding of the provided metrics. The auth level for the endpoint can be NoAuth (public access)
// or GoogleAuth which uses a Google account for authorization.
type metricsEndpoint struct {
	path      string
	metrics   Metrics
	authLevel networking.AuthLevel
}

// MakeMetricsEndpoint creates an Endpoint for the monitoring server with the desired auth level.
// The endpoint's response is always a JSON dump of the provided metrics.
func MakeMetricsEndpoint(path string, metrics Metrics, authLevel networking.AuthLevel) (networking.Endpoint, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("endpoint path must begin with /")
	}
	s := &metricsEndpoint{
		path:      path,
		metrics:   metrics,
		authLevel: authLevel,
	}
	return s, nil
}

func (m *metricsEndpoint) GetAuthLevel() networking.AuthLevel {
	return m.authLevel
}

func (m *metricsEndpoint) GetPath() string {
	return m.path
}

// GetHandlerFunc returns a HandlerFunc that writes the JSON representation of the metrics
// that's passed into the endpoint.
func (m *metricsEndpoint) GetHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json, e := m.metrics.MarshalJSON()
		if e != nil {
			log.Printf("error marshalling metrics json: %s\n", e)
			http.Error(w, e.Error(), 500)
			return
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, e = w.Write(json)
		if e != nil {
			log.Printf("error writing to the response writer: %s\n", e)
		}
	}
}
