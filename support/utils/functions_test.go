package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDedupe(t *testing.T) {
	testCases := []struct {
		input []string
		want  []string
	}{
		{
			input: []string{"a", "a", "b"},
			want:  []string{"a", "b"},
		}, {
			input: []string{"a", "b", "b"},
			want:  []string{"a", "b"},
		}, {
			input: []string{"a", "b", "a"},
			want:  []string{"a", "b"},
		}, {
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
	}

	for _, kase := range testCases {
		t.Run(fmt.Sprintf("%s", kase.input), func(t *testing.T) {
			output := Dedupe(kase.input)
			if !assert.Equal(t, kase.want, output) {
				return
			}
		})
	}
}
