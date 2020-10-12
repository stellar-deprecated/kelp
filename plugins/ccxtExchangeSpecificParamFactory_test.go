package plugins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinanceTransformLimit(t *testing.T) {
	testCases := []struct {
		limit        int
		wantNewLimit int
		wantError    error
	}{
		{
			limit:        1,
			wantNewLimit: 5,
			wantError:    nil,
		}, {
			limit:        2,
			wantNewLimit: 5,
			wantError:    nil,
		}, {
			limit:        3,
			wantNewLimit: 5,
			wantError:    nil,
		}, {
			limit:        4,
			wantNewLimit: 5,
			wantError:    nil,
		}, {
			limit:        5,
			wantNewLimit: 5,
			wantError:    nil,
		}, {
			limit:        6,
			wantNewLimit: 10,
			wantError:    nil,
		}, {
			limit:        10,
			wantNewLimit: 10,
			wantError:    nil,
		}, {
			limit:        11,
			wantNewLimit: 20,
			wantError:    nil,
		}, {
			limit:        21,
			wantNewLimit: 50,
			wantError:    nil,
		}, {
			limit:        51,
			wantNewLimit: 100,
			wantError:    nil,
		}, {
			limit:        101,
			wantNewLimit: 500,
			wantError:    nil,
		}, {
			limit:        501,
			wantNewLimit: 1000,
			wantError:    nil,
		}, {
			limit:        1001,
			wantNewLimit: 5000,
			wantError:    nil,
		}, {
			limit:        5001,
			wantNewLimit: -1,
			wantError:    fmt.Errorf("limit requested (5001) is higher than the maximum limit allowed (5000)"),
		},
	}

	binanceParamFactory := makeCcxtExchangeSpecificParamFactoryBinance()
	for _, k := range testCases {
		t.Run(fmt.Sprintf("%d", k.limit), func(t *testing.T) {
			newLimit, e := binanceParamFactory.transformLimit(k.limit)
			assert.Equal(t, k.wantNewLimit, newLimit)
			assert.Equal(t, k.wantError, e)

			// repeat values to test caching
			newLimit, e = binanceParamFactory.transformLimit(k.limit)
			assert.Equal(t, k.wantNewLimit, newLimit)
			assert.Equal(t, k.wantError, e)
		})
	}
}
