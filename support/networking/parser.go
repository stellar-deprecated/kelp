package networking

import (
	"fmt"
	"reflect"

	"github.com/stellar/kelp/model"
)

const numberPrecision = 10

// PrefixFieldNotFound is what is returned in the error when we cannot find a field in the map
const PrefixFieldNotFound = "could not find field in map"

func checkKeyPresent(m map[string]interface{}, key string) (interface{}, error) {
	v, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("%s: %s", PrefixFieldNotFound, key)
	}

	return v, nil
}

func makeParseError(field string, dataType string, methodAPI string, value interface{}) error {
	return fmt.Errorf("could not parse the field '%s' as a %s in the response from %s: value=%v, type=%s", field, dataType, methodAPI, value, reflect.TypeOf(value))
}

// ParseString helps to parse a string value out of the map
func ParseString(m map[string]interface{}, key string, methodAPI string) (string, error) {
	v, e := checkKeyPresent(m, key)
	if e != nil {
		return "", e
	}

	s, ok := v.(string)
	if !ok {
		return "", makeParseError(key, "string", methodAPI, v)
	}

	return s, nil
}

// ParseBool helps to parse a bool value out of the map
func ParseBool(m map[string]interface{}, key string, methodAPI string) (bool, error) {
	v, e := checkKeyPresent(m, key)
	if e != nil {
		return false, e
	}

	b, ok := v.(bool)
	if !ok {
		return false, makeParseError(key, "bool", methodAPI, v)
	}

	return b, nil
}

// ParseNumber helps to parse a model.Number value out of the map
func ParseNumber(m map[string]interface{}, key string, methodAPI string) (*model.Number, error) {
	v, e := checkKeyPresent(m, key)
	if e != nil {
		return nil, e
	}

	switch v.(type) {
	case string:
		return parseStringAsNumber(m, key, methodAPI)
	case float64:
		return parseFloatAsNumber(m, key, methodAPI)
	default:
		return nil, makeParseError(key, "number", methodAPI, v)
	}
}

func parseStringAsNumber(m map[string]interface{}, key string, methodAPI string) (*model.Number, error) {
	s, e := ParseString(m, key, methodAPI)
	if e != nil {
		return nil, e
	}

	n, e := model.NumberFromString(s, numberPrecision)
	if e != nil {
		return nil, fmt.Errorf("unable to convert the string field '%s' to a number in the response from %s: value=%v, error=%s", key, methodAPI, s, e)
	}

	return n, nil
}

func parseFloatAsNumber(m map[string]interface{}, key string, methodAPI string) (*model.Number, error) {
	v, e := checkKeyPresent(m, key)
	if e != nil {
		return nil, e
	}

	f, ok := v.(float64)
	if !ok {
		return nil, makeParseError(key, "float", methodAPI, v)
	}

	return model.NumberFromFloat(f, numberPrecision), nil
}
