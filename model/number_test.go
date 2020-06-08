package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumberFromFloat(t *testing.T) {
	testCases := []struct {
		f          float64
		precision  int8
		wantString string
		wantFloat  float64
	}{
		{
			f:          1.1,
			precision:  1,
			wantString: "1.1",
			wantFloat:  1.1,
		}, {
			f:          1.1,
			precision:  2,
			wantString: "1.10",
			wantFloat:  1.10,
		}, {
			f:          1.12,
			precision:  1,
			wantString: "1.1",
			wantFloat:  1.1,
		}, {
			f:          1.15,
			precision:  1,
			wantString: "1.2",
			wantFloat:  1.2,
		}, {
			f:          0.12,
			precision:  1,
			wantString: "0.1",
			wantFloat:  0.1,
		},
	}

	for _, kase := range testCases {
		t.Run(fmt.Sprintf("%f_%d", kase.f, kase.precision), func(t *testing.T) {
			n := NumberFromFloat(kase.f, kase.precision)
			if !assert.Equal(t, kase.wantString, n.AsString()) {
				return
			}
			if !assert.Equal(t, kase.wantFloat, n.AsFloat()) {
				return
			}
		})
	}
}

func TestMath(t *testing.T) {
	testCases := []struct {
		n1           *Number
		n2           *Number
		wantAdd      float64
		wantSubtract float64
		wantMultiply float64
		wantDivide   float64
	}{
		{
			n1:           NumberFromFloat(1.1, 1),
			n2:           NumberFromFloat(2.1, 1),
			wantAdd:      3.2,
			wantSubtract: -1.0,
			wantMultiply: 2.3,
			wantDivide:   0.5,
		}, {
			n1:           NumberFromFloat(1.15, 1),
			n2:           NumberFromFloat(2.1, 1),
			wantAdd:      3.3,
			wantSubtract: -0.9,
			wantMultiply: 2.5,
			wantDivide:   0.6,
		}, {
			n1:           NumberFromFloat(1.15, 2),
			n2:           NumberFromFloat(2.1, 1),
			wantAdd:      3.3,
			wantSubtract: -1.0,
			wantMultiply: 2.4,
			wantDivide:   0.5,
		}, {
			n1:           NumberFromFloat(1.15, 2),
			n2:           NumberFromFloat(2.1, 2),
			wantAdd:      3.25,
			wantSubtract: -0.95,
			wantMultiply: 2.42,
			wantDivide:   0.55,
		}, {
			n1:           NumberFromFloat(1.12, 2),
			n2:           NumberFromFloat(2.1, 1),
			wantAdd:      3.2,
			wantSubtract: -1.0,
			wantMultiply: 2.4,
			wantDivide:   0.5,
		},
	}

	for i, kase := range testCases {
		t.Run(fmt.Sprintf("%d__%f_%d__%f_%d", i, kase.n1.AsFloat(), kase.n1.Precision(), kase.n2.AsFloat(), kase.n2.Precision()), func(t *testing.T) {
			n := kase.n1.Add(*kase.n2)
			if !assert.Equal(t, kase.wantAdd, n.AsFloat()) {
				return
			}

			n = kase.n1.Subtract(*kase.n2)
			if !assert.Equal(t, kase.wantSubtract, n.AsFloat()) {
				return
			}

			n = kase.n1.Multiply(*kase.n2)
			if !assert.Equal(t, kase.wantMultiply, n.AsFloat()) {
				return
			}

			n = kase.n1.Divide(*kase.n2)
			if !assert.Equal(t, kase.wantDivide, n.AsFloat()) {
				return
			}
		})
	}
}

func TestScale(t *testing.T) {
	testCases := []struct {
		n           *Number
		scaleFactor float64
		wantString  string
		wantFloat   float64
	}{
		{
			n:           NumberFromFloat(1.1, 1),
			scaleFactor: 2.103,
			wantString:  "2.3",
			wantFloat:   2.3,
		}, {
			n:           NumberFromFloat(1.1, 2),
			scaleFactor: 2.103,
			wantString:  "2.31",
			wantFloat:   2.31,
		}, {
			n:           NumberFromFloat(1.1, 2),
			scaleFactor: 1 / 2.103,
			wantString:  "0.52",
			wantFloat:   0.52,
		},
	}

	for i, kase := range testCases {
		t.Run(fmt.Sprintf("%d__%f_%d__%f", i, kase.n.AsFloat(), kase.n.Precision(), kase.scaleFactor), func(t *testing.T) {
			n := kase.n.Scale(kase.scaleFactor)

			if !assert.Equal(t, kase.wantString, n.AsString()) {
				return
			}

			if !assert.Equal(t, kase.wantFloat, n.AsFloat()) {
				return
			}
		})
	}
}

