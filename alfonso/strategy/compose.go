package strategy

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/number"

	"github.com/lightyeario/kelp/alfonso/strategy/sideStrategy"
	kelp "github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/errors"
)

// ComposeStrategy is a strategy that is composed of two sub-strategies
type ComposeStrategy struct {
	assetBase  *horizon.Asset
	assetQuote *horizon.Asset
	buyStrat   sideStrategy.SideStrategy
	sellStrat  sideStrategy.SideStrategy
}

// ensure it implements Strategy
var _ Strategy = &ComposeStrategy{}

// MakeComposeStrategy is a factory method for ComposeStrategy
func MakeComposeStrategy(
	assetBase *horizon.Asset,
	assetQuote *horizon.Asset,
	buyStrat sideStrategy.SideStrategy,
	sellStrat sideStrategy.SideStrategy,
) Strategy {
	return &ComposeStrategy{
		assetBase:  assetBase,
		assetQuote: assetQuote,
		buyStrat:   buyStrat,
		sellStrat:  sellStrat,
	}
}

// PruneExistingOffers impl
func (s *ComposeStrategy) PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer, []horizon.Offer) {
	pruneOps1, newBuyingAOffers := s.buyStrat.PruneExistingOffers(buyingAOffers)
	pruneOps2, newSellingAOffers := s.sellStrat.PruneExistingOffers(sellingAOffers)
	pruneOps1 = append(pruneOps1, pruneOps2...)
	return pruneOps1, newBuyingAOffers, newSellingAOffers
}

// PreUpdate impl
func (s *ComposeStrategy) PreUpdate(maxAssetBase float64, maxAssetQuote float64, trustBase float64, trustQuote float64) error {
	// swap assets (base/quote) for buying strategy
	e1 := s.buyStrat.PreUpdate(maxAssetQuote, maxAssetBase, trustQuote, trustBase)
	// assets maintain same ordering for selling
	e2 := s.sellStrat.PreUpdate(maxAssetBase, maxAssetQuote, trustBase, trustQuote)

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
	// buy side, flip newTopBuyPrice because it will be inverted from this parent strategy's context of base/quote
	buyOps, newTopBuyPriceInverted, e1 := s.buyStrat.UpdateWithOps(buyingAOffers)
	newTopBuyPrice := number.Invert(newTopBuyPriceInverted)
	// sell side
	sellOps, _, e2 := s.sellStrat.UpdateWithOps(sellingAOffers)

	// check for errors
	ops := []build.TransactionMutator{}
	if e1 != nil && e2 != nil {
		return ops, fmt.Errorf("errors on both sides: buying (= %s) and selling (= %s)", e1, e2)
	} else if e1 != nil {
		return ops, errors.Wrap(e1, "error in buying sub-strategy")
	} else if e2 != nil {
		return ops, errors.Wrap(e2, "error in selling sub-strategy")
	}

	// combine ops correctly based on possible crossing offers
	if newTopBuyPrice != nil && len(sellingAOffers) > 0 && newTopBuyPrice.AsFloat() >= kelp.PriceAsFloat(sellingAOffers[0].Price) {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	} else {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	}
	return ops, nil
}

// PostUpdate impl
func (s *ComposeStrategy) PostUpdate() error {
	return nil
}
