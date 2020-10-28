package plugins

import (
	"testing"

	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"

	"github.com/stellar/kelp/model"
)

func TestTransformOrders(t *testing.T) {
	testCases := []struct {
		name                  string
		inputPrice            *model.Number
		inputVolume           *model.Number
		orderAction           model.OrderAction
		priceMultiplier       float64
		volumeMultiplier      float64
		maybeMaxVolumeBaseCap *float64
		wantPrice             *model.Number
		wantVolume            *model.Number
	}{
		{
			name:                  "buy below capped",
			inputPrice:            model.NumberFromFloat(0.15, 6),
			inputVolume:           model.NumberFromFloat(51.5, 5),
			orderAction:           model.OrderActionBuy,
			priceMultiplier:       0.90,
			volumeMultiplier:      0.25,
			maybeMaxVolumeBaseCap: pointy.Float64(15.0),
			wantPrice:             model.NumberFromFloat(0.135, 6),
			wantVolume:            model.NumberFromFloat(12.875, 5),
		}, {
			name:                  "sell below capped",
			inputPrice:            model.NumberFromFloat(1.15, 6),
			inputVolume:           model.NumberFromFloat(1.5123, 5),
			orderAction:           model.OrderActionSell,
			priceMultiplier:       0.90,
			volumeMultiplier:      0.25,
			maybeMaxVolumeBaseCap: pointy.Float64(15.0),
			wantPrice:             model.NumberFromFloat(1.035, 6),
			wantVolume:            model.NumberFromFloat(0.37808, 5), // round up
		}, {
			name:                  "buy above capped",
			inputPrice:            model.NumberFromFloat(0.15, 6),
			inputVolume:           model.NumberFromFloat(80.0, 5),
			orderAction:           model.OrderActionBuy,
			priceMultiplier:       0.90,
			volumeMultiplier:      0.25,
			maybeMaxVolumeBaseCap: pointy.Float64(15.0),
			wantPrice:             model.NumberFromFloat(0.135, 6),
			wantVolume:            model.NumberFromFloat(15.0, 5),
		}, {
			name:                  "sell above capped",
			inputPrice:            model.NumberFromFloat(1.15, 6),
			inputVolume:           model.NumberFromFloat(151.23, 5),
			orderAction:           model.OrderActionSell,
			priceMultiplier:       0.90,
			volumeMultiplier:      0.25,
			maybeMaxVolumeBaseCap: pointy.Float64(15.0),
			wantPrice:             model.NumberFromFloat(1.035, 6),
			wantVolume:            model.NumberFromFloat(15.0, 5),
		}, {
			name:                  "buy with 0 cap",
			inputPrice:            model.NumberFromFloat(0.15, 6),
			inputVolume:           model.NumberFromFloat(80.0, 5),
			orderAction:           model.OrderActionBuy,
			priceMultiplier:       0.90,
			volumeMultiplier:      0.25,
			maybeMaxVolumeBaseCap: nil,
			wantPrice:             model.NumberFromFloat(0.135, 6),
			wantVolume:            model.NumberFromFloat(20.0, 5),
		}, {
			name:                  "sell with 0 cap",
			inputPrice:            model.NumberFromFloat(1.15, 6),
			inputVolume:           model.NumberFromFloat(151.23, 5),
			orderAction:           model.OrderActionSell,
			priceMultiplier:       0.90,
			volumeMultiplier:      0.25,
			maybeMaxVolumeBaseCap: nil,
			wantPrice:             model.NumberFromFloat(1.035, 6),
			wantVolume:            model.NumberFromFloat(37.8075, 5),
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			order := model.Order{
				Pair:        &model.TradingPair{Base: model.XLM, Quote: model.USDT},
				OrderAction: k.orderAction,
				OrderType:   model.OrderTypeLimit,
				Price:       k.inputPrice,
				Volume:      k.inputVolume,
			}
			transformOrders([]model.Order{order}, k.priceMultiplier, k.volumeMultiplier, k.maybeMaxVolumeBaseCap)

			assert.Equal(t, &model.TradingPair{Base: model.XLM, Quote: model.USDT}, order.Pair)
			assert.Equal(t, k.orderAction, order.OrderAction)
			assert.Equal(t, model.OrderTypeLimit, order.OrderType)
			assert.Equal(t, k.wantPrice, order.Price)
			assert.Equal(t, k.wantVolume, order.Volume)
		})
	}
}

