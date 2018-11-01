package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsRecorder_UpdateMetrics(t *testing.T) {
	m := &metricsRecorder{
		records: map[string]interface{}{},
	}
	m.UpdateMetrics(map[string]interface{}{
		"max_qty":      100,
		"volume":       200000,
		"kelp_version": "1.1",
	})
	assert.Equal(t, 100, m.records["max_qty"])
	assert.Equal(t, 200000, m.records["volume"])
	assert.Equal(t, "1.1", m.records["kelp_version"])
	assert.Equal(t, nil, m.records["nonexistent"])
	m.UpdateMetrics(map[string]interface{}{
		"max_qty": 200,
	})
	assert.Equal(t, 200, m.records["max_qty"])
}

func TestMetricsRecorder_MarshalJSON(t *testing.T) {
	m := &metricsRecorder{
		records: map[string]interface{}{},
	}
	m.UpdateMetrics(map[string]interface{}{
		"statuses": map[string]string{
			"a": "ok",
			"b": "error",
		},
		"trade_ids": []int64{1, 2, 3, 4, 5},
		"version":   "10.0.1",
	})
	json, e := m.MarshalJSON()
	if !assert.Nil(t, e) {
		return
	}
	assert.Equal(t, `{"statuses":{"a":"ok","b":"error"},"trade_ids":[1,2,3,4,5],"version":"10.0.1"}`, string(json))
}
