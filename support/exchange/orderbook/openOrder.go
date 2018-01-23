package orderbook

import (
	"github.com/lightyeario/kelp/support/exchange/dates"
	"github.com/lightyeario/kelp/support/exchange/number"
)

// OpenOrder represents an open order for a trading account
type OpenOrder struct {
	Order
	ID             string
	StartTime      *dates.Timestamp
	ExpireTime     *dates.Timestamp
	VolumeExecuted *number.Number
}
