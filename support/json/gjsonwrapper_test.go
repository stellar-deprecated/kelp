package json

import (
	"encoding/json"
	"fmt"
	"github.com/stellar/kelp/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGJsonWrapper_GetRawJsonValue_Error(t *testing.T) {
	response := TestJsonParserWrapperResponse{
		Data: TestJsonParserWrapperRates{
			Raw: nil,
		},
	}

	json, _ := json.Marshal(response)
	path := "data.raw.non_existent_field"

	jsonParserWrapper := NewJsonParserWrapper()

	rawValue, err := jsonParserWrapper.GetRawJsonValue(json, path)

	assert.EqualError(t, err, fmt.Sprintf("json parser wrapper error: could not find json for path %s in %s", path, json))
	assert.Equal(t, "", rawValue)
}

func TestGJsonWrapper_GetRawJsonValue_RawValue(t *testing.T) {
	target := tests.RandomString()

	raw := make(map[string]string)

	r := tests.RandomString()

	raw[r] = target
	raw[tests.RandomString()] = tests.RandomString()
	raw[tests.RandomString()] = tests.RandomString()
	raw[tests.RandomString()] = tests.RandomString()

	response := TestJsonParserWrapperResponse{
		Data: TestJsonParserWrapperRates{
			Raw: raw,
		},
	}

	json, _ := json.Marshal(response)
	path := fmt.Sprintf("data.raw.%s", r)

	jsonParserWrapper := NewJsonParserWrapper()

	actual, err := jsonParserWrapper.GetRawJsonValue(json, path)

	expected := fmt.Sprintf("\"%s\"", target)

	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}

func TestGJsonWrapper_GetNum_Input_Float64(t *testing.T) {
	expected := tests.RandomFloat64()

	float := make(map[string]float64)

	f := tests.RandomString()

	float[f] = expected
	float[tests.RandomString()] = tests.RandomFloat64()
	float[tests.RandomString()] = tests.RandomFloat64()
	float[tests.RandomString()] = tests.RandomFloat64()

	response := TestJsonParserWrapperResponse{
		Data: TestJsonParserWrapperRates{
			Float: float,
		},
	}

	json, _ := json.Marshal(response)
	path := fmt.Sprintf("data.float.%s", f)

	jsonParserWrapper := NewJsonParserWrapper()

	actual, err := jsonParserWrapper.GetNum(json, path)

	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}

func TestGJsonWrapper_GetNum_Input_Int(t *testing.T) {
	expected := tests.RandomInt()

	number := make(map[string]int)

	num := tests.RandomString()

	number[num] = expected
	number[tests.RandomString()] = tests.RandomInt()
	number[tests.RandomString()] = tests.RandomInt()
	number[tests.RandomString()] = tests.RandomInt()

	response := TestJsonParserWrapperResponse{
		Data: TestJsonParserWrapperRates{
			Number: number,
		},
	}

	json, _ := json.Marshal(response)
	path := fmt.Sprintf("data.number.%s", num)

	jsonParserWrapper := NewJsonParserWrapper()

	actual, err := jsonParserWrapper.GetNum(json, path)

	assert.Equal(t, float64(expected), actual)
	assert.Nil(t, err)
}

type TestJsonParserWrapperResponse struct {
	Data TestJsonParserWrapperRates `json:"data"`
}

type TestJsonParserWrapperRates struct {
	Raw    map[string]string  `json:"raw"`
	Float  map[string]float64 `json:"float"`
	Number map[string]int     `json:"number"`
}