func TestFilterOrdersByVolume(t *testing.T) {
	type amtPrice struct {
		a float64
		p float64
	}

	testCases := []struct {
		name             string
		inputOrderValues []amtPrice
		minBaseVolume    float64
		wantOrderValues  []amtPrice
	}{
		{
			name: "keep first",
			inputOrderValues: []amtPrice{
				{a: 1.0, p: 1.0},
				{a: 0.5, p: 1.1},
			},
			minBaseVolume: 1.0,
			wantOrderValues: []amtPrice{
				{a: 1.0, p: 1.0},
			},
		}, {
			name: "keep second",
			inputOrderValues: []amtPrice{
				{a: 1.0, p: 1.0},
				{a: 2.5, p: 1.1},
			},
			minBaseVolume: 1.000001,
			wantOrderValues: []amtPrice{
				{a: 2.5, p: 1.1},
			},
		}, {
			name: "keep none",
			inputOrderValues: []amtPrice{
				{a: 1.0, p: 1.0},
				{a: 2.5, p: 1.1},
			},
			minBaseVolume:   2.500001,
			wantOrderValues: []amtPrice{},
		}, {
			name: "keep all",
			inputOrderValues: []amtPrice{
				{a: 1.0, p: 1.0},
				{a: 2.5, p: 1.1},
				{a: 3.5, p: 2.1},
			},
			minBaseVolume: 1.0,
			wantOrderValues: []amtPrice{
				{a: 1.0, p: 1.0},
				{a: 2.5, p: 1.1},
				{a: 3.5, p: 2.1},
			},
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			// convert input to Orders
			inputOrders := []model.Order{}
			for _, ap := range k.inputOrderValues {
				inputOrders = append(inputOrders, model.Order{
					Pair:        &model.TradingPair{Base: model.XLM, Quote: model.USDT},
					OrderAction: model.OrderActionSell,
					OrderType:   model.OrderTypeLimit,
					Price:       model.NumberFromFloat(ap.p, 5),
					Volume:      model.NumberFromFloat(ap.a, 5),
				})
			}

			outputOrders := filterOrdersByVolume(inputOrders, k.minBaseVolume)

			// convert output from Orders
			output := []amtPrice{}
			for _, o := range outputOrders {
				output = append(output, amtPrice{
					a: o.Volume.AsFloat(),
					p: o.Price.AsFloat(),
				})
			}
			assert.Equal(t, k.wantOrderValues, output)
		})
	}
}

