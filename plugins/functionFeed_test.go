package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFunctionParts(t *testing.T) {
	testCases := []struct {
		inputURL string
		wantName string
		wantArgs string
	}{
		{
			inputURL: "max(test)",
			wantName: "max",
			wantArgs: "test",
		}, {
			inputURL: "invert(max(test))",
			wantName: "invert",
			wantArgs: "max(test)",
		},
	}

	for _, k := range testCases {
		t.Run(k.inputURL, func(t *testing.T) {
			fnName, fnArgs, e := extractFunctionParts(k.inputURL)
			if !assert.NoError(t, e) {
				return
			}

			assert.Equal(t, k.wantName, fnName)
			assert.Equal(t, k.wantArgs, fnArgs)
		})
	}
}
