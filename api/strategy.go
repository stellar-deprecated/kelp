package api

import (
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/model"
)

// Strategy represents some logic for a bot to follow while doing market making
type Strategy interface {
	PruneExistingOffers(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer, []horizon.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(buyingAOffers []horizon.Offer, sellingAOffers []horizon.Offer) ([]build.TransactionMutator, error)
	PostUpdate() error
	GetFillHandlers() ([]FillHandler, error)
}

// SideStrategy represents a strategy on a single side of the orderbook
type SideStrategy interface {
	PruneExistingOffers(offers []horizon.Offer) ([]build.TransactionMutator, []horizon.Offer)
	PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error
	UpdateWithOps(offers []horizon.Offer) (ops []build.TransactionMutator, newTopOffer *model.Number, e error)
	PostUpdate() error
	GetFillHandlers() ([]FillHandler, error)
}
