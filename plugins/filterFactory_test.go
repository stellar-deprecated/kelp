package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMarketIdsArray(t *testing.T) {
	testCases := []struct {
		marketIdsArrayString string
		want                 []string
	}{
		{
			marketIdsArrayString: "[abcde,gfhij]",
			want:                 []string{"abcde", "gfhij"},
		}, {
			marketIdsArrayString: "[abcde, afhij]",
			want:                 []string{"abcde", "afhij"},
		}, {
			marketIdsArrayString: "[abcde]",
			want:                 []string{"abcde"},
		},
	}

	for _, kase := range testCases {
		t.Run(kase.marketIdsArrayString, func(t *testing.T) {
			output, e := parseMarketIdsArray(kase.marketIdsArrayString)
			if !assert.NoError(t, e) {
				return
			}

			assert.Equal(t, kase.want, output)
		})
	}
}
