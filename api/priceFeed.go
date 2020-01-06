package api

import "log"

// PriceFeed allows you to fetch the price of a feed
type PriceFeed interface {
	GetPrice() (float64, error)
}

// TODO this should be structured as a specific impl. of the PriceFeed interface
// FeedPair is the struct representing a price feed for a trading pair
type FeedPair struct {
	FeedA PriceFeed
	FeedB PriceFeed
}

// GetFeedPairPrice fetches the price by dividing FeedA by FeedB
func (p *FeedPair) GetFeedPairPrice() (float64, error) {
	pA, err := p.FeedA.GetPrice()
	if err != nil {
		return 0, err
	}

	var pB float64
	pB, err = p.FeedB.GetPrice()
	if err != nil {
		return 0, err
	}

	price := pA / pB
	log.Printf("feedPair prices: feedA=%.8f, feedB=%.8f; price=%.8f\n", pA, pB, price)
	return price, nil
}
