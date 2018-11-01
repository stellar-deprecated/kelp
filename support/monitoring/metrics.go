package monitoring

// Metrics is an interface that allows a client to pass in key value pairs (keys must be strings)
// and it can dump the metrics as JSON.
type Metrics interface {
	UpdateMetrics(metrics map[string]interface{})
	MarshalJSON() ([]byte, error)
}
