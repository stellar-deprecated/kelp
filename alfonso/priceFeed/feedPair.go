package priceFeed

// FeedPair is the struct representing a price feed for a trading pair
type FeedPair struct {
	feedA priceFeed
	feedB priceFeed
}

// MakeFeedPair is the factory method that we expose
func MakeFeedPair(dataTypeA, dataFeedAUrl, dataTypeB, dataFeedBUrl string) *FeedPair {
	return &FeedPair{
		feedA: priceFeedFactory(dataTypeA, dataFeedAUrl),
		feedB: priceFeedFactory(dataTypeB, dataFeedBUrl),
	}
}

// GetCenterPrice fetches the center price from this feed
func (p *FeedPair) GetCenterPrice() (float64, error) {
	pA, err := p.feedA.getPrice()
	if err != nil {
		return 0, err
	}

	var pB float64
	pB, err = p.feedB.getPrice()
	if err != nil {
		return 0, err
	}

	return pA / pB, nil
}
