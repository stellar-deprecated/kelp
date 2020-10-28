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

func TestToMapStringInterface_SuccessMap(t *testing.T) {
	success := map[string]interface{}{
		"test": true,
	}
	m, e := ToMapStringInterface(success)
	assert.Equal(t, success, m)
	assert.NoError(t, e)
}

func TestToMapStringInterface_SuccessStruct(t *testing.T) {
	type NonEmptyStruct struct {
		Value int `json:"value"`
	}

	type SuccessStruct struct {
		True           bool
		False          bool
		EmptyString    string
		NonEmptyString string
		Int            int
		Float          float64
		EmptyStruct    struct{}
		NonEmptyStruct NonEmptyStruct
		EmptyInts      []int
		NonEmptyInts   []int
	}

	success := SuccessStruct{
		True:           true,
		False:          false,
		EmptyString:    "",
		NonEmptyString: "hello world",
		Int:            1,
		Float:          1.0,
		EmptyStruct:    struct{}{},
		NonEmptyStruct: NonEmptyStruct{1},
		EmptyInts:      []int{},
		NonEmptyInts:   []int{1, 2},
	}

	want := map[string]interface{}{
		"EmptyInts":      []interface{}{},
		"EmptyString":    "",
		"EmptyStruct":    map[string]interface{}{},
		"False":          false,
		"Float":          1.,
		"Int":            1.,
		"NonEmptyInts":   []interface{}{1., 2.},
		"NonEmptyString": "hello world",
		"NonEmptyStruct": map[string]interface{}{"value": 1.},
		"True":           true,
	}

	m, e := ToMapStringInterface(success)
	assert.Equal(t, want, m)
	assert.NoError(t, e)
}

func TestToMapStringInterface_FailureFloat(t *testing.T) {
	failure := 0.0
	m, e := ToMapStringInterface(failure)
	assert.EqualError(t, e, "could not unmarshal json to interface: json: cannot unmarshal number into Go value of type map[string]interface {}")
	assert.Nil(t, m)
}

func TestMergeMaps(t *testing.T) {
	original := map[string]interface{}{
		"1": 1,
		"2": "two",
	}
	override := map[string]interface{}{
		"3": 3,
		"4": "four",
	}
	wantMap := map[string]interface{}{
		"1": 1,
		"2": "two",
		"3": 3,
		"4": "four",
	}

	gotMap, e := MergeMaps(original, override)
	assert.Equal(t, wantMap, gotMap)
	assert.NoError(t, e)
}
