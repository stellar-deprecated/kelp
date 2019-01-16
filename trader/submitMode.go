package trader

// SubmitMode is the type of mode to be used when submitting orders to the trader bot
type SubmitMode uint8

// constants for the SubmitMode
const (
	SubmitModeMakerOnly SubmitMode = iota
	SubmitModeTakerOnly
	SubmitModeBoth
)

// ParseSubmitMode converts a string to the SubmitMode constant
func ParseSubmitMode(submitMode string) SubmitMode {
	if submitMode == "maker_only" {
		return SubmitModeMakerOnly
	}

	if submitMode == "taker_only" {
		return SubmitModeTakerOnly
	}

	return SubmitModeBoth
}
