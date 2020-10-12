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
		// buy/sell (base on primary) x primary has available/partial/empty x backing has available/partial/empty x starting is zero / non-zero
		// 2 x 3 x 3 x 2 = 36 cases in total
		{
			name: "1. sell primary-available backing-available zero",
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
			name: "2. sell primary-available backing-available non-zero",
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
			name: "3. sell primary-available backing-partial zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(0.1, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(0.89286, 5),
			wantNewBackingVolume:   model.NumberFromFloat(0.1, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(0.89286, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(0.1, 5),
		}, {
			name: "4. sell primary-available backing-partial non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(5.12, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(0.1, 5),
				placedBackingUnits: model.NumberFromFloat(0.02, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(0.71429, 5),
			wantNewBackingVolume:   model.NumberFromFloat(0.08, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(5.83429, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(0.1, 5),
		}, {
			name: "5. sell primary-available backing-empty zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  false,
			wantNewPrimaryVolume:   nil,
			wantNewBackingVolume:   nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "6. sell primary-available backing-empty non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(4.82, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(5042.487, 5),
				placedBackingUnits: model.NumberFromFloat(5042.487, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  false,
			wantNewPrimaryVolume:   nil,
			wantNewBackingVolume:   nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(4.82, 6),
			wantPlacedBackingUnits: model.NumberFromFloat(5042.487, 5),
		}, {
			name: "7. sell primary-partial backing-available zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(505.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(150.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(506.0, 5),
			inputPrice:             model.NumberFromFloat(0.212, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(505.12, 5),
			wantNewBackingVolume:   model.NumberFromFloat(107.08544, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(505.12, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(107.08544, 5),
		}, {
			name: "8. sell primary-partial backing-available non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(505.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(105.12, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(150.0, 5),
				placedBackingUnits: model.NumberFromFloat(10.1, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(506.0, 5),
			inputPrice:             model.NumberFromFloat(0.212, 5),
			wantHasBackingBalance:  true,
			wantNewPrimaryVolume:   model.NumberFromFloat(400.0, 5),
			wantNewBackingVolume:   model.NumberFromFloat(84.8, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(505.12, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(94.9, 5),
		}, {
			name: "9. sell primary-partial backing-partial zero",
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
			name: "10. sell primary-partial backing-partial non-zero",
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
			assert.Equal(t, k.wantNewPrimaryVolume, newPrimaryVolume)
			assert.Equal(t, k.wantNewBackingVolume, newBackingVolume)
			assert.Equal(t, k.wantPlacedPrimaryUnits.AsString(), k.bc.getPlacedPrimaryUnits().AsString())
			assert.Equal(t, k.wantPlacedBackingUnits.AsString(), k.bc.getPlacedBackingUnits().AsString())
		})
	}
}
