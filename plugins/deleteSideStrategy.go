package plugins

import (
	"log"

	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

// deleteSideStrategy is a sideStrategy to delete the orders for a given currency pair on one side of the orderbook
type deleteSideStrategy struct {
	sdex       *SDEX
	assetBase  *hProtocol.Asset
	assetQuote *hProtocol.Asset
}

// ensure it implements SideStrategy
var _ api.SideStrategy = &deleteSideStrategy{}

// makeDeleteSideStrategy is a factory method for deleteSideStrategy
func makeDeleteSideStrategy(
	sdex *SDEX,
	assetBase *hProtocol.Asset,
	assetQuote *hProtocol.Asset,
) api.SideStrategy {
	return &deleteSideStrategy{
		sdex:       sdex,
		assetBase:  assetBase,
		assetQuote: assetQuote,
	}
}

// PruneExistingOffers impl
func (s *deleteSideStrategy) PruneExistingOffers(offers []hProtocol.Offer) ([]build.TransactionMutator, []hProtocol.Offer) {
	log.Printf("deleteSideStrategy: deleting %d offers\n", len(offers))
	pruneOps := []txnbuild.Operation{}
	for i := 0; i < len(offers); i++ {
		pOp := s.sdex.DeleteOffer(offers[i])
		pruneOps = append(pruneOps, &pOp)
	}
	return api.ConvertOperation2TM(pruneOps), []hProtocol.Offer{}
}

// PreUpdate impl
func (s *deleteSideStrategy) PreUpdate(maxAssetBase float64, maxAssetQuote float64, trustBase float64, trustQuote float64) error {
	return nil
}

// UpdateWithOps impl
func (s *deleteSideStrategy) UpdateWithOps(offers []hProtocol.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error) {
	return []build.TransactionMutator{}, nil, nil
}

// PostUpdate impl
func (s *deleteSideStrategy) PostUpdate() error {
	return nil
}

// GetFillHandlers impl
func (s *deleteSideStrategy) GetFillHandlers() ([]api.FillHandler, error) {
	return nil, nil
}
