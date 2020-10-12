package plugins

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/model"
)

func TestBalanceCoordinatorCheckBalance(t *testing.T) {
	// imagine prices such that we are trading base asset as XLM and quote asset as USD
	testCases := []struct {
		name                   string
		bc                     *balanceCoordinator
		inputVol               *model.Number
		inputPrice             *model.Number
		wantHasBackingBalance  bool
		wantNewPrimaryVolume   *model.Number
		wantNewBackingVolume   *model.Number
		wantPlacedPrimaryUnits *model.Number
		wantPlacedBackingUnits *model.Number
	}{
		// buy/sell (primary) x base/quote x available/partial/empty x starting is zero / non-zero = 24 cases in total
		{
			name: "sell base available zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(101.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(1.0, 5),
			wantNewBackingVolume:   model.NumberFromFloat(0.112, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(1.0, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(0.112, 5),
		}, {
			name: "sell base available non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(100.56, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(101.0, 5),
				placedBackingUnits: model.NumberFromFloat(8.457, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(3.14, 5),
			inputPrice:             model.NumberFromFloat(0.131, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(3.14, 5),
			wantNewBackingVolume:   model.NumberFromFloat(0.41134, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(103.7, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(8.86834, 5),
		}, {
			name: "sell base partial zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(101.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(2000.57, 5),
			inputPrice:             model.NumberFromFloat(0.21, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(480.95238, 5),
			wantNewBackingVolume:   model.NumberFromFloat(101.0, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(480.95238, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(101.0, 5),
		}, {
			name: "sell base partial non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(200.56, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(101.0, 5),
				placedBackingUnits: model.NumberFromFloat(51.3245, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(2000.57, 5),
			inputPrice:             model.NumberFromFloat(0.21, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(236.55, 5),
			wantNewBackingVolume:   model.NumberFromFloat(49.6755, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(437.11, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(101.0, 5),
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			hasBackingBalance, newPrimaryVolume, newBackingVolume := k.bc.checkBalance(k.inputVol, k.inputPrice)
			assert.Equal(t, k.wantHasBackingBalance, hasBackingBalance)
			assert.Equal(t, k.wantNewPrimaryVolume.AsString(), newPrimaryVolume.AsString())
			assert.Equal(t, k.wantNewBackingVolume.AsString(), newBackingVolume.AsString())
			assert.Equal(t, k.wantPlacedPrimaryUnits.AsString(), k.bc.getPlacedPrimaryUnits().AsString())
			assert.Equal(t, k.wantPlacedBackingUnits.AsString(), k.bc.getPlacedBackingUnits().AsString())
		})
	}
}
