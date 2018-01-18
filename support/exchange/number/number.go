package number

import (
	"log"
	"math"
	"strconv"
)

const precision = 7

// Number abstraction
type Number struct {
	f float64
	s string
}

// FromFloat makes a Number from a float
func FromFloat(f float64) *Number {
	fixed := toFixed(f, precision)
	return &Number{
		f: fixed,
		s: strconv.FormatFloat(fixed, 'f', precision, 64),
	}
}

// FromString makes a Number from a string, by calling FromFloat
func FromString(s string) (*Number, error) {
	parsed, e := strconv.ParseFloat(s, 64)
	if e != nil {
		return nil, e
	}
	return FromFloat(parsed), nil
}

// MustFromString panics when there's an error
func MustFromString(s string) *Number {
	parsed, e := strconv.ParseFloat(s, 64)
	if e != nil {
		log.Panic(e)
	}
	return FromFloat(parsed)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, prec int) float64 {
	output := math.Pow(10, float64(prec))
	return float64(round(num*output)) / output
}
