package plugins

import (
	"fmt"
	"testing"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stretchr/testify/assert"
)

func TestComputeOffersToPrune(t *testing.T) {
	testCases := []struct {
		offerPrices []float64
		levelPrices []float64
		want        []bool
	}{
		{
			offerPrices: []float64{},
			levelPrices: []float64{},
			want:        []bool{},
		}, {
			offerPrices: []float64{},
			levelPrices: []float64{1.0},
			want:        []bool{},
		}, {
			offerPrices: []float64{1.0},
			levelPrices: []float64{1.0},
			want:        []bool{false},
		}, {
			offerPrices: []float64{1.0},
			levelPrices: []float64{1.0, 1.2},
			want:        []bool{false},
		}, {
			offerPrices: []float64{0.9},
			levelPrices: []float64{1.0, 1.2},
			want:        []bool{false},
		}, {
			offerPrices: []float64{1.0, 1.2},
			levelPrices: []float64{1.0},
			want:        []bool{false, true},
		}, {
			offerPrices: []float64{0.9, 1.0},
			levelPrices: []float64{1.0},
			want:        []bool{true, false},
		}, {
			offerPrices: []float64{10.0, 11.0},
			levelPrices: []float64{1.0},
			want:        []bool{false, true},
		}, {
			offerPrices: []float64{0.9, 1.2},
			levelPrices: []float64{1.0},
			want:        []bool{true, false},
		}, {
			offerPrices: []float64{1.0, 1.0},
			levelPrices: []float64{1.0},
			want:        []bool{false, true},
		}, {
			offerPrices: []float64{1.0, 1.0, 1.2},
			levelPrices: []float64{1.0},
			want:        []bool{false, true, true},
		}, {
			offerPrices: []float64{1.0, 1.0},
			levelPrices: []float64{1.0, 1.0},
			want:        []bool{false, false},
		}, {
			offerPrices: []float64{1.0, 1.2},
			levelPrices: []float64{1.0, 1.2},
			want:        []bool{false, false},
		}, {
			offerPrices: []float64{1.0},
			levelPrices: []float64{1.0, 1.0},
			want:        []bool{false},
		}, {
			offerPrices: []float64{1.0},
			levelPrices: []float64{1.0, 1.2},
			want:        []bool{false},
		},
	}

	for i, kase := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if !assert.Equal(t, len(kase.want), len(kase.offerPrices), "invalid test case") {
				return
			}

			offers := []hProtocol.Offer{}
			for _, p := range kase.offerPrices {
				num, den, e := model.NumberFromFloat(p, 8).AsRatio()
				if !assert.NoError(t, e) {
					return
				}
				offer := hProtocol.Offer{}
				offer.PriceR.N = num
				offer.PriceR.D = den
				offers = append(offers, offer)
			}

			levels := []api.Level{}
			for _, p := range kase.levelPrices {
				levels = append(levels, api.Level{
					Price:  *model.NumberFromFloat(p, 8),
					Amount: *model.NumberFromFloat(1, 8),
				})
			}

			result := computeOffersToPrune(offers, levels)
			assert.Equal(t, kase.want, result)
		})
	}
}
