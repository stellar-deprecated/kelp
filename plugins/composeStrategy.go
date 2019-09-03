package plugins

import (
	"fmt"

	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"

	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/kelp/support/utils"
)

// composeStrategy is a strategy that is composed of two sub-strategies
type composeStrategy struct {
	assetBase  *hProtocol.Asset
	assetQuote *hProtocol.Asset
	buyStrat   api.SideStrategy
	sellStrat  api.SideStrategy
}

// ensure it implements Strategy
var _ api.Strategy = &composeStrategy{}

// makeComposeStrategy is a factory method for composeStrategy
func makeComposeStrategy(
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
	buyStrat api.SideStrategy,
	sellStrat api.SideStrategy,
) api.Strategy {
	return &composeStrategy{
		assetBase:  assetBase,
		assetQuote: assetQuote,
		buyStrat:   buyStrat,
		sellStrat:  sellStrat,
	}
}

// PruneExistingOffers impl
func (s *composeStrategy) PruneExistingOffers(buyingAOffers []hProtocol.Offer, sellingAOffers []hProtocol.Offer) ([]build.TransactionMutator, []hProtocol.Offer, []hProtocol.Offer) {
	pruneOps1, newBuyingAOffers := s.buyStrat.PruneExistingOffers(buyingAOffers)
	pruneOps2, newSellingAOffers := s.sellStrat.PruneExistingOffers(sellingAOffers)
	pruneOps1 = append(pruneOps1, pruneOps2...)
	return pruneOps1, newBuyingAOffers, newSellingAOffers
}

// PreUpdate impl
func (s *composeStrategy) PreUpdate(maxAssetBase float64, maxAssetQuote float64, trustBase float64, trustQuote float64) error {
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
func (s *composeStrategy) UpdateWithOps(
	buyingAOffers []hProtocol.Offer,
	sellingAOffers []hProtocol.Offer,
) ([]build.TransactionMutator, error) {
	// buy side, flip newTopBuyPrice because it will be inverted from this parent strategy's context of base/quote
	buyOps, newTopBuyPriceInverted, e1 := s.buyStrat.UpdateWithOps(buyingAOffers)
	newTopBuyPrice := model.InvertNumber(newTopBuyPriceInverted)
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
	if newTopBuyPrice != nil && len(sellingAOffers) > 0 && newTopBuyPrice.AsFloat() >= utils.PriceAsFloat(sellingAOffers[0].Price) {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	} else {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	}
	return ops, nil
}

// PostUpdate impl
func (s *composeStrategy) PostUpdate() error {
	return nil
}

// GetFillHandlers impl
func (s *composeStrategy) GetFillHandlers() ([]api.FillHandler, error) {
	buyFillHandlers, e := s.buyStrat.GetFillHandlers()
	if e != nil {
		return nil, fmt.Errorf("error while getting fill handlers for buy side")
	}
	sellFillHandlers, e := s.sellStrat.GetFillHandlers()
	if e != nil {
		return nil, fmt.Errorf("error while getting fill handlers for sell side")
	}

	handlers := []api.FillHandler{}
	if buyFillHandlers != nil {
		handlers = append(handlers, buyFillHandlers...)
	}
	if sellFillHandlers != nil {
		handlers = append(handlers, sellFillHandlers...)
	}
	return handlers, nil
}
