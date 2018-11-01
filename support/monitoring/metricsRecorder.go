package monitoring

import "encoding/json"

// MetricsRecorder uses a map to store metrics and implements the api.Metrics interface.
type metricsRecorder struct {
	records map[string]interface{}
}

var _ Metrics = &metricsRecorder{}

// MakeMetricsRecorder makes a metrics recorder with the records map as the underlying map. If records
// is nil, then an empty map will be initialized for you.
func MakeMetricsRecorder(records map[string]interface{}) (Metrics, error) {
	if records == nil {
		return &metricsRecorder{
			records: map[string]interface{}{},
		}, nil
	}
	return &metricsRecorder{
		records: records,
	}, nil
}

// UpdateMetrics updates (or adds if non-existent) metrics in the records for all key-value
// pairs in the provided map of metrics.
func (m *metricsRecorder) UpdateMetrics(metrics map[string]interface{}) {
	for k, v := range metrics {
		m.records[k] = v
	}
}

// MarshalJSON gives the JSON representation of the records.
func (m *metricsRecorder) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.records)
}
