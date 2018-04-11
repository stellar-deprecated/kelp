package sideStrategy

import (
	"github.com/lightyeario/kelp/support/exchange/number"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// SideStrategy represents a strategy on a single side of the orderbook
type SideStrategy interface {
	PruneExistingOffers(offers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64) error
	UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *number.Number, e error)
	PostUpdate() error
}
