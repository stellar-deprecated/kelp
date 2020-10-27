package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeEventProps(t *testing.T) {
	commonProps := map[string]interface{}{
		"1": 1,
		"2": "two",
	}
	eventProps := map[string]interface{}{
		"3": 3,
		"4": "four",
	}
	wantMap := map[string]interface{}{
		"1": 1,
		"2": "two",
		"3": 3,
		"4": "four",
	}

	gotMap, e := mergeEventProps(commonProps, eventProps)
	assert.Equal(t, wantMap, gotMap)
	assert.NoError(t, e)
}

func TestToMapStringInterface(t *testing.T) {
	success := map[string]interface{}{
		"test": true,
	}
	m, e := toMapStringInterface(success)
	assert.Equal(t, success, m)
	assert.NoError(t, e)

	failure := 0.0
	m, e = toMapStringInterface(failure)
	assert.EqualError(t, e, "could not create map[string]interface{}")
	assert.Nil(t, m)
}
