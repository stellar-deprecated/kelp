package model

import (
	"fmt"
	"log"
	"math"
	"strconv"
)

// NumberConstants holds some useful constants
var NumberConstants = struct {
	Zero *Number
	One  *Number
}{
	Zero: NumberFromFloat(0.0, 16),
	One:  NumberFromFloat(1.0, 16),
}

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

// AsRatio returns an integer numerator and denominator
func (n Number) AsRatio() (int32, int32, error) {
	denominator := int32(math.Pow(10, float64(n.Precision())))

	// add an adjustment because the computed value should not have any digits beyond the decimal
	// and we want to roll over values that are not computed correctly rather than using the expensive math/big library
	adjustment := 0.1
	if n.AsFloat() < 0 {
		adjustment = -0.1
	}
	numerator := int32(n.AsFloat()*float64(denominator) + adjustment)

	if float64(numerator)/float64(denominator) != n.AsFloat() {
		return 0, 0, fmt.Errorf("invalid conversion to a ratio probably caused by an overflow, float input: %f, numerator: %d, denominator: %d", n.AsFloat(), numerator, denominator)
	}

	return numerator, denominator, nil
}

// Abs returns the absolute of the number
func (n Number) Abs() *Number {
	if n.AsFloat() < 0 {
		return n.Negate()
	}
	return &n
}

// Negate returns the negative value of the number
func (n Number) Negate() *Number {
	return NumberConstants.Zero.Subtract(n)
}

// Add returns a new Number after adding the passed in Number
func (n Number) Add(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloat(n.AsFloat()+n2.AsFloat(), newPrecision)
}

// Subtract returns a new Number after subtracting out the passed in Number
func (n Number) Subtract(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloat(n.AsFloat()-n2.AsFloat(), newPrecision)
}

// Multiply returns a new Number after multiplying with the passed in Number
func (n Number) Multiply(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloat(n.AsFloat()*n2.AsFloat(), newPrecision)
}

// Divide returns a new Number after dividing by the passed in Number
func (n Number) Divide(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloat(n.AsFloat()/n2.AsFloat(), newPrecision)
}

// Scale takes in a scalar with which to multiply the number using the same precision of the original number
func (n Number) Scale(scaleFactor float64) *Number {
	return NumberFromFloat(n.AsFloat()*scaleFactor, n.precision)
}

// EqualsPrecisionNormalized returns true if the two numbers are the same after comparing them at the same (lowest) precision level
func (n Number) EqualsPrecisionNormalized(n2 Number, epsilon float64) bool {
	return n.Subtract(n2).Abs().AsFloat() < epsilon
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
		log.Fatal(e)
	}
	return parsed
}

// InvertNumber inverts a number, returns nil if the original number is nil, preserves precision
func InvertNumber(n *Number) *Number {
	if n == nil {
		return nil
	}
	return NumberConstants.One.Divide(*n)
}

// NumberByCappingPrecision returns a number with a precision that is at max the passed in precision
func NumberByCappingPrecision(n *Number, precision int8) *Number {
	if n.Precision() > precision {
		return NumberFromFloat(n.AsFloat(), precision)
	}
	return n
}

func round(num float64) int64 {
	return int64(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int8) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func minPrecision(n1 Number, n2 Number) int8 {
	if n1.precision < n2.precision {
		return n1.precision
	}
	return n2.precision
}
