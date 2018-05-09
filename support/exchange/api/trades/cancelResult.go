package trades

// CancelOrderResult is the result of a CancelOrder call
type CancelOrderResult int8

// These are the available types
const (
	CancelResultCancelSuccessful CancelOrderResult = 0
	CancelResultPending          CancelOrderResult = 1
	CancelResultFailed           CancelOrderResult = 2
)

// String is the stringer function
func (r CancelOrderResult) String() string {
	if r == CancelResultCancelSuccessful {
		return "cancelled"
	} else if r == CancelResultPending {
		return "pending"
	} else if r == CancelResultFailed {
		return "failed"
	}
	return "error, unrecognized CancelOrderResult"
}
