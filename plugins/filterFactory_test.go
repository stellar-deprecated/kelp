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
			marketIdsArrayString: "[abcde1234Z,01234gFHij]",
			want:                 []string{"abcde1234Z", "01234gFHij"},
		}, {
			marketIdsArrayString: "[abcde1234Z, 01234gFHij]",
			want:                 []string{"abcde1234Z", "01234gFHij"},
		}, {
			marketIdsArrayString: "[abcde1234Z]",
			want:                 []string{"abcde1234Z"},
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
