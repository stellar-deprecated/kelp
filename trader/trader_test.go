package trader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
)

func TestIsStateSynchronized(t *testing.T) {
	balanceUnit1f := &api.Balance{
		Balance: 1.0,
		Trust:   1.0,
		Reserve: 0.0,
	}
	balanceUnit2f := &api.Balance{
		Balance: 2.0,
		Trust:   2.0,
		Reserve: 0.0,
	}
	offers1 := []hProtocol.Offer{
		hProtocol.Offer{},
	}
	offers2 := []hProtocol.Offer{
		hProtocol.Offer{},
		hProtocol.Offer{},
	}
	testCases := []struct {
		name                    string
		trades                  []model.Trade
		baseBalance1            *api.Balance
		quoteBalance1           *api.Balance
		sellingAOffers1         []hProtocol.Offer
		buyingAOffers1          []hProtocol.Offer
		baseBalance2            *api.Balance
		quoteBalance2           *api.Balance
		sellingAOffers2         []hProtocol.Offer
		buyingAOffers2          []hProtocol.Offer
		wantIsStateSynchronized bool
	}{
		{
			name:                    "nothing changed, empty offers",
			trades:                  []model.Trade{},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         []hProtocol.Offer{},
			buyingAOffers1:          []hProtocol.Offer{},
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         []hProtocol.Offer{},
			buyingAOffers2:          []hProtocol.Offer{},
			wantIsStateSynchronized: true,
		}, {
			name:                    "nothing changed, empty offers, nil trades",
			trades:                  nil,
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         []hProtocol.Offer{},
			buyingAOffers1:          []hProtocol.Offer{},
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         []hProtocol.Offer{},
			buyingAOffers2:          []hProtocol.Offer{},
			wantIsStateSynchronized: true,
		}, {
			name:                    "nothing changed, non-empty offers",
			trades:                  []model.Trade{},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         offers1,
			buyingAOffers1:          offers1,
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         offers1,
			buyingAOffers2:          offers1,
			wantIsStateSynchronized: true,
		}, {
			name:                    "only sell offers changed",
			trades:                  []model.Trade{},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         offers1,
			buyingAOffers1:          offers1,
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         offers2,
			buyingAOffers2:          offers1,
			wantIsStateSynchronized: false,
		}, {
			name:                    "only buy offers changed",
			trades:                  []model.Trade{},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         offers1,
			buyingAOffers1:          offers1,
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         offers1,
			buyingAOffers2:          offers2,
			wantIsStateSynchronized: false,
		}, {
			name:                    "only base balance changed",
			trades:                  []model.Trade{},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         offers1,
			buyingAOffers1:          offers1,
			baseBalance2:            balanceUnit2f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         offers1,
			buyingAOffers2:          offers1,
			wantIsStateSynchronized: false,
		}, {
			name:                    "only quote balance changed",
			trades:                  []model.Trade{},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         offers1,
			buyingAOffers1:          offers1,
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit1f,
			sellingAOffers2:         offers1,
			buyingAOffers2:          offers1,
			wantIsStateSynchronized: false,
		}, {
			name: "non-empty trades",
			trades: []model.Trade{
				model.Trade{},
			},
			baseBalance1:            balanceUnit1f,
			quoteBalance1:           balanceUnit2f,
			sellingAOffers1:         []hProtocol.Offer{},
			buyingAOffers1:          []hProtocol.Offer{},
			baseBalance2:            balanceUnit1f,
			quoteBalance2:           balanceUnit2f,
			sellingAOffers2:         []hProtocol.Offer{},
			buyingAOffers2:          []hProtocol.Offer{},
			wantIsStateSynchronized: false,
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			actual := isStateSynchronized(
				k.trades,
				k.baseBalance1,
				k.quoteBalance1,
				k.sellingAOffers1,
				k.buyingAOffers1,
				k.baseBalance2,
				k.quoteBalance2,
				k.sellingAOffers2,
				k.buyingAOffers2,
			)
			assert.Equal(t, k.wantIsStateSynchronized, actual)
		})
	}
}
