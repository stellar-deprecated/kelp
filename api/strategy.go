package api

import (
	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/model"
)

// Strategy represents some logic for a bot to follow while doing market making
type Strategy interface {
	PruneExistingOffers(buyingAOffers []hProtocol.Offer, sellingAOffers []hProtocol.Offer) ([]build.TransactionMutator, []hProtocol.Offer, []hProtocol.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(buyingAOffers []hProtocol.Offer, sellingAOffers []hProtocol.Offer) ([]build.TransactionMutator, error)
	PostUpdate() error
	GetFillHandlers() ([]FillHandler, error)
}

// SideStrategy represents a strategy on a single side of the orderbook
type SideStrategy interface {
	PruneExistingOffers(offers []hProtocol.Offer) ([]build.TransactionMutator, []hProtocol.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(offers []hProtocol.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error)
	PostUpdate() error
	GetFillHandlers() ([]FillHandler, error)
}
