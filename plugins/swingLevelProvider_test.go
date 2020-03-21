package plugins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastPriceFromMap(t *testing.T) {
	price2LastPriceMap := map[float64]float64{
		0.115: 0.110,
		0.105: 0.100,
		0.095: 0.090,
		0.085: 0.080,
		0.075: 0.070,
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
			tradePrice:      0.105,
			isBuy:           true,
			wantTradePrice:  0.105,
			wantLastPrice:   0.100,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.105,
			isBuy:           false,
			wantTradePrice:  0.105,
			wantLastPrice:   0.100,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.096,
			isBuy:           true,
			wantTradePrice:  0.095,
			wantLastPrice:   0.090,
		}, {
			price2LastPrice: price2LastPriceMap,
			tradePrice:      0.096,
			isBuy:           false,
			wantTradePrice:  0.095,
			wantLastPrice:   0.090,
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
