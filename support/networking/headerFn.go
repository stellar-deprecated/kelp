package networking

import (
	"fmt"
	"strings"

	"github.com/stellar/kelp/support/utils"
)

// HeaderFn represents a function that transforms headers
type HeaderFn func(string, string, string) string // (string httpMethod, string requestPath, string body)

// makeStaticHeaderFn is a convenience method
func makeStaticHeaderFn(value string) (HeaderFn, error) {
	// need to convert to HeaderFn to work as a api.ExchangeHeader.Value
	return HeaderFn(func(method string, requestPath string, body string) string {
		return value
	}), nil
}

// HeaderFnFactory is a factory method for the HeaderFn
type HeaderFnFactory func(string) (HeaderFn, error)

var defaultMappings = map[string]HeaderFnFactory{
	"STATIC": HeaderFnFactory(makeStaticHeaderFn),
}

func headerFnNames(maps ...map[string]HeaderFnFactory) []string {
	names := []string{}
	for _, m := range maps {
		if m != nil {
			for k, _ := range m {
				names = append(names, k)
			}
		}
	}
	return utils.Dedupe(names)
}

// MakeHeaderFn is a factory method that makes a HeaderFn
func MakeHeaderFn(value string, primaryMappings map[string]HeaderFnFactory) (HeaderFn, error) {
	numSeparators := strings.Count(value, ":")

	if numSeparators == 0 {
		// LOH-1 - support backward-compatible case of not having any pre-specified function
		return makeStaticHeaderFn(value)
	} else if numSeparators != 1 {
		names := headerFnNames(primaryMappings, defaultMappings)
		return nil, fmt.Errorf("invalid format of header value (%s), needs exactly one colon (:) to separate the header function from the input value to that function. list of available header functions: [%s]", value, strings.Join(names, ", "))
	}

	valueParts := strings.Split(value, ":")
	fnType := valueParts[0]
	fnInputValue := valueParts[1]

	if primaryMappings != nil {
		if makeHeaderFn, ok := primaryMappings[fnType]; ok {
			return makeHeaderFn(fnInputValue)
		}
	}

	if makeHeaderFn, ok := defaultMappings[fnType]; ok {
		return makeHeaderFn(fnInputValue)
	}

	names := headerFnNames(primaryMappings, defaultMappings)
	return nil, fmt.Errorf("invalid function prefix (%s) as part of header value (%s). list of available header functions: [%s]", fnType, value, strings.Join(names, ", "))
}
