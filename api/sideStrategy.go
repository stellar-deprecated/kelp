package api

import (
	"github.com/lightyeario/kelp/model"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// SideStrategy represents a strategy on a single side of the orderbook
type SideStrategy interface {
	PruneExistingOffers(offers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error)
	PostUpdate() error
}
