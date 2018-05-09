package orderbook

import (
	"fmt"

	"github.com/lightyeario/kelp/support/exchange/api/dates"
	"github.com/lightyeario/kelp/support/exchange/api/number"
)

// OpenOrder represents an open order for a trading account
type OpenOrder struct {
	Order
	ID             string
	StartTime      *dates.Timestamp
	ExpireTime     *dates.Timestamp
	VolumeExecuted *number.Number
}

// String is the stringer function
func (o OpenOrder) String() string {
	return fmt.Sprintf("OpenOrder[order=%s, ID=%s, startTime=%d, expireTime=%d, volumeExecuted=%s]",
		o.Order.String(),
		o.ID,
		o.StartTime.AsInt64(),
		o.ExpireTime.AsInt64(),
		o.VolumeExecuted.AsString(),
	)
}
