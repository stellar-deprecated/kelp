package plugins

import (
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// SubmitFilter allows you to filter out operations before submitting to the network
type SubmitFilter interface {
	Apply(
		ops []build.TransactionMutator,
		sellingOffers []horizon.Offer, // quoted quote/base
		buyingOffers []horizon.Offer, // quoted base/quote
	) ([]build.TransactionMutator, error)
}
