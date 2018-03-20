package strategy

import (
	"fmt"

	kelp "github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/errors"
)

// ComposeStrategy is a strategy that is composed of two sub-strategies
type ComposeStrategy struct {
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
	buyStrat   Strategy
	sellStrat  Strategy
}

// ensure it implements Strategy
var _ Strategy = &ComposeStrategy{}

// MakeComposeStrategy is a factory method for ComposeStrategy
func MakeComposeStrategy(
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	buyStrat Strategy,
	sellStrat Strategy,
) Strategy {
	return &ComposeStrategy{
		assetBase:  assetBase,
		assetQuote: assetQuote,
		buyStrat:   buyStrat,
		sellStrat:  sellStrat,
	}
}

// PruneExistingOffers impl
func (s *ComposeStrategy) PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]horizon.Offer, []horizon.Offer) {
	newBuyingAOffers, _ := s.buyStrat.PruneExistingOffers(buyingAOffers, nil)
	_, newSellingAOffers := s.sellStrat.PruneExistingOffers(nil, sellingAOffers)
	return newBuyingAOffers, newSellingAOffers
}

// PreUpdate impl
func (s *ComposeStrategy) PreUpdate(
	maxAssetBase float64,
	maxAssetQuote float64,
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) error {
	// swap assets (base/quote) for buying strategy
	e1 := s.buyStrat.PreUpdate(maxAssetQuote, maxAssetBase, buyingAOffers, nil)
	// assets maintain same ordering for selling
	e2 := s.sellStrat.PreUpdate(maxAssetBase, maxAssetQuote, nil, sellingAOffers)

	if e1 == nil && e2 == nil {
		return nil
	}

	if e1 != nil && e2 != nil {
		return fmt.Errorf("errors on both sides: buying (= %s) and selling (= %s)", e1, e2)
	}

	if e1 != nil {
		return errors.Wrap(e1, "error in buying sub-strategy")
	}
	return errors.Wrap(e2, "error in selling sub-strategy")
}

// UpdateWithOps impl
func (s *ComposeStrategy) UpdateWithOps(
	buyingAOffers []horizon.Offer,
	sellingAOffers []horizon.Offer,
) ([]build.TransactionMutator, error) {
	buyOps, e1 := s.buyStrat.UpdateWithOps(buyingAOffers, nil)
	sellOps, e2 := s.sellStrat.UpdateWithOps(nil, sellingAOffers)

	if len(ob.Bids()) > 0 && len(sellingAOffers) > 0 && ob.Bids()[0].Price.AsFloat() >= kelp.PriceAsFloat(sellingAOffers[0].Price) {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	} else {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	}

	return []build.TransactionMutator{}, nil
}

// PostUpdate impl
func (s *ComposeStrategy) PostUpdate() error {
	return nil
}
