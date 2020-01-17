package plugins

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/stellar/kelp/api"
)

type functionFeed struct {
	getPriceFn func() (float64, error)
}

var _ api.PriceFeed = &functionFeed{}

// makeFeedFromFn is a convenience factory method that converts a GetPrice function in a PriceFeed struct
func makeFunctionFeed(fn func() (float64, error)) api.PriceFeed {
	return &functionFeed{
		getPriceFn: fn,
	}
}

// GetPrice delegator function
func (f *functionFeed) GetPrice() (float64, error) {
	return f.getPriceFn()
}

func makeFunctionPriceFeed(url string) (api.PriceFeed, error) {
	name, argsString, e := extractFunctionParts(url)
	if e != nil {
		return nil, fmt.Errorf("unable to extract function name from URL: %s", e)
	}

	f, ok := fnFactoryMap[name]
	if !ok {
		return nil, fmt.Errorf("the passed in URL does not have the registered function '%s'", name)
	}

	feeds, e := makeFeedsArray(argsString)
	if e != nil {
		return nil, fmt.Errorf("error when makings feeds array: %s", e)
	}

	pf, e := f(feeds)
	if e != nil {
		return nil, fmt.Errorf("error when invoking price feed function '%s': %s", name, e)
	}

	return pf, nil
}

func extractFunctionParts(url string) (name string, args string, e error) {
	fnNameRegex, e := regexp.Compile("^([a-zA-Z]+)\\((.*)\\)$")
	if e != nil {
		return "", "", fmt.Errorf("unable to make regexp (programmer error)")
	}

	submatches := fnNameRegex.FindStringSubmatch(url)
	if len(submatches) != 3 {
		return "", "", fmt.Errorf("incorrect number of matches, expected 3 entries in the returned array (matchedString, subgroup1, subgroup2), but found %v", submatches)
	}

	return submatches[1], submatches[2], nil
}

func makeFeedsArray(feedsStringCSV string) ([]api.PriceFeed, error) {
	parts := strings.Split(feedsStringCSV, ",")
	arr := []api.PriceFeed{}

	for _, argPart := range parts {
		feedSpecParts := strings.SplitN(argPart, "/", 2)
		if len(feedSpecParts) != 2 {
			return nil, fmt.Errorf("unable to correctly split arg into a price feed spec: %s", argPart)
		}
		priceFeedType := feedSpecParts[0]
		priceFeedURL := feedSpecParts[1]

		feed, e := MakePriceFeed(priceFeedType, priceFeedURL)
		if e != nil {
			return nil, fmt.Errorf("error creating a price feed (typ='%s', url='%s'): %s", priceFeedType, priceFeedURL, e)
		}
		arr = append(arr, feed)
	}

	return arr, nil
}
