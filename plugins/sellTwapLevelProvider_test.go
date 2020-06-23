package plugins

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateFloorCeil(t *testing.T) {
	timeLayout := time.RFC3339 // 2006-01-02T15:04:05Z
	testCases := []struct {
		input     string
		wantFloor string
		wantCeil  string
	}{
		{
			input:     "2006-01-02T15:04:05Z",
			wantFloor: "2006-01-02T00:00:00Z",
			wantCeil:  "2006-01-02T23:59:59Z",
		}, {
			input:     "2006-01-02T15:04:05+07:00",
			wantFloor: "2006-01-02T00:00:00+07:00",
			wantCeil:  "2006-01-02T23:59:59+07:00",
		}, {
			input:     "2020-02-28T10:34:59-04:30",
			wantFloor: "2020-02-28T00:00:00-04:30",
			wantCeil:  "2020-02-28T23:59:59-04:30",
		}, {
			input:     "2020-02-29T10:34:59-04:30",
			wantFloor: "2020-02-29T00:00:00-04:30",
			wantCeil:  "2020-02-29T23:59:59-04:30",
		},
	}

	for _, k := range testCases {
		t.Run(k.input, func(t *testing.T) {
			input, e := time.Parse(timeLayout, k.input)
			if !assert.NoError(t, e) {
				return
			}

			floor := floorDate(input).Format(timeLayout)
			assert.Equal(t, k.wantFloor, floor, "floor")

			ceil := ceilDate(input).Format(timeLayout)
			assert.Equal(t, k.wantCeil, ceil, "ceil")
		})
	}
}
