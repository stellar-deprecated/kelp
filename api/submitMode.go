package api

import (
	"fmt"
)

// SubmitMode is the type of mode to be used when submitting orders to the trader bot
type SubmitMode uint8

// constants for the SubmitMode
const (
	SubmitModeMakerOnly SubmitMode = iota
	SubmitModeBoth
)

// ParseSubmitMode converts a string to the SubmitMode constant
func ParseSubmitMode(submitMode string) (SubmitMode, error) {
	if submitMode == "maker_only" {
		return SubmitModeMakerOnly, nil
	} else if submitMode == "both" || submitMode == "" {
		return SubmitModeBoth, nil
	}

	return SubmitModeBoth, fmt.Errorf("unable to parse submit mode: %s", submitMode)
}

func (s *SubmitMode) String() string {
	if *s == SubmitModeMakerOnly {
		return "maker_only"
	}

	return "both"
}
