package plugins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastPriceFromMap(t *testing.T) {
	price2LastPriceMap := map[float64]float64{
		0.075: 0.070, // sell side because offer price (key) is greater than last price (value)
		0.074: 0.080, // buy side because offer price (key) is less than last price (value)
	}

	testCases := []struct {
		price2LastPrice map[float64]float64
		tradePrice      float64
		isBuy           bool
		wantTradePrice  float64
		wantLastPrice   float64
	}{
		{
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.075,
			isBuy:           false,
			wantTradePrice:  0.075,
			wantLastPrice:   0.070,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.074,
			isBuy:           false,
			wantTradePrice:  0.075,
			wantLastPrice:   0.070,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.0745,
			isBuy:           false,
			wantTradePrice:  0.075,
			wantLastPrice:   0.070,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.074,
			isBuy:           true,
			wantTradePrice:  0.074,
			wantLastPrice:   0.080,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.075,
			isBuy:           true,
			wantTradePrice:  0.074,
			wantLastPrice:   0.080,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.0745,
			isBuy:           true,
			wantTradePrice:  0.074,
			wantLastPrice:   0.080,
		},
	}

	for _, kase := range testCases {
		t.Run(fmt.Sprintf("%.4f/%v", kase.tradePrice, kase.isBuy), func(t *testing.T) {
			lastTradePrice, lastPrice := getLastPriceFromMap(kase.price2LastPrice, kase.tradePrice, kase.isBuy)
			if !assert.Equal(t, kase.wantTradePrice, lastTradePrice) {
				return
			}

			if !assert.Equal(t, kase.wantLastPrice, lastPrice) {
				return
			}
		})
	}
}
