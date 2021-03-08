package plugins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

func TestManageOffer2Order(t *testing.T) {
	testCases := []struct {
		op         *txnbuild.ManageSellOffer
		oc         *model.OrderConstraints
		wantAction model.OrderAction
		wantAmount float64
		wantPrice  float64
	}{
		{
			op:         makeSellOpAmtPrice(0.0018, 50500.0),
			oc:         model.MakeOrderConstraints(2, 4, 0.001),
			wantAction: model.OrderActionSell,
			wantAmount: 0.0018,
			wantPrice:  50500.0,
		}, {
			op:         makeBuyOpAmtPrice(0.0018, 50500.0),
			oc:         model.MakeOrderConstraints(2, 4, 0.001),
			wantAction: model.OrderActionBuy,
			wantAmount: 0.0018,
			// 1/50500.0 = 0.000019801980198, we need to reduce it to 7 decimals precision because of sdex op, giving 0.0000198 which when inverted is 50505.05 at price precision = 2
			wantPrice: 50505.05,
		},
	}

	for _, k := range testCases {
		baseAsset := utils.Asset2Asset2(testBaseAsset)
		quoteAsset := utils.Asset2Asset2(testQuoteAsset)
		order, e := manageOffer2Order(k.op, baseAsset, quoteAsset, k.oc)
		if !assert.NoError(t, e) {
			return
		}

		assert.Equal(t, k.wantAction, order.OrderAction, fmt.Sprintf("expected '%s' but got '%s'", k.wantAction.String(), order.OrderAction.String()))
		assert.Equal(t, k.wantPrice, order.Price.AsFloat())
		assert.Equal(t, k.wantAmount, order.Volume.AsFloat())
	}
}