func TestEqualsPrecisionNormalized(t *testing.T) {
	testCases := []struct {
		n1      *Number
		n2      *Number
		epsilon float64
		want    bool
	}{
		{
			n1:      NumberFromFloat(2.0, 1),
			n2:      NumberFromFloat(1.0, 1),
			epsilon: 0.01,
			want:    false,
		}, {
			n1:      NumberFromFloat(1.0, 1),
			n2:      NumberFromFloat(2.0, 1),
			epsilon: 0.01,
			want:    false,
		}, {
			n1:      NumberFromFloat(-1.0, 1),
			n2:      NumberFromFloat(1.0, 1),
			epsilon: 0.01,
			want:    false,
		}, {
			n1:      NumberFromFloat(1.0, 1),
			n2:      NumberFromFloat(-1.0, 1),
			epsilon: 0.01,
			want:    false,
		}, {
			n1:      NumberFromFloat(-1.0, 1),
			n2:      NumberFromFloat(-1.0, 1),
			epsilon: 0.01,
			want:    true,
		}, {
			n1:      NumberFromFloat(0.0, 2),
			n2:      NumberFromFloat(0.0, 1),
			epsilon: 0.01,
			want:    true,
		}, {
			n1:      NumberFromFloat(2.1001, 4),
			n2:      NumberFromFloat(2.10009, 5),
			epsilon: 0.00001,
			want:    true,
		}, {
			n1:      NumberFromFloat(2.1001, 4),
			n2:      NumberFromFloat(2.10009, 5),
			epsilon: 0.0001,
			want:    true,
		},
	}

	for i, kase := range testCases {
		t.Run(fmt.Sprintf("%d__%f_%d__%f_%d", i, kase.n1.AsFloat(), kase.n1.Precision(), kase.n2.AsFloat(), kase.n2.Precision()), func(t *testing.T) {
			res := kase.n1.EqualsPrecisionNormalized(*kase.n2, kase.epsilon)
			assert.Equal(t, kase.want, res)
		})
	}
}

func TestUnaryOperations(t *testing.T) {
	testCases := []struct {
		n          *Number
		wantAbs    float64
		wantNegate float64
		wantInvert float64
	}{
		{
			n:          NumberFromFloat(-0.2, 1),
			wantAbs:    0.2,
			wantNegate: 0.2,
			wantInvert: -5.0,
		}, {
			n:          NumberFromFloat(0.2812, 3),
			wantAbs:    0.281,
			wantNegate: -0.281,
			wantInvert: 3.558718861209964,
		},
	}

	for _, kase := range testCases {
		t.Run(kase.n.AsString(), func(t *testing.T) {
			abs := kase.n.Abs()
			if !assert.Equal(t, kase.wantAbs, abs.AsFloat()) {
				return
			}

			negative := kase.n.Negate()
			if !assert.Equal(t, kase.wantNegate, negative.AsFloat()) {
				return
			}

			inverted := InvertNumber(kase.n)
			if !assert.Equal(t, kase.wantInvert, inverted.AsFloat()) {
				return
			}
			if !assert.Equal(t, int8(15), inverted.precision) {
				return
			}
		})
	}
}

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
		}, {
			n:     NumberFromFloat(5274.26, 8),
			wantN: 263713,
			wantD: 50,
		}, {
			n:     NumberFromFloat(-5274.26, 8),
			wantN: -263713,
			wantD: 50,
		}, {
			n:     NumberFromFloat(10.0, 4),
			wantN: 10,
			wantD: 1,
		}, {
			n:     NumberFromFloat(-10.0, 4),
			wantN: -10,
			wantD: 1,
		}, {
			n:     NumberFromFloat(2.599999999, 9),
			wantN: 1039999997,
			wantD: 399999999,
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
