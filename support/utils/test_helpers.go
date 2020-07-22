package utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var epsilon = 0.00000001

// AssetFloatEquals is a float comparison within a pre-defined epsilon error
func AssetFloatEquals(t *testing.T, want float64, actual float64) {
	if want == 0.0 {
		assert.True(t, math.Abs(want-actual) <= epsilon, fmt.Sprintf("expected: %.f, actual: %.f", want, actual))
	} else {
		assert.InEpsilon(t, want, actual, epsilon)
	}
}
