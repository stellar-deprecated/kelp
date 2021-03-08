package model

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"

	"github.com/stellar/go/price"
)

// NumberConstants holds some useful constants
var NumberConstants = struct {
	Zero *Number
	One  *Number
}{
	Zero: NumberFromFloat(0.0, 16),
	One:  NumberFromFloat(1.0, 16),
}

// InvertPrecision is the precision of the number after it is inverted
// this is only 11 becuase if we keep it larger such as 15 then inversions are inaccurate for larger numbers such as inverting 0.00002
const InvertPrecision = 11

// InternalCalculationsPrecision is the precision to be used for internal calculations in a function
const InternalCalculationsPrecision = 15

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
	return fmt.Sprintf(fmt.Sprintf("%%.%df", n.Precision()), n.AsFloat())
}

// AsRatio returns an integer numerator and denominator
func (n Number) AsRatio() (int32, int32, error) {
	p, e := price.Parse(n.AsString())
	if e != nil {
		return 0, 0, fmt.Errorf("unable to convert number to ratio: %s", e)
	}
	return int32(p.N), int32(p.D), nil
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

// Multiply returns a new Number after multiplying with the passed in Number by rounding up based on the smaller precision
func (n Number) Multiply(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloat(n.AsFloat()*n2.AsFloat(), newPrecision)
}

// MultiplyRoundTruncate returns a new Number after multiplying with the passed in Number by truncating based on the smaller precision
func (n Number) MultiplyRoundTruncate(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloatRoundTruncate(n.AsFloat()*n2.AsFloat(), newPrecision)
}

// Divide returns a new Number after dividing by the passed in Number by rounding up based on the smaller precision
func (n Number) Divide(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloat(n.AsFloat()/n2.AsFloat(), newPrecision)
}

// DivideRoundTruncate returns a new Number after dividing by the passed in Number by truncating based on the smaller precision
func (n Number) DivideRoundTruncate(n2 Number) *Number {
	newPrecision := minPrecision(n, n2)
	return NumberFromFloatRoundTruncate(n.AsFloat()/n2.AsFloat(), newPrecision)
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

// NumberFromFloat makes a Number from a float by rounding up
func NumberFromFloat(f float64, precision int8) *Number {
	return &Number{
		value:     toFixed(f, precision, RoundUp),
		precision: precision,
	}
}

// NumberFromFloatRoundTruncate makes a Number from a float by truncating beyond the specified precision
func NumberFromFloatRoundTruncate(f float64, precision int8) *Number {
	return &Number{
		value:     toFixed(f, precision, RoundTruncate),
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

	// return 0 for the inverse of 0 to keep it safe
	if n.AsFloat() == 0 {
		log.Printf("trying to invert the number 0, returning the same number to keep it safe")
		return n
	}

	bigNum := big.NewRat(1, 1)
	bigNum = bigNum.SetFloat64(n.AsFloat())
	bigInv := bigNum.Inv(bigNum)

	bigInvFloat64, _ := bigInv.Float64()
	return NumberFromFloat(bigInvFloat64, InvertPrecision)
}

// NumberByCappingPrecision returns a number with a precision that is at max the passed in precision
func NumberByCappingPrecision(n *Number, precision int8) *Number {
	if n.Precision() > precision {
		return NumberFromFloat(n.AsFloat(), precision)
	}
	return n
}

func round(num float64, rounding Rounding) int64 {
	if rounding == RoundUp {
		return int64(num + math.Copysign(0.5, num))
	} else if rounding == RoundTruncate {
		return int64(num)
	} else {
		panic(fmt.Sprintf("unknown rounding type %v", rounding))
	}
}

// Rounding is a type that defines various approaching to rounding numbers
type Rounding int

// Rounding types
const (
	RoundUp Rounding = iota
	RoundTruncate
)

func toFixed(num float64, precision int8, rounding Rounding) float64 {
	bigNum := big.NewRat(1, 1)
	bigNum = bigNum.SetFloat64(num)
	bigPow := big.NewRat(1, 1)
	bigPow = bigPow.SetFloat64(math.Pow(10, float64(precision)))

	// multiply
	bigMultiply := bigNum.Mul(bigNum, bigPow)

	// convert to int after rounding
	bigMultiplyFloat64, _ := bigMultiply.Float64()
	roundedInt64 := round(bigMultiplyFloat64, rounding)
	bigMultiplyIntFloat64 := big.NewRat(1, 1)
	bigMultiplyIntFloat64 = bigMultiplyIntFloat64.SetInt64(roundedInt64)

	// divide it
	bigPowInverse := bigPow.Inv(bigPow)
	bigResult := bigMultiply.Mul(bigMultiplyIntFloat64, bigPowInverse)

	br, _ := bigResult.Float64()
	return br
}

func minPrecision(n1 Number, n2 Number) int8 {
	if n1.precision < n2.precision {
		return n1.precision
	}
	return n2.precision
}