func TestBalanceCoordinatorCheckBalance(t *testing.T) {
	// imagine prices such that we are trading base asset as XLM and quote asset as USD
	testCases := []struct {
		name                   string
		bc                     *balanceCoordinator
		inputVol               *model.Number
		inputPrice             *model.Number
		wantHasBackingBalance  bool
		wantNewBaseVolume      *model.Number
		wantNewQuoteVolume     *model.Number
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
			wantNewBaseVolume:      model.NumberFromFloat(1.0, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(0.112, 5),
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
			wantNewBaseVolume:      model.NumberFromFloat(3.13999, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(0.41133, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(103.69999, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(8.86833, 5),
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
			wantNewBaseVolume:      model.NumberFromFloat(0.89285, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(0.09999, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(0.89285, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(0.09999, 5),
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
			wantNewBaseVolume:      model.NumberFromFloat(0.71428, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(0.07999, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(5.83428, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(0.09999, 5),
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
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
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
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
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
			wantNewBaseVolume:      model.NumberFromFloat(505.12, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(107.08544, 5),
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
			wantNewBaseVolume:      model.NumberFromFloat(400.0, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(84.8, 5),
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
			wantNewBaseVolume:      model.NumberFromFloat(480.95238, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(100.99999, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(480.95238, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(100.99999, 5),
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
			wantNewBaseVolume:      model.NumberFromFloat(236.55, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(49.6755, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(437.11, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(101.0, 5),
		}, {
			name: "11. sell primary-partial backing-empty zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(2000.57, 5),
			inputPrice:             model.NumberFromFloat(0.21, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "12. sell primary-partial backing-empty non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1005.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(1.56, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(200.56, 5),
				placedBackingUnits: model.NumberFromFloat(200.56, 6),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(2000.57, 5),
			inputPrice:             model.NumberFromFloat(0.21, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(1.56, 6),
			wantPlacedBackingUnits: model.NumberFromFloat(200.56, 6),
		}, {
			name: "13. sell primary-empty backing-available zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(0.0, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(200.56, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(45.157, 5),
			inputPrice:             model.NumberFromFloat(0.45, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "14. sell primary-empty backing-available non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1.23, 6),
				placedPrimaryUnits: model.NumberFromFloat(1.23, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(200.56, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(45.157, 5),
			inputPrice:             model.NumberFromFloat(0.45, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(1.23, 6),
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "15. sell primary-empty backing-partial zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(0.0, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(200.56, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(45.157, 5),
			inputPrice:             model.NumberFromFloat(0.45, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "16. sell primary-empty backing-partial non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1.23, 6),
				placedPrimaryUnits: model.NumberFromFloat(1.23, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(200.56, 5),
				placedBackingUnits: model.NumberFromFloat(14.03, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(4.157, 5),
			inputPrice:             model.NumberFromFloat(0.125, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(1.23, 6),
			wantPlacedBackingUnits: model.NumberFromFloat(14.03, 5),
		}, {
			name: "17. sell primary-empty backing-empty zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberConstants.Zero,
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(4.157, 5),
			inputPrice:             model.NumberFromFloat(0.125, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "18. sell primary-empty backing-empty non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1.23, 6),
				placedPrimaryUnits: model.NumberFromFloat(1.23, 6),
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(200.56, 5),
				placedBackingUnits: model.NumberFromFloat(200.56, 5),
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(4.157, 5),
			inputPrice:             model.NumberFromFloat(0.125, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(1.23, 6),
			wantPlacedBackingUnits: model.NumberFromFloat(200.56, 5),
		}, {
			name: "19. buy primary-available backing-available zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(1001.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(1.0, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(0.112, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(0.112, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(1.0, 5),
		}, {
			name: "20. buy primary-available backing-available non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(5.4, 6),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(1001.0, 5),
				placedBackingUnits: model.NumberFromFloat(3.12, 6),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.112, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(1.0, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(0.112, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(5.512, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(4.12, 5),
		}, {
			name: "21. buy primary-available backing-partial zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(51.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(51.00001, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(51.0, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(12.75, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(12.75, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(51.0, 5),
		}, {
			name: "22. buy primary-available backing-partial non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(57.4, 6),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(51.0, 5),
				placedBackingUnits: model.NumberFromFloat(5.45, 5),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(46.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(45.55, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(11.3875, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(68.7875, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(51.0, 5),
		}, {
			name: "23. buy primary-available backing-empty zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(51.00001, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "24. buy primary-available backing-empty non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(15.32, 6),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(51.00001, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(15.32, 6),
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "25. buy primary-partial backing-available zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(500.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(421.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(420.48, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(105.12, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(105.12, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(420.48, 5),
		}, {
			name: "26. buy primary-partial backing-available non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(0.1, 6),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(500.0, 5),
				placedBackingUnits: model.NumberFromFloat(0.1, 5),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(421.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(420.08, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(105.02, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(105.12, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(420.18, 5),
		}, {
			name: "27. buy primary-partial backing-partial zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(500.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(501.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(420.48, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(105.12, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(105.12, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(420.48, 5),
		}, {
			name: "28. buy primary-partial backing-partial non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(0.1, 6),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(500.0, 5),
				placedBackingUnits: model.NumberFromFloat(0.1, 5),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(501.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(420.08, 5),
			wantNewQuoteVolume:     model.NumberFromFloat(105.02, 5),
			wantPlacedPrimaryUnits: model.NumberFromFloat(105.12, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(420.18, 5),
		}, {
			name: "29. buy primary-partial backing-empty zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(501.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "30. buy primary-partial backing-empty non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(105.12, 6),
				placedPrimaryUnits: model.NumberFromFloat(0.1, 6),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(0.1, 5),
				placedBackingUnits: model.NumberFromFloat(0.1, 5),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(601.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(0.1, 6),
			wantPlacedBackingUnits: model.NumberFromFloat(0.1, 5),
		}, {
			name: "31. buy primary-empty backing-available zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberConstants.Zero,
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(10.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "32. buy primary-empty backing-available non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(10.0, 5),
				placedPrimaryUnits: model.NumberFromFloat(10.0, 5),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(10.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(1.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(10.0, 5),
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "33. buy primary-empty backing-partial zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(0.0, 5),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(0.1, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(15.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "34. buy primary-empty backing-partial non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(4.2, 5),
				placedPrimaryUnits: model.NumberFromFloat(4.2, 5),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(1000.1, 5),
				placedBackingUnits: model.NumberFromFloat(0.1, 5),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(15.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(4.2, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(0.1, 5),
		}, {
			name: "35. buy primary-empty backing-empty zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(0.0, 5),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(0.0, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(15.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberConstants.Zero,
			wantPlacedBackingUnits: model.NumberConstants.Zero,
		}, {
			name: "36. buy primary-empty backing-empty non-zero",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(100.0, 5),
				placedPrimaryUnits: model.NumberFromFloat(100.0, 5),
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(100.45, 5),
				placedBackingUnits: model.NumberFromFloat(100.45, 5),
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(15.0, 5),
			inputPrice:             model.NumberFromFloat(0.25, 5),
			wantHasBackingBalance:  false,
			wantNewBaseVolume:      nil,
			wantNewQuoteVolume:     nil,
			wantPlacedPrimaryUnits: model.NumberFromFloat(100.0, 5),
			wantPlacedBackingUnits: model.NumberFromFloat(100.45, 5),
		},
		// these are the tests spawned from https://github.com/stellar/kelp/issues/541
		{
			name: "rounding - buy base on primary - truncate rounding",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(646.1, 7),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "quote",
				isPrimaryBuy:       true,
				backingBalance:     model.NumberFromFloat(1.9, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "base",
			},
			inputVol:               model.NumberFromFloat(2.0, 2),   // intentionally use a less precise volume
			inputPrice:             model.NumberFromFloat(368.0, 8), // intentionally use a precision of price more than volume
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(1.75, 2),  // don't modify precision volume nunmbers
			wantNewQuoteVolume:     model.NumberFromFloat(644.0, 2), // don't modify precision volume nunmbers
			wantPlacedPrimaryUnits: model.NumberFromFloat(644.0, 2),
			wantPlacedBackingUnits: model.NumberFromFloat(1.75, 2),
		}, {
			name: "rounding - sell base on primary - truncate rounding",
			bc: &balanceCoordinator{
				primaryBalance:     model.NumberFromFloat(1.9, 7),
				placedPrimaryUnits: model.NumberConstants.Zero,
				primaryAssetType:   "base",
				isPrimaryBuy:       false,
				backingBalance:     model.NumberFromFloat(646.1, 5),
				placedBackingUnits: model.NumberConstants.Zero,
				backingAssetType:   "quote",
			},
			inputVol:               model.NumberFromFloat(2.0, 2),   // intentionally use a less precise volume
			inputPrice:             model.NumberFromFloat(368.0, 8), // intentionally use a precision of price more than volume
			wantHasBackingBalance:  true,
			wantNewBaseVolume:      model.NumberFromFloat(1.75, 2),  // don't modify precision volume nunmbers
			wantNewQuoteVolume:     model.NumberFromFloat(644.0, 2), // don't modify precision volume nunmbers
			wantPlacedPrimaryUnits: model.NumberFromFloat(1.75, 2),
			wantPlacedBackingUnits: model.NumberFromFloat(644.0, 2),
		},
	}

	for _, k := range testCases {
		t.Run(k.name, func(t *testing.T) {
			hasBackingBalance, newBaseVolume, newQuoteVolume := k.bc.checkBalance(k.inputVol, k.inputPrice)
			assert.Equal(t, k.wantHasBackingBalance, hasBackingBalance)
			assert.Equal(t, k.wantNewBaseVolume, newBaseVolume)
			assert.Equal(t, k.wantNewQuoteVolume, newQuoteVolume)
			assert.Equal(t, k.wantPlacedPrimaryUnits.AsString(), k.bc.getPlacedPrimaryUnits().AsString())
			assert.Equal(t, k.wantPlacedBackingUnits.AsString(), k.bc.getPlacedBackingUnits().AsString())
		})
	}
}
