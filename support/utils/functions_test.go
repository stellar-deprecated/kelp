package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashString(t *testing.T) {
	testCases := []struct {
		s        string
		wantHash uint32
	}{
		{
			s:        "hello",
			wantHash: 0x4f9f2cab,
		}, {
			s:        "world",
			wantHash: 0x37a3e893,
		}, {
			s:        "1",
			wantHash: 0x340ca71c, // same as test for ToJSONHash(int(1)) below
		},
	}

	for _, kase := range testCases {
		t.Run(kase.s, func(t *testing.T) {
			result, e := HashString(kase.s)
			if !assert.Nil(t, e) {
				return
			}

			assert.Equal(t, kase.wantHash, result)
		})
	}
}

func TestToJSONHash(t *testing.T) {
	testCases := []struct {
		v        interface{}
		wantHash uint32
	}{
		{
			v:        "hello",
			wantHash: 0xdf47ee8b, // different from HashString result in test above
		}, {
			v:        "world",
			wantHash: 0x9a7c4703, // different from HashString result in test above
		}, {
			v:        "true",
			wantHash: 0x4f9ddea5,
		}, {
			v:        true, // should be different from "true"
			wantHash: 0x4db211e5,
		}, {
			v:        "1", // different from HashString result in test above
			wantHash: 0x12a777a,
		}, {
			v:        int(1),     // should be different from "1"
			wantHash: 0x340ca71c, // this is same as HashString result in test above for string 1
		}, {
			v:        1.0,
			wantHash: 0x340ca71c, // JSON treats this as the same as int(1) so expected value is the same
		}, {
			v:        1.0000001,
			wantHash: 0xddd9e8d5, // should be different from "1", and int(1) and 1.0
		},
	}

	for _, kase := range testCases {
		t.Run(fmt.Sprintf("%v", kase.v), func(t *testing.T) {
			result, e := ToJSONHash(kase.v)
			if !assert.Nil(t, e) {
				return
			}

			assert.Equal(t, kase.wantHash, result)
		})
	}
}

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
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, success, m)
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

	// Note that we lose type information in non-typed structs and arrays.
	// This is because deserializing JSON collections will lose type information.
	// (The exception is when JSON tags are used.)
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

	// Note that since the JSON standard only has number, and the package
	// defaults to numbers as float64 absent marshaling into a specific struct, we expect
	// that integers are transformed into floats.
	want := map[string]interface{}{
		"True":           true,
		"False":          false,
		"EmptyString":    "",
		"NonEmptyString": "hello world",
		"Int":            1.0,
		"Float":          1.0,
		"EmptyStruct":    map[string]interface{}{},
		"NonEmptyStruct": map[string]interface{}{"value": 1.0},
		"EmptyInts":      []interface{}{},
		"NonEmptyInts":   []interface{}{1.0, 2.0},
	}

	m, e := ToMapStringInterface(success)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, want, m)
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
		"2": 2,
		"3": 3,
		"4": "four",
	}
	wantMap := map[string]interface{}{
		"1": 1,
		"2": 2,
		"3": 3,
		"4": "four",
	}

	gotMap, e := MergeMaps(original, override)
	if !assert.NoError(t, e) {
		return
	}

	assert.Equal(t, wantMap, gotMap)
}
