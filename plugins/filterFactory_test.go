package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIdsArray(t *testing.T) {
	testCases := []struct {
		arrayString string
		want        []string
	}{
		{
			arrayString: "[abcde1234Z,01234gFHij]",
			want:        []string{"abcde1234Z", "01234gFHij"},
		}, {
			arrayString: "[abcde1234Z, 01234gFHij]",
			want:        []string{"abcde1234Z", "01234gFHij"},
		}, {
			arrayString: "[abcde1234Z]",
			want:        []string{"abcde1234Z"},
		}, {
			arrayString: "[]",
			want:        []string{},
		},
	}

	for _, kase := range testCases {
		t.Run(kase.arrayString, func(t *testing.T) {
			output, e := parseIdsArray(kase.arrayString)
			if !assert.NoError(t, e) {
				return
			}

			assert.Equal(t, kase.want, output)
		})
	}
}
