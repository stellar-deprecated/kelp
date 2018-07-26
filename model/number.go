package model

import (
	"log"
	"math"
	"strconv"
)

// Number abstraction
type Number struct {
	value     float64
	precision int8
}

// AsFloat gives a float64 representation
func (n Number) AsFloat() float64 {
	return n.value
}

// Precision gives the precision of the Number
func (n Number) Precision() int8 {
	return n.precision
}

// AsString gives a string representation
func (n Number) AsString() string {
	return strconv.FormatFloat(n.AsFloat(), 'f', int(n.Precision()), 64)
}

// String is the Stringer interface impl.
func (n Number) String() string {
	return n.AsString()
}

// NumberFromFloat makes a Number from a float
func NumberFromFloat(f float64, precision int8) *Number {
	return &Number{
		value:     toFixed(f, precision),
		precision: precision,
	}
}

// NumberFromString makes a Number from a string, by calling NumberFromFloat
func NumberFromString(s string, precision int8) (*Number, error) {
	parsed, e := strconv.ParseFloat(s, 64)
	if e != nil {
		return nil, e
	}
	return NumberFromFloat(parsed, precision), nil
}

// MustNumberFromString panics when there's an error
func MustNumberFromString(s string, precision int8) *Number {
	parsed, e := NumberFromString(s, precision)
	if e != nil {
		log.Panic(e)
	}
	return parsed
}

// InvertNumber inverts a number, returns nil if the original number is nil, preserves precision
func InvertNumber(n *Number) *Number {
	if n == nil {
		return nil
	}
	return NumberFromFloat(1/n.AsFloat(), n.Precision())
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int8) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
