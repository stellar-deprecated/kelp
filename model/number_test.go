package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsRatio(t *testing.T) {
	testCases := []struct {
		n     *Number
		wantN int32
		wantD int32
	}{
		{
			n:     NumberFromFloat(0.251523, 6),
			wantN: 251523,
			wantD: 1000000,
		}, {
			n:     NumberFromFloat(-0.251523, 6),
			wantN: -251523,
			wantD: 1000000,
		}, {
			n:     NumberFromFloat(0.251841, 6),
			wantN: 251841,
			wantD: 1000000,
		}, {
			n:     NumberFromFloat(-0.251841, 6),
			wantN: -251841,
			wantD: 1000000,
		},
	}

	for _, kase := range testCases {
		t.Run(kase.n.AsString(), func(t *testing.T) {
			num, den, e := kase.n.AsRatio()
			if !assert.NoError(t, e) {
				return
			}

			if !assert.Equal(t, kase.wantN, num) {
				return
			}
			assert.Equal(t, kase.wantD, den)
		})
	}
}

func TestAsRatio_Error(t *testing.T) {
	testCases := []struct {
		n *Number
	}{
		{
			n: NumberFromFloat(1.0, 10),
		}, {
			n: NumberFromFloat(2.5, 9),
		}, {
			n: NumberFromFloat(-2.5, 9),
		},
	}

	for _, kase := range testCases {
		t.Run(kase.n.AsString(), func(t *testing.T) {
			num, den, e := kase.n.AsRatio()
			if !assert.Error(t, e, fmt.Sprintf("got back num=%d, den=%d", num, den)) {
				return
			}
		})
	}
}
